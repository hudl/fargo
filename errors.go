package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"net/http"
)

type unsuccessfulHTTPResponse struct {
	statusCode    int
	messagePrefix string
}

func (u unsuccessfulHTTPResponse) Error() string {
	return fmt.Sprint(u.messagePrefix, ", rcode = ", u.statusCode)
}

type unsuccessfulRegistrationResponse struct {
	unsuccessfulHTTPResponse
}

type unsuccessfulDeregistrationResponse struct {
	unsuccessfulHTTPResponse
}

type unsuccessfulUpdateResponse struct {
	unsuccessfulHTTPResponse
}

type unsuccessfulMetadataUpdateResponse struct {
	unsuccessfulHTTPResponse
}

type unsuccessfulRetrievalResponse struct {
	unsuccessfulHTTPResponse
}

// InstanceWasInvalid returns true if the error arose during an instance registration attempt
// due to Eureka rejecting the proposed instance as invalid.
func InstanceWasInvalid(err error) bool {
	if u, ok := err.(unsuccessfulRegistrationResponse); ok {
		return u.statusCode == http.StatusBadRequest
	}
	return false
}

// InstanceWasMissing returns true if the error arose during an instance retrieval, status update,
// metadata update, or deregistration attempt due to the target instance not being registered,
// being unknown to Eureka.
func InstanceWasMissing(err error) bool {
	switch u := err.(type) {
	case unsuccessfulDeregistrationResponse:
		return u.statusCode == http.StatusNotFound
	case unsuccessfulUpdateResponse:
		return u.statusCode == http.StatusNotFound
	case unsuccessfulMetadataUpdateResponse:
		return u.statusCode == http.StatusInternalServerError
	case unsuccessfulRetrievalResponse:
		return u.statusCode == http.StatusNotFound
	}
	return false
}

type AppNotFoundError struct {
	specific string
}

func (e AppNotFoundError) Error() string {
	return "Application not found for name=" + e.specific
}
