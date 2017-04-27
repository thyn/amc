package design

import (
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("amc", func() { // API defines the microservice endpoint and
	Title("Aerospike Management Console")                                                                  // other global properties. There should be one
	Description("Aerospike Management Console is used to manage and monitor Aerospike Database Clusters.") // and exactly one API definition appearing in
	Scheme("http", "https")                                                                                // the design.
	BasePath("api/v1")                                                                                     // Base path to all API endpoints
	Consumes("application/json")                                                                           // Media types supported by the API
	Produces("application/json")                                                                           // Media types generated by the API
	Origin("*", func() {                                                                                   // Define CORS policy, may be prefixed with "*" wildcard
		Headers("*")                     // One or more authorized headers, use "*" to authorize all
		Methods("GET", "POST", "DELETE") // One or more authorized HTTP methods
		Expose("X-Time")                 // One or more headers exposed to clients
		MaxAge(600)                      // How long to cache a preflight request response
		Credentials()                    // Sets Access-Control-Allow-Credentials header
	})
})