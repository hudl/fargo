package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"os"
	"strings"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("fargo")
var metadataLog = logging.MustGetLogger("fargo.metadata")
var marshalLog = logging.MustGetLogger("fargo.marshal")
var logLevel = logging.INFO
func init() {
	switch levelOverride := strings.ToUpper(os.Getenv("FARGO_LOG_LEVEL")); levelOverride {
		case "DEBUG":
			logLevel = logging.DEBUG
		case "INFO":
			logLevel = logging.INFO
		case "NOTICE":
			logLevel = logging.NOTICE
		case "WARNING":
			logLevel = logging.WARNING
		case "ERROR":
			logLevel = logging.ERROR
		case "CRITICAL":
			logLevel = logging.CRITICAL
 	}
	logging.SetLevel(logLevel, "")

	logging.SetLevel(logging.WARNING, "fargo.metadata")
	logging.SetLevel(logging.WARNING, "fargo.marshal")
}
