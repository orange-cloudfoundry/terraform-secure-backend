// Package decoder provide a way to decode credentials from a service to a structure
// It provide a cloud tag to help user match the correct credentials
//
// This is what you can pass as a structure:
//  // Name is key of a service credentials, decoder will look at any matching credentials which have the key name and will pass the value of this credentials
//  	Name    string `cloud:"name"`           // note: by default if you don't provide a cloud tag the key will be the field name in snake_case
//  	Uri     decoder.ServiceUri              // ServiceUri is a special type. Decoder will expect an uri as a value and will give a ServiceUri
//  	User    string `cloud:".*user.*,regex"` // by passing `regex` in cloud tag it will say to decoder that the expected key must be match the regex
//  	Password string `cloud:".*user.*,regex" cloud-default:"apassword"` // by passing a tag named `cloud-default` decoder will understand that if the key is not found it must fill the field with this value
//      Aslice   []string `cloud:"aslice" cloud-default:"value1,value2"` // you can also pass a slice
//  }
package decoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azer/snakecase"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	identifier               = "cloud"
	identifier_default_value = "cloud-default"
	regexTag                 = "regex"
	defaultTag               = "default"
	skipTag                  = "-"
)

type Tag struct {
	Name         string
	Skip         bool
	IsRegex      bool
	DefaultValue string
}
type ServiceUri struct {
	Username string
	Password string
	Scheme   string
	Host     string
	Name     string
	Query    []QueryUri
	RawQuery string
	Port     int
}
type QueryUri struct {
	Key   string
	Value string
}

// Decode a map of credentials into a reflected Value
func UnmarshalToValue(serviceCredentials map[string]interface{}, ps reflect.Value, noDefaultVal bool) error {
	v := ps
	if ps.Kind() == reflect.Ptr {
		v = ps.Elem()
	}
	t := v.Type()
	var err error
	for index := 0; index < v.NumField(); index++ {
		vField := v.Field(index)
		tField := t.Field(index)
		if !vField.CanAddr() || !vField.CanSet() {
			continue
		}
		tag := parseInTag(tField.Tag.Get(identifier), tField.Name)
		if tag.Skip {
			continue
		}
		key := tag.Name
		if tag.IsRegex {
			key = getKeyFromRegex(serviceCredentials, tag.Name)
		}
		defaultValueFromTag := tField.Tag.Get(identifier_default_value)
		if defaultValueFromTag != "" {
			tag.DefaultValue = defaultValueFromTag
		}
		defaultValue := tag.DefaultValue
		if noDefaultVal {
			defaultValue = ""
		}
		data := retrieveFinalData(vField.Type(), serviceCredentials, key, defaultValue)
		if data == nil {
			continue
		}
		dataKind := reflect.TypeOf(data).Kind()
		if dataKind == reflect.String && reflect.TypeOf(data) != reflect.TypeOf(json.Number("")) {
			data, err = convertStringValue(data.(string), vField)
			if err != nil {
				return NewErrDecode(fmt.Sprintf(
					"Error on field '%s' when trying to convert value '%s' in '%s': %s",
					tField.Name,
					defaultValue,
					vField.Kind().String(),
					err.Error(),
				))
			}
		}
		err = affect(data, vField, noDefaultVal)
		if err != nil {
			return NewErrDecode(fmt.Sprintf("Error on field '%s': %s", tField.Name, err.Error()))
		}
	}
	return nil
}

// Decode a map of credentials into a structure
func Unmarshal(serviceCredentials map[string]interface{}, obj interface{}) error {
	ps := reflect.ValueOf(obj)
	return UnmarshalToValue(serviceCredentials, ps, false)
}

