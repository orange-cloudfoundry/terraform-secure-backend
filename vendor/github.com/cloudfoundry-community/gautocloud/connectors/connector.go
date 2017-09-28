// It handles the conversion of a service to a real client or structure which can be manipulated after by user.
package connectors



// this is the interface to be implemented to create a new connector
// You should add an init function in the same package of your connector and register it automatically in gautocloud when importing your connector
// Example of init function:
//  func init() {
//    gautocloud.RegisterConnector(NewMyConnector())
//  }
// see implementation of any raw connectors to see how to implement a connector
type Connector interface {
	// This is the id of your connector and it must be unique and not have the same id of another connector
	// Note: if a connector id is already taken gautocloud will complain
	Id() string
	// Name is the name of a service to lookup in the cloud environment
	// Note: a regex can be passed
	Name() string
	// This should return a list of tags which designed what kind of service you want
	// example: mysql, database, rmdb ...
	// Note: a regex can be passed on each tag
	Tags() []string
	// The parameter is a filled schema you gave in the function Schema
	// The first value to return is what you want and you have no obligation to give always the same type. gautocloud is interface agnostic
	// You can give an error if an error occurred, this error will appear in logs
	Load(interface{}) (interface{}, error)
	// It must return a structure
	// this structure will be used by the decoder to create a structure of the same type and filled with service's credentials found by a cloud environment
	// Here an example of what kind of structure you can return:
	//  type MyStruct struct {
	//  // Name is key of a service credentials, decoder will look at any matching credentials which have the key name and will pass the value of this credentials
	//  	Name    string `cloud:"name"`           // note: by default if you don't provide a cloud tag the key will be the field name in snake_case
	//  	Uri     decoder.ServiceUri              // ServiceUri is a special type. Decoder will expect an uri as a value and will give a ServiceUri
	//  	User    string `cloud:".*user.*,regex"` // by passing `regex` in cloud tag it will say to decoder that the expected key must be match the regex
	//  	Password string `cloud:".*user.*,regex" cloud-default:"apassword"` // by passing a tag named `cloud-default` decoder will understand that if the key is not found it must fill the field with this value
	//  }
	Schema() interface{}
}