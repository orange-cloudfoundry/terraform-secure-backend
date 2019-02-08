package server

import (
	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-community/gautocloud"
	"github.com/cloudfoundry-community/gautocloud/connectors/generic"
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	CREDHUB_PREFIX = "/terraform-secure-backend/tfstate"
	LOCK_SUFFIX    = "-lock"
)

func init() {
	gautocloud.RegisterConnector(generic.NewConfigGenericConnector(ServerConfig{}))
}

type ServerConfig struct {
	Host               string   `json:"host" yaml:"host"`
	Name               string   `json:"name" yaml:"name"`
	Port               int      `json:"port" yaml:"port"`
	Cert               string   `json:"cert" yaml:"cert" cloud-default:"server.crt"`
	Key                string   `json:"key" yaml:"key" cloud-default:"server.key"`
	LogLevel           string   `json:"log_level" yaml:"log_level" cloud-default:"info"`
	LogJson            bool     `json:"log_json" yaml:"log_json"`
	NoColor            bool     `json:"no_color" yaml:"no_color"`
	LetsEncryptDomains []string `json:"lets_encrypt_domains" yaml:"lets_encrypt_domains"`
	Username           string   `json:"username" yaml:"username"`
	Password           string   `json:"password" yaml:"password"`
	CredhubServer      string   `json:"credhub_server" yaml:"credhub_server"`
	CredhubUsername    string   `json:"credhub_username" yaml:"credhub_username"`
	CredhubPassword    string   `json:"credhub_password" yaml:"credhub_password"`
	CredhubClient      string   `json:"credhub_client" yaml:"credhub_client"`
	CredhubSecret      string   `json:"credhub_secret" yaml:"credhub_secret"`
	CredhubCaCert      string   `json:"credhub_ca_cert" yaml:"credhub_ca_cert"`
	SkipSslValidation  bool     `json:"skip_ssl_validation" yaml:"skip_ssl_validation"`
	ShowError          bool     `json:"show_error" yaml:"show_error"`
	CEF                bool     `json:"cef" yaml:"cef"`
	CEFFile            string   `json:"cef-file" yaml:"cef-file"`
	DryRun             bool     `json:"dry-run" yaml:"dry-run"`
}

type Server struct {
	config  *ServerConfig
	handler http.Handler
	version string
}

func NewServer(version string, config *ServerConfig) (*Server, error) {
	server := &Server{config: config, version: version}
	err := server.Load()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func NewCloudServer(version string) (*Server, error) {
	config := &ServerConfig{}
	err := gautocloud.Inject(config)
	if err != nil {
		return nil, err
	}
	log.Info("Loading config from cloud environment")
	return NewServer(version, config)
}

func (s Server) loadLogConfig() {
	if s.config.LogJson {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: s.config.NoColor,
		})
	}
	if s.config.LogLevel == "" {
		return
	}
	switch strings.ToUpper(s.config.LogLevel) {
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
		return
	case "WARN":
		log.SetLevel(log.WarnLevel)
		return
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		return
	case "PANIC":
		log.SetLevel(log.PanicLevel)
		return
	case "FATAL":
		log.SetLevel(log.FatalLevel)
		return
	}
	return
}

