package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"errors"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTPResponseStatusCode(t *testing.T) {
	Convey("An nil error should have no HTTP status code", t, func() {
		_, present := HTTPResponseStatusCode(nil)
		So(present, ShouldBeFalse)
	})
	Convey("A foreign error should have no detectable HTTP status code", t, func() {
		_, present := HTTPResponseStatusCode(errors.New("other"))
		So(present, ShouldBeFalse)
	})
	Convey("A fargo error generated from a response from Eureka", t, func() {
		verify := func(err *unsuccessfulHTTPResponse) {
			Convey("should have the given HTTP status code", func() {
				code, present := HTTPResponseStatusCode(err)
				So(present, ShouldBeTrue)
				So(code, ShouldEqual, err.statusCode)
				Convey("should produce a message", func() {
					msg := err.Error()
					if len(err.messagePrefix) == 0 {
						Convey("that lacks a prefx", func() {
							So(msg, ShouldNotStartWith, ",")
						})
					} else {
						Convey("that starts with the given prefix", func() {
							So(msg, ShouldStartWith, err.messagePrefix)
						})
					}
					Convey("that contains the status code in decimal notation", func() {
						So(msg, ShouldContainSubstring, strconv.Itoa(err.statusCode))
					})
				})
			})
		}
		Convey("with a message prefix", func() {
			verify(&unsuccessfulHTTPResponse{500, "operation failed"})
		})
		Convey("without a message prefix", func() {
			verify(&unsuccessfulHTTPResponse{statusCode: 500})
		})
	})
}
