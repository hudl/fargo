package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestConfigs(t *testing.T) {
	Convey("Reading a blank config to test defaults.", t, func() {
		conf, err := fargo.ReadConfig("./config_sample/blank.gcfg")
		So(err, ShouldBeNil)
		So(conf.Eureka.InTheCloud, ShouldEqual, false)
		So(conf.Eureka.ConnectTimeoutSeconds, ShouldEqual, 10)
		So(conf.Eureka.UseDNSForServiceUrls, ShouldEqual, false)
		So(conf.Eureka.ServerDNSName, ShouldEqual, "")
		So(len(conf.Eureka.ServiceUrls), ShouldEqual, 0)
		So(conf.Eureka.ServerPort, ShouldEqual, 7001)
		So(conf.Eureka.PollIntervalSeconds, ShouldEqual, 30)
		So(conf.Eureka.EnableDelta, ShouldEqual, false)
		So(conf.Eureka.PreferSameZone, ShouldEqual, false)
		So(conf.Eureka.RegisterWithEureka, ShouldEqual, false)
	})

	Convey("Testing a config that connects to local eureka instances", t, func() {
		conf, err := fargo.ReadConfig("./config_sample/local.gcfg")
		So(err, ShouldBeNil)
		So(conf.Eureka.InTheCloud, ShouldEqual, false)
		So(conf.Eureka.ConnectTimeoutSeconds, ShouldEqual, 2)
		Convey("Both test servers should be in the service URL list", func() {
			So(conf.Eureka.ServiceUrls, ShouldContain, "http://172.17.0.2:8080/eureka/v2")
			So(conf.Eureka.ServiceUrls, ShouldContain, "http://172.17.0.3:8080/eureka/v2")
		})
		So(conf.Eureka.UseDNSForServiceUrls, ShouldEqual, false)
	})
}
