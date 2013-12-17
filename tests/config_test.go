package fargo_test

/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2013 Ryan S. Brown <sb@ryansb.com>
 * Copyright (c) 2013 Hudl <@Hudl>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */

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
			So(conf.Eureka.ServiceUrls, ShouldContain, "http://172.16.0.11:8080/eureka/v2")
			So(conf.Eureka.ServiceUrls, ShouldContain, "http://172.16.0.22:8080/eureka/v2")
		})
		So(conf.Eureka.UseDNSForServiceUrls, ShouldEqual, false)
	})
}