func (s *Server) Load() error {
	s.loadLogConfig()
	if s.config.Port == 0 {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		s.config.Port = port
	}
	if gautocloud.IsInACloudEnv() && gautocloud.CurrentCloudEnv().Name() != "localcloud" {
		if _, ok := gautocloud.GetAppInfo().Properties["port"]; ok {
			s.config.Port = gautocloud.GetAppInfo().Properties["port"].(int)
		}
	}
	var err error
	if s.config.Name == "" {
		return fmt.Errorf("You must define a name to your backend to not conflict with other backend in credhub.")
	}
	s.config.CredhubCaCert, err = s.getTlsPem(s.config.CredhubCaCert)
	if err != nil {
		return err
	}

	err = s.loadHandler()
	if err != nil {
		return err
	}

	s.config.Cert, err = s.getTlsFilePath(s.config.Cert)
	if err != nil {
		return err
	}
	s.config.Key, err = s.getTlsFilePath(s.config.Key)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) loadHandler() error {
	credhubClient, err := s.CreateCredhubCli()
	if err != nil {
		return err
	}
	store := NewLockStore(credhubClient)
	controller := NewApiController(s.config.Name, credhubClient, store)
	rtr := mux.NewRouter()
	if s.config.CEF {
		var cefW io.Writer = os.Stdout
		if s.config.CEFFile != "" {
			cefW, err = os.OpenFile(s.config.CEFFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
			if err != nil {
				return fmt.Errorf("Error when opening cef file log: %s", err.Error())
			}
		}
		cefMiddleware := NewCEFMiddleware(cefW, s.version)
		rtr.Use(cefMiddleware.Middleware)
	}
	apiRtr := rtr.PathPrefix("/states").Subrouter()
	apiRtr.HandleFunc("/{name}", controller.Store).Methods("POST")
	apiRtr.HandleFunc("/{name}", controller.Retrieve).Methods("GET")
	apiRtr.HandleFunc("/{name}", controller.Delete).Methods("DELETE")
	apiRtr.HandleFunc("/{name}", controller.Lock).Methods("LOCK")
	apiRtr.HandleFunc("/{name}", controller.UnLock).Methods("UNLOCK")
	rtr.HandleFunc("/states", controller.List).Methods("GET")
	if s.config.Username != "" {
		rtr.Use(httpauth.SimpleBasicAuth(s.config.Username, s.config.Password))
	}
	s.handler = rtr
	return nil
}

func (s Server) runTls(servAddr string, handler http.Handler) (bool, error) {
	if s.config.Cert == "" || s.config.Key == "" {
		return false, fmt.Errorf("No certificate or key provided")
	}
	err := http.ListenAndServeTLS(servAddr, s.config.Cert, s.config.Key, handler)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Server) Run() error {
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer s.panicRecover(w)
		s.handler.ServeHTTP(w, req)
	})
	servAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	if s.config.LetsEncryptDomains != nil && len(s.config.LetsEncryptDomains) > 0 {
		log.Info("Serving in https on ':443' with let's encrypt certificate (443 is mandatory by let's encrypt).")
		return http.Serve(autocert.NewListener(s.config.LetsEncryptDomains...), finalHandler)
	}
	log.Infof("Serving in https on address '%s'", servAddr)
	inTls, err := s.runTls(servAddr, finalHandler)
	if err != nil {
		log.Warn("Server wasn't start with tls, maybe you didn't set a cert and key file.")
		log.Warn("For security reasons you should use tls.")
		log.Warn("You can use tls easily by setting lets encrypt through config with key 'lets_encrypt_domains'")
		log.Warnf("Errors given: '%s'", err.Error())
	}
	if inTls {
		return nil
	}
	log.Infof("Serving an insecure server in http on address '%s'", servAddr)
	return http.ListenAndServe(servAddr, finalHandler)
}

func (s Server) getTlsPem(tlsConf string) (string, error) {
	if tlsConf == "" {
		return "", nil
	}
	_, err := os.Stat(tlsConf)
	if err != nil {
		return tlsConf, nil
	}
	b, err := ioutil.ReadFile(tlsConf)
	return string(b), err
}

func (s Server) getTlsFilePath(tlsConf string) (string, error) {
	if tlsConf == "" {
		return "", nil
	}
	_, err := os.Stat(tlsConf)
	if err == nil {
		return tlsConf, nil
	}
	f, err := ioutil.TempFile("", "terraform-secure-backend")
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.WriteString(tlsConf)
	return f.Name(), nil
}

func (s Server) CreateCredhubCli() (CredhubClient, error) {
	if s.config.DryRun {
		return &NullCredhubClient{}, nil
	}
	apiEndpoint := strings.TrimPrefix(s.config.CredhubServer, "http://")
	if !strings.HasPrefix(apiEndpoint, "https://") {
		apiEndpoint = "https://" + apiEndpoint
	}
	username := s.config.CredhubUsername
	password := s.config.CredhubPassword
	clientId := s.config.CredhubClient
	clientSecret := s.config.CredhubSecret
	if (username == "" || password == "") && (clientId == "" || clientSecret == "") {
		return nil, fmt.Errorf("One of pair Username/Password or Client_id/client_secret must be set.")
	}
	options := make([]credhub.Option, 0)
	if username != "" && password != "" {
		if clientId == "" {
			clientId = "credhub_cli"
		}
		options = append(options, credhub.Auth(auth.UaaPassword(clientId, clientSecret, username, password)))
	} else {
		options = append(options, credhub.Auth(auth.UaaClientCredentials(clientId, clientSecret)))
	}
	if s.config.SkipSslValidation {
		options = append(options, credhub.SkipTLSValidation(true))
	}
	caCert := s.config.CredhubCaCert
	if caCert != "" {
		options = append(options, credhub.CaCerts(caCert))
	}
	return credhub.New(apiEndpoint, options...)
}

func (s Server) panicRecover(w http.ResponseWriter) {
	err := recover()
	if err == nil {
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	if s.config.ShowError {
		w.Header().Set("Content-Type", "application/json")
		errMsg := struct {
			Status  int    `json:"status"`
			Title   string `json:"title"`
			Details string `json:"details"`
		}{http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), fmt.Sprint(err)}
		b, _ := json.MarshalIndent(errMsg, "", "\t")
		w.Write([]byte(b))
	}
}
