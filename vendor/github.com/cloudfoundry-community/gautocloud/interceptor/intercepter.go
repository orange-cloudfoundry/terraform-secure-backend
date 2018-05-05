// This permit to intercept data which will be given back by gautocloud and modified it before giving back to user.
// Interceptor should be used in a connector, to do so, connector have to implement ConnectorIntercepter:
//  type ConnectorIntercepter interface {
//  	Intercepter() interceptor.Intercepter
//  }
// An interceptor work like a http middleware.
package interceptor

// This is the interface to implement to create an interceptor
type Intercepter interface {
	// Current is interface given by user when using gautocloud.Inject(interfaceFromUser{}),
	// this can be nil if user doesn't use Inject functions from gautocloud.
	//
	// Found is interface found by gautocloud
	//
	// It should return an interface which must be the same type as found.
	// Tips: current and found have always the same type, this type is the type given by connector from its function Schema()
	Intercept(current, found interface{}) (interface{}, error)
}

// The IntercepterFunc type is an adapter to allow the use of
// ordinary functions as Intercepter. If f is a function
// with the appropriate signature, IntercepterFunc(f) is a
// Intercepter that calls f.
type IntercepterFunc func(current, found interface{}) (interface{}, error)

func (f IntercepterFunc) Intercept(current, found interface{}) (interface{}, error) {
	return f(current, found)
}