// Decode a map of credentials into a structure without default values
func UnmarshalNoDefault(serviceCredentials map[string]interface{}, obj interface{}) error {
	ps := reflect.ValueOf(obj)
	return UnmarshalToValue(serviceCredentials, ps, true)
}
func parseForInt(data interface{}, vField reflect.Value) interface{} {
	if reflect.ValueOf(data).Kind() != reflect.Float32 &&
		reflect.ValueOf(data).Kind() != reflect.Float64 &&
		reflect.TypeOf(data) != reflect.TypeOf(json.Number("")) {
		return data
	}
	if reflect.TypeOf(data) == reflect.TypeOf(json.Number("")) {
		jsonInt, err := data.(json.Number).Int64()
		if err != nil {
			panic(err)
		}
		val, _ := convertStringValue(fmt.Sprintf("%d", jsonInt), vField)
		return val
	}
	if reflect.ValueOf(data).Kind() == reflect.Float32 {
		val, _ := convertStringValue(fmt.Sprintf("%.0f", data.(float32)), vField)
		return val
	}
	val, _ := convertStringValue(fmt.Sprintf("%.0f", data.(float64)), vField)
	return val
}
func parseForFloat(data interface{}, vField reflect.Value) float64 {
	if reflect.TypeOf(data) == reflect.TypeOf(json.Number("")) {
		floatData, err := data.(json.Number).Float64()
		if err != nil {
			panic(err)
		}
		return floatData
	}
	if vField.Kind() == reflect.Float32 {
		return float64(data.(float32))
	}
	return data.(float64)
}

