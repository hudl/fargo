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

func TestGetApps(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	Convey("Pull applications", t, func() {
		a, _ := e.GetApps()
		So(len(a["EUREKA"].Instances), ShouldEqual, 2)
	})
	Convey("Pull single application", t, func() {
		a, _ := e.GetApp("EUREKA")
		So(len(a.Instances), ShouldEqual, 2)
		for idx, ins := range a.Instances {
			if ins.HostName == "node1.localdomain" {
				So(a.Instances[idx].IPAddr, ShouldEqual, "172.16.0.11")
				So(a.Instances[idx].HostName, ShouldEqual, "node1.localdomain")
			} else {
				So(a.Instances[idx].IPAddr, ShouldEqual, "172.16.0.22")
				So(a.Instances[idx].HostName, ShouldEqual, "node2.localdomain")
			}
		}
	})
}

func TestRegistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	Convey("Register an instance to TESTAPP", t, func() {
		i := fargo.Instance{
			HostName:         "i-123456",
			Port:             9090,
			App:              "TESTAPP",
			IPAddr:           "127.0.0.10",
			VipAddress:       "127.0.0.10",
			DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
			SecureVipAddress: "127.0.0.10",
			Status:           fargo.UP,
		}
		err := e.RegisterInstance(&i)
		So(err, ShouldBeNil)
	})
	Convey("Check in for TESTAPP", t, func() {
		i := fargo.Instance{
			HostName:         "i-123456",
			Port:             9090,
			App:              "TESTAPP",
			IPAddr:           "127.0.0.10",
			VipAddress:       "127.0.0.10",
			DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
			SecureVipAddress: "127.0.0.10",
			Status:           fargo.UP,
		}
		err := e.HeartBeatInstance(&i)
		So(err, ShouldBeNil)
	})
}

func TestConnectionCreation(t *testing.T) {
	Convey("Pull applications", t, func() {
		cfg, err := fargo.ReadConfig("./config_sample/local.gcfg")
		So(err, ShouldBeNil)
		e := fargo.NewConnFromConfig(cfg)
		apps, err := e.GetApps()
		So(err, ShouldBeNil)
		So(len(apps["EUREKA"].Instances), ShouldEqual, 2)
	})
}
