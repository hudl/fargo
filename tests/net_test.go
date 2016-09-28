package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
)

func shouldNotBearAnHTTPStatusCode(actual interface{}, expected ...interface{}) string {
	if code, present := fargo.HTTPResponseStatusCode(actual.(error)); present {
		return fmt.Sprintf("Expected: no HTTP status code\nActual:   %d", code)
	}
	return ""
}

func shouldBearHTTPStatusCode(actual interface{}, expected ...interface{}) string {
	expectedCode := expected[0]
	code, present := fargo.HTTPResponseStatusCode(actual.(error))
	if !present {
		return fmt.Sprintf("Expected: %d\nActual:   no HTTP status code", expectedCode)
	}
	if code != expectedCode {
		return fmt.Sprintf("Expected: %d\nActual:   %d", expectedCode, code)
	}
	return ""
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

func TestGetApps(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Pull applications", t, func() {
			a, _ := e.GetApps()
			So(len(a["EUREKA"].Instances), ShouldEqual, 2)
		})
		Convey("Pull single application", t, func() {
			a, _ := e.GetApp("EUREKA")
			So(len(a.Instances), ShouldEqual, 2)
			for _, ins := range a.Instances {
				So(ins.IPAddr, ShouldBeIn, []string{"172.17.0.2", "172.17.0.3"})
			}
		})
	}
}

func TestRegistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
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
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Fail to heartbeat a non-existent instance", t, func() {
			j := fargo.Instance{
				HostName:         "i-6543",
				Port:             9090,
				App:              "TESTAPP",
				IPAddr:           "127.0.0.10",
				VipAddress:       "127.0.0.10",
				DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
				SecureVipAddress: "127.0.0.10",
				Status:           fargo.UP,
			}
			err := e.HeartBeatInstance(&j)
			So(err, ShouldNotBeNil)
			So(err, shouldBearHTTPStatusCode, http.StatusNotFound)
		})
		Convey("Register an instance to TESTAPP", t, func() {
			Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				So(err, ShouldBeNil)
			})
			Convey("Instance can check in", func() {
				err := e.HeartBeatInstance(&i)
				So(err, ShouldBeNil)
			})
		})
	}
}

func TestReregistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")

	for _, j := range []bool{false, true} {
		e.UseJson = j

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

		Convey("Register a TESTAPP instance", t, func() {
			Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				So(err, ShouldBeNil)
			})
		})

		Convey("Reregister the TESTAPP instance", t, func() {
			Convey("Instance reregisters correctly", func() {
				err := e.ReregisterInstance(&i)
				So(err, ShouldBeNil)
			})

			Convey("Instance can check in", func() {
				err := e.HeartBeatInstance(&i)
				So(err, ShouldBeNil)
			})

			Convey("Instance can be gotten correctly", func() {
				ii, err := e.GetInstance(i.App, i.HostName)
				So(err, ShouldBeNil)
				So(ii.App, ShouldEqual, i.App)
				So(ii.HostName, ShouldEqual, i.HostName)
			})
		})
	}
}

func DontTestDeregistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
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
	Convey("Register a TESTAPP instance", t, func() {
		Convey("Instance registers correctly", func() {
			err := e.RegisterInstance(&i)
			So(err, ShouldBeNil)
		})
	})
	Convey("Deregister the TESTAPP instance", t, func() {
		Convey("Instance deregisters correctly", func() {
			err := e.DeregisterInstance(&i)
			So(err, ShouldBeNil)
		})
		Convey("Instance cannot check in", func() {
			err := e.HeartBeatInstance(&i)
			So(err, ShouldNotBeNil)
			So(err, shouldBearHTTPStatusCode, http.StatusNotFound)
		})
	})
}

func TestUpdateStatus(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
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
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Register an instance to TESTAPP", t, func() {
			Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				So(err, ShouldBeNil)
			})
		})
		Convey("Update an instance status", t, func() {
			Convey("Instance updates to OUT_OF_SERVICE correctly", func() {
				err := e.UpdateInstanceStatus(&i, fargo.OUTOFSERVICE)
				So(err, ShouldBeNil)
			})
			Convey("Instance updates to UP corectly", func() {
				err := e.UpdateInstanceStatus(&i, fargo.UP)
				So(err, ShouldBeNil)
			})
		})
	}
}

func TestMetadataReading(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Read empty instance metadata", t, func() {
			a, err := e.GetApp("EUREKA")
			So(err, ShouldBeNil)
			i := a.Instances[0]
			_, err = i.Metadata.GetString("SomeProp")
			So(err, ShouldBeNil)
		})
		Convey("Read valid instance metadata", t, func() {
			a, err := e.GetApp("TESTAPP")
			So(err, ShouldBeNil)
			So(len(a.Instances), ShouldBeGreaterThan, 0)
			if len(a.Instances) == 0 {
				return
			}
			i := a.Instances[0]
			err = e.AddMetadataString(i, "SomeProp", "AValue")
			So(err, ShouldBeNil)
			v, err := i.Metadata.GetString("SomeProp")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "AValue")
		})
	}
}
