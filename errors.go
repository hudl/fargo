package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
)

type unsuccessfulHTTPResponse struct {
	statusCode    int
	messagePrefix string
}

func (u *unsuccessfulHTTPResponse) Error() string {
	if len(u.messagePrefix) > 0 {
		return fmt.Sprint(u.messagePrefix, ", rcode = ", u.statusCode)
	}
	return fmt.Sprint("rcode = ", u.statusCode)
}

// HTTPResponseStatusCode extracts the HTTP status code for the response from Eureka that motivated
// the supplied error, if any. If the returned present value is true, the returned code is an HTTP
// status code.
func HTTPResponseStatusCode(err error) (code int, present bool) {
	if u, ok := err.(*unsuccessfulHTTPResponse); ok {
		return u.statusCode, true
	}
	return 0, false
}

type AppNotFoundError struct {
	specific string
}

func (e AppNotFoundError) Error() string {
	return "Application not found for name=" + e.specific
}
