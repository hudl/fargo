package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"os"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("fargo")
var metadataLog = logging.MustGetLogger("fargo.metadata")
var marshalLog = logging.MustGetLogger("fargo.marshal")

func init() {
	if len(os.Getenv("FARGO_VERBOSE")) > 0 {
		logging.SetLevel(logging.DEBUG, "")
	}
	logging.SetLevel(logging.WARNING, "fargo.metadata")
	logging.SetLevel(logging.WARNING, "fargo.marshal")
}
