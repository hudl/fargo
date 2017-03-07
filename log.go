package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"github.com/op/go-logging"
)

type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Critical(args ...interface{})
	Criticalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Notice(args ...interface{})
	Noticef(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

var log, metadataLog, marshalLog Logger

func SetLogger(l Logger) {
	log = l
}

func SetMarshalLogger(l Logger) {
	metadataLog = l
}

func SetMetadataLogger(l Logger) {
	marshalLog = l
}

func init() {
	SetLogger(logging.MustGetLogger("fargo"))
	SetMarshalLogger(logging.MustGetLogger("fargo.metadata"))
	SetMetadataLogger(logging.MustGetLogger("fargo.marshal"))

	logging.SetLevel(logging.WARNING, "fargo.metadata")
	logging.SetLevel(logging.WARNING, "fargo.marshal")
}
