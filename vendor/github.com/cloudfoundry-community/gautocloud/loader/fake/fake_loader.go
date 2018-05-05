package fake

import (
	"github.com/cloudfoundry-community/gautocloud/cloudenv"
	"github.com/cloudfoundry-community/gautocloud/connectors"
	"github.com/cloudfoundry-community/gautocloud/loader"
	"github.com/golang/mock/gomock"
	"log"
)

// Mock of Loader interface
type MockLoader struct {
	ctrl     *gomock.Controller
	recorder *_MockLoaderRecorder
}

// Recorder for MockLoader (not exported)
type _MockLoaderRecorder struct {
	mock *MockLoader
}

type MockTestReporter struct{}

func (g MockTestReporter) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR GAUTOCLOUD MOCK LOADER] "+format, args...)
}

func (g MockTestReporter) Fatalf(format string, args ...interface{}) {
	log.Fatalf("[FAIL GAUTOCLOUD MOCK LOADER] "+format, args...)
}
func NewMockLoader() *MockLoader {

	mock := &MockLoader{ctrl: gomock.NewController(MockTestReporter{})}
	mock.recorder = &_MockLoaderRecorder{mock}
	return mock
}

func (_m *MockLoader) EXPECT() *_MockLoaderRecorder {
	return _m.recorder
}

func (_m *MockLoader) CleanConnectors() {
	_m.ctrl.Call(_m, "CleanConnectors")
}

func (_mr *_MockLoaderRecorder) CleanConnectors() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CleanConnectors")
}

func (_m *MockLoader) CloudEnvs() []cloudenv.CloudEnv {
	ret := _m.ctrl.Call(_m, "CloudEnvs")
	ret0, _ := ret[0].([]cloudenv.CloudEnv)
	return ret0
}

func (_mr *_MockLoaderRecorder) CloudEnvs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CloudEnvs")
}

func (_m *MockLoader) Connectors() map[string]connectors.Connector {
	ret := _m.ctrl.Call(_m, "Connectors")
	ret0, _ := ret[0].(map[string]connectors.Connector)
	return ret0
}

func (_mr *_MockLoaderRecorder) Connectors() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connectors")
}

func (_m *MockLoader) CurrentCloudEnv() cloudenv.CloudEnv {
	ret := _m.ctrl.Call(_m, "CurrentCloudEnv")
	ret0, _ := ret[0].(cloudenv.CloudEnv)
	return ret0
}

func (_mr *_MockLoaderRecorder) CurrentCloudEnv() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CurrentCloudEnv")
}

func (_m *MockLoader) GetAll(_param0 string) ([]interface{}, error) {
	ret := _m.ctrl.Call(_m, "GetAll", _param0)
	ret0, _ := ret[0].([]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockLoaderRecorder) GetAll(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetAll", arg0)
}

func (_m *MockLoader) GetAppInfo() cloudenv.AppInfo {
	ret := _m.ctrl.Call(_m, "GetAppInfo")
	ret0, _ := ret[0].(cloudenv.AppInfo)
	return ret0
}

func (_mr *_MockLoaderRecorder) GetAppInfo() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetAppInfo")
}

func (_m *MockLoader) GetFirst(_param0 string) (interface{}, error) {
	ret := _m.ctrl.Call(_m, "GetFirst", _param0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockLoaderRecorder) GetFirst(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFirst", arg0)
}

func (_m *MockLoader) Inject(_param0 interface{}) error {
	ret := _m.ctrl.Call(_m, "Inject", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockLoaderRecorder) Inject(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Inject", arg0)
}

func (_m *MockLoader) InjectFromId(_param0 string, _param1 interface{}) error {
	ret := _m.ctrl.Call(_m, "InjectFromId", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockLoaderRecorder) InjectFromId(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "InjectFromId", arg0, arg1)
}

func (_m *MockLoader) IsInACloudEnv() bool {
	ret := _m.ctrl.Call(_m, "IsInACloudEnv")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockLoaderRecorder) IsInACloudEnv() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsInACloudEnv")
}

func (_m *MockLoader) RegisterConnector(_param0 connectors.Connector) {
	_m.ctrl.Call(_m, "RegisterConnector", _param0)
}

func (_mr *_MockLoaderRecorder) RegisterConnector(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RegisterConnector", arg0)
}

func (_m *MockLoader) ReloadConnectors() {
	_m.ctrl.Call(_m, "ReloadConnectors")
}

func (_mr *_MockLoaderRecorder) ReloadConnectors() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReloadConnectors")
}

func (_m *MockLoader) ShowPreviousLog() {
	_m.ctrl.Call(_m, "ShowPreviousLog")
}

func (_mr *_MockLoaderRecorder) ShowPreviousLog() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ShowPreviousLog")
}

func (_m *MockLoader) Store() map[string][]loader.StoredService {
	ret := _m.ctrl.Call(_m, "Store")
	ret0, _ := ret[0].(map[string][]loader.StoredService)
	return ret0
}

func (_mr *_MockLoaderRecorder) Store() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Store")
}
