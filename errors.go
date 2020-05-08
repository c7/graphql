package graphql

import "strings"

// Errors contains all the errors that were returned by the GraphQL server.
type Errors []Error

func (ee Errors) Error() string {
	if len(ee) == 0 {
		return "no errors"
	}

	errs := make([]string, len(ee))
	for i, e := range ee {
		errs[i] = e.Message
	}

	return "graphql: " + strings.Join(errs, "; ")
}

// An Error contains error information returned by the GraphQL server.
type Error struct {
	// Message contains the error message.
	Message string
	// Locations contains the locations in the GraphQL document that caused the
	// error if the error can be associated to a particular point in the
	// requested GraphQL document.
	Locations []Location
	// Path contains the key path of the response field which experienced the
	// error. This allows clients to identify whether a nil result is
	// intentional or caused by a runtime error.
	Path []interface{}
	// Extensions may contain additional fields set by the GraphQL service,
	// such as	an error code.
	Extensions map[string]interface{}
}

// A Location is a location in the GraphQL query that resulted in an error.
// The location may be returned as part of an error response.
type Location struct {
	Line   int
	Column int
}

func (e Error) Error() string {
	return "graphql: " + e.Message
}