func affect(data interface{}, vField reflect.Value, noDefaultVal bool) error {
	switch vField.Kind() {
	case reflect.String:
		vField.SetString(data.(string))
		break
	case reflect.Int:
		vField.SetInt(int64(parseForInt(data, vField).(int)))
		break
	case reflect.Int8:
		vField.SetInt(int64(parseForInt(data, vField).(int8)))
		break
	case reflect.Int16:
		vField.SetInt(int64(parseForInt(data, vField).(int16)))
		break
	case reflect.Int32:
		vField.SetInt(int64(parseForInt(data, vField).(int32)))
		break
	case reflect.Int64:
		vField.SetInt(parseForInt(data, vField).(int64))
		break
	case reflect.Uint:
		vField.SetUint(uint64(parseForInt(data, vField).(uint)))
		break
	case reflect.Uint8:
		vField.SetUint(uint64(parseForInt(data, vField).(uint8)))
		break
	case reflect.Uint16:
		vField.SetUint(uint64(parseForInt(data, vField).(uint16)))
		break
	case reflect.Uint32:
		vField.SetUint(uint64(parseForInt(data, vField).(uint32)))
		break
	case reflect.Uint64:
		vField.SetUint(parseForInt(data, vField).(uint64))
		break
	case reflect.Slice:
		if vField.IsNil() {
			vField.Set(reflect.MakeSlice(reflect.SliceOf(vField.Type().Elem()), 0, 0))
		}
		if reflect.ValueOf(data).Kind() != reflect.Slice {
			return errors.New(fmt.Sprintf("Type '%s' have not receive a slice.", vField.String()))
		}

		dataValue := reflect.ValueOf(data)
		if dataValue.Type().Kind() == reflect.Interface {
			dataValue = dataValue.Elem()
		}
		for i := 0; i < dataValue.Len(); i++ {
			var newElem reflect.Value
			dataValueElem := dataValue.Index(i)
			if dataValueElem.Type().Kind() == reflect.Interface {
				dataValueElem = dataValueElem.Elem()
			}
			newElem = dataValueElem
			if dataValueElem.Type() == reflect.TypeOf(make(map[string]interface{})) {
				newElem = reflect.New(vField.Type().Elem())
				UnmarshalToValue(dataValueElem.Interface().(map[string]interface{}), newElem, noDefaultVal)
				newElem = newElem.Elem()
			}
			vField.Set(reflect.Append(vField, newElem))
		}
		break
	case reflect.Interface:
		vField.Set(reflect.ValueOf(data))
		break
	case reflect.Bool:
		vField.SetBool(data.(bool))
		break
	case reflect.Float32:
		vField.SetFloat(parseForFloat(data, vField))
		break
	case reflect.Float64:
		vField.SetFloat(parseForFloat(data, vField))
		break
	case reflect.Ptr:
		if vField.IsNil() {
			vField.Set(reflect.New(vField.Type().Elem()))
		}
		err := affect(data, vField.Elem(), noDefaultVal)
		if err != nil {
			return err
		}
		break
	default:
		servUriType := reflect.TypeOf(ServiceUri{})
		if vField.Type() != servUriType && reflect.TypeOf(data) != reflect.TypeOf(make(map[string]interface{})) {
			return NewErrTypeNotSupported(vField)
		}
		if vField.Kind() == reflect.Map &&
			reflect.TypeOf(data) == reflect.TypeOf(make(map[string]interface{})) {
			return unmarshalUntypedMap(data.(map[string]interface{}), vField, noDefaultVal)
		}
		if reflect.TypeOf(data) == reflect.TypeOf(make(map[string]interface{})) {
			return UnmarshalToValue(data.(map[string]interface{}), vField, noDefaultVal)
		}
		serviceUrl, err := url.Parse(data.(string))
		if err != nil {
			return err
		}
		serviceUri := urlToServiceUri(serviceUrl)
		vField.Set(reflect.ValueOf(serviceUri))
		break
	}
	return nil
}
func unmarshalUntypedMap(data map[string]interface{}, vField reflect.Value, noDefaultVal bool) error {
	if vField.Type() == reflect.TypeOf(make(map[string]interface{})) {
		vField.Set(reflect.ValueOf(data))
		return nil
	}
	if vField.IsNil() {
		vField.Set(reflect.MakeMap(vField.Type()))
	}
	for name, val := range data {
		if reflect.TypeOf(val) != reflect.TypeOf(make(map[string]interface{})) {
			vField.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(val))
			continue
		}
		typeElem := vField.Type().Elem()
		if typeElem.Kind() == reflect.Ptr {
			typeElem = typeElem.Elem()
		}
		newElem := reflect.New(typeElem)
		err := UnmarshalToValue(val.(map[string]interface{}), newElem, noDefaultVal)
		if err != nil {
			return err
		}
		if vField.Type().Elem().Kind() != reflect.Ptr {
			newElem = newElem.Elem()
		}
		vField.SetMapIndex(reflect.ValueOf(name), newElem)
	}
	return nil
}
func parseInTag(tag, fieldName string) Tag {
	if tag == "" {
		return Tag{
			Name: snakecase.SnakeCase(fieldName),
		}
	}
	tag = strings.TrimSpace(tag)
	splitedTag := strings.Split(tag, ",")
	name := splitedTag[0]
	skipped := false
	if name == skipTag {
		skipped = true
	}
	if name == "" {
		name = snakecase.SnakeCase(fieldName)
	}

	return Tag{
		Name:         name,
		Skip:         skipped,
		IsRegex:      hasRegexTag(splitedTag[1:]),
		DefaultValue: getDefaultTagValue(splitedTag[1:]),
	}
}
func hasRegexTag(tags []string) bool {
	for _, tag := range tags {
		if tag == regexTag {
			return true
		}
	}
	return false
}
func getDefaultTagValue(tags []string) string {

	for _, tag := range tags {
		splitedDefTag := strings.Split(tag, "=")
		if len(splitedDefTag) < 2 || splitedDefTag[0] != defaultTag {
			continue
		}
		return strings.TrimSpace(strings.Join(splitedDefTag[1:], "="))
	}
	return ""
}
func retrieveFinalData(typeOf reflect.Type, serviceCredentials map[string]interface{}, key, defaultValue string) interface{} {
	valueExists := isValueExists(serviceCredentials, key)
	if !valueExists &&
		defaultValue == "" &&
		typeOf.Kind() != reflect.Struct {
		return nil
	}
	if valueExists {
		return serviceCredentials[key]
	}
	if !valueExists && defaultValue != "" {
		return defaultValue
	}
	if !valueExists && typeOf.Kind() == reflect.Struct && reflect.TypeOf(ServiceUri{}) == typeOf {
		return make(map[string]interface{})
	}
	if !valueExists && typeOf.Kind() == reflect.Struct {
		return serviceCredentials
	}
	return nil
}
func isValueExists(serviceCredentials map[string]interface{}, key string) bool {
	if key == "" {
		return false
	}
	_, ok := serviceCredentials[key]
	return ok
}
func match(matcher, content string) bool {
	regex, err := regexp.Compile("(?i)^" + matcher + "$")
	if err != nil {
		return false
	}
	return regex.MatchString(content)
}
func getKeyFromRegex(serviceCredentials map[string]interface{}, regexKey string) string {
	for key, _ := range serviceCredentials {
		if match(regexKey, key) {
			return key
		}
	}
	return ""
}
func urlToServiceUri(url *url.URL) ServiceUri {
	username := ""
	password := ""
	if url.User != nil {
		if url.User.Username() != "" {
			username = url.User.Username()
		}
		_, hasPassword := url.User.Password()
		if hasPassword {
			password, _ = url.User.Password()
		}
	}
	queries := make([]QueryUri, 0)
	for key, value := range url.Query() {
		queries = append(queries, QueryUri{
			Key:   key,
			Value: value[0],
		})
	}
	host := url.Host
	port := 0
	splitedHost := strings.Split(host, ":")
	if len(splitedHost) == 2 {
		host = splitedHost[0]
		port, _ = strconv.Atoi(splitedHost[1])
	}
	return ServiceUri{
		Scheme:   url.Scheme,
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Name:     strings.TrimPrefix(url.Path, "/"),
		Query:    queries,
		RawQuery: url.RawQuery,
	}
}

