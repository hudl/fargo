package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"net/http"
)

type httpOperation byte

const (
	opRegistration httpOperation = iota
	opDeregistration
	opMetadataUpdate
	opStatusUpdate
	opLeaseRenewal
	opRetrieval
)

type unsuccessfulHTTPResponse struct {
	op            httpOperation
	statusCode    int
	messagePrefix string
}

func (u unsuccessfulHTTPResponse) Error() string {
	return fmt.Sprint(u.messagePrefix, ", rcode = ", u.statusCode)
}

func httpResponseIndicatesInvalidInstance(op httpOperation, statusCode int) bool {
	switch op {
	case opRegistration:
		return statusCode == http.StatusBadRequest
	default:
		return false
	}
}

// InstanceWasInvalid returns true if the error arose during an instance registration attempt
// due to Eureka rejecting the proposed instance as invalid.
func InstanceWasInvalid(err error) bool {
	if u, ok := err.(*unsuccessfulHTTPResponse); ok {
		return httpResponseIndicatesInvalidInstance(u.op, u.statusCode)
	}
	return false
}

func httpResponseIndicatesMissingInstance(op httpOperation, statusCode int) bool {
	switch op {
	case opDeregistration, opStatusUpdate, opLeaseRenewal, opRetrieval:
		return statusCode == http.StatusNotFound
	case opMetadataUpdate:
		return statusCode == http.StatusInternalServerError
	default:
		return false
	}
}

// InstanceWasMissing returns true if the error arose during an instance retrieval, status update,
// metadata update, or lease renewal, or deregistration attempt due to the target instance not being
// registered, being unknown to Eureka.
func InstanceWasMissing(err error) bool {
	if u, ok := err.(*unsuccessfulHTTPResponse); ok {
		return httpResponseIndicatesMissingInstance(u.op, u.statusCode)
	}
	return false
}

type AppNotFoundError struct {
	specific string
}

func (e AppNotFoundError) Error() string {
	return "Application not found for name=" + e.specific
}