func convertStringValue(defVal string, vField reflect.Value) (interface{}, error) {
	switch vField.Kind() {
	case reflect.String:
		return defVal, nil
	case reflect.Interface:
		return defVal, nil
	case reflect.Int:
		return strconv.Atoi(defVal)
	case reflect.Int8:
		val, err := strconv.ParseInt(defVal, 10, 8)
		if err != nil {
			return "", err
		}
		return int8(val), nil
	case reflect.Int16:
		val, err := strconv.ParseInt(defVal, 10, 16)
		if err != nil {
			return "", err
		}
		return int16(val), nil
	case reflect.Int32:
		val, err := strconv.ParseInt(defVal, 10, 32)
		if err != nil {
			return "", err
		}
		return int32(val), nil
	case reflect.Int64:
		val, err := strconv.ParseInt(defVal, 10, 64)
		if err != nil {
			return "", err
		}
		return int64(val), nil
	case reflect.Uint:
		val, err := strconv.ParseUint(defVal, 10, int(strconv.IntSize))
		if err != nil {
			return "", err
		}
		return uint(val), nil
	case reflect.Uint8:
		val, err := strconv.ParseUint(defVal, 10, 8)
		if err != nil {
			return "", err
		}
		return uint8(val), nil
	case reflect.Uint16:
		val, err := strconv.ParseUint(defVal, 10, 16)
		if err != nil {
			return "", err
		}
		return uint16(val), nil
	case reflect.Uint32:
		val, err := strconv.ParseUint(defVal, 10, 32)
		if err != nil {
			return "", err
		}
		return uint32(val), nil
	case reflect.Uint64:
		val, err := strconv.ParseUint(defVal, 10, 64)
		if err != nil {
			return "", err
		}
		return uint64(val), nil
	case reflect.Bool:
		return strconv.ParseBool(defVal)
	case reflect.Float32:
		val, err := strconv.ParseFloat(defVal, 32)
		if err != nil {
			return "", err
		}
		return float32(val), nil
	case reflect.Float64:
		val, err := strconv.ParseFloat(defVal, 64)
		if err != nil {
			return "", err
		}
		return float64(val), nil
	case reflect.Slice:
		finalField := reflect.MakeSlice(reflect.SliceOf(vField.Type().Elem()), 0, 0)
		defValSlice := strings.Split(defVal, ",")
		for _, aDefVal := range defValSlice {
			finDefVal, err := convertStringValue(strings.TrimSpace(aDefVal), reflect.New(vField.Type().Elem()))
			if err != nil {
				return "", err
			}
			finalField = reflect.Append(finalField, reflect.ValueOf(finDefVal))
		}
		return finalField.Interface(), nil
	case reflect.Ptr:
		if vField.IsNil() {
			vField.Set(reflect.New(vField.Type().Elem()))
		}
		return convertStringValue(defVal, vField.Elem())
	default:
		servUriType := reflect.TypeOf(ServiceUri{})
		if vField.Type() != servUriType {
			return "", NewErrTypeNotSupported(vField)
		}
		return defVal, nil
	}
	return "", NewErrTypeNotSupported(vField)
}
