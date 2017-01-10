package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

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

func withCustomRegisteredInstance(e *fargo.EurekaConnection, application string, hostName string, f func(i *fargo.Instance)) func() {
	return func() {
		vipAddress := "app"
		i := &fargo.Instance{
			HostName:         hostName,
			Port:             9090,
			App:              application,
			IPAddr:           "127.0.0.10",
			VipAddress:       vipAddress,
			DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
			SecureVipAddress: vipAddress,
			Status:           fargo.UP,
			LeaseInfo: fargo.LeaseInfo{
				DurationInSecs: 90,
			},
		}
		So(e.ReregisterInstance(i), ShouldBeNil)

		var wg sync.WaitGroup
		stop := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					if err := e.HeartBeatInstance(i); err != nil {
						if code, ok := fargo.HTTPResponseStatusCode(err); ok && code == http.StatusNotFound {
							e.ReregisterInstance(i)
						}
					}
				}
			}
		}()

		Reset(func() {
			close(stop)
			wg.Wait()
			So(e.DeregisterInstance(i), ShouldBeNil)
		})

		f(i)
	}
}

func withRegisteredInstance(e *fargo.EurekaConnection, f func(i *fargo.Instance)) func() {
	return withCustomRegisteredInstance(e, "TESTAPP", "i-123456", f)
}

func TestConnectionCreation(t *testing.T) {
	Convey("Pull applications", t, func() {
		cfg, err := fargo.ReadConfig("./config_sample/local.gcfg")
		So(err, ShouldBeNil)
		e := fargo.NewConnFromConfig(cfg)
		apps, err := e.GetApps()
		So(err, ShouldBeNil)
		app := apps["EUREKA"]
		So(app, ShouldNotBeNil)
		So(len(app.Instances), ShouldEqual, 2)
	})
}

func TestGetApps(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Pull applications", t, func() {
			apps, err := e.GetApps()
			So(err, ShouldBeNil)
			app := apps["EUREKA"]
			So(app, ShouldNotBeNil)
			So(len(app.Instances), ShouldEqual, 2)
		})
		Convey("Pull single application", t, func() {
			app, err := e.GetApp("EUREKA")
			So(err, ShouldBeNil)
			So(app, ShouldNotBeNil)
			So(len(app.Instances), ShouldEqual, 2)
			for _, ins := range app.Instances {
				So(ins.IPAddr, ShouldBeIn, []string{"172.17.0.2", "172.17.0.3"})
			}
		})
	}
}

func TestGetInstancesByNonexistentVIPAddress(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	for _, e.UseJson = range []bool{false, true} {
		Convey("Get instances by VIP address", t, func() {
			Convey("when the VIP address has no instances", func() {
				instances, err := e.GetInstancesByVIPAddress("nonexistent", false)
				So(err, ShouldBeNil)
				So(instances, ShouldBeEmpty)
			})
			Convey("when the secure VIP address has no instances", func() {
				instances, err := e.GetInstancesByVIPAddress("nonexistent", true)
				So(err, ShouldBeNil)
				So(instances, ShouldBeEmpty)
			})
		})
	}
}

func TestGetSingleInstanceByVIPAddress(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	cacheDelay := 35 * time.Second
	vipAddress := "app"
	for _, e.UseJson = range []bool{false, true} {
		Convey("When the VIP address has one instance", t, withRegisteredInstance(&e, func(instance *fargo.Instance) {
			time.Sleep(cacheDelay)
			instances, err := e.GetInstancesByVIPAddress(vipAddress, false)
			So(err, ShouldBeNil)
			So(instances, ShouldHaveLength, 1)
			So(instances[0].VipAddress, ShouldEqual, vipAddress)
			Convey("requesting the instances by that VIP address with status UP should provide that one", func() {
				instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.ThatAreUp)
				So(err, ShouldBeNil)
				So(instances, ShouldHaveLength, 1)
				So(instances[0].VipAddress, ShouldEqual, vipAddress)
				Convey("and when the instance has a status other than UP", func() {
					otherStatus := fargo.OUTOFSERVICE
					So(otherStatus, ShouldNotEqual, fargo.UP)
					err := e.UpdateInstanceStatus(instance, otherStatus)
					So(err, ShouldBeNil)
					Convey("selecting instances with that other status should provide that one", func() {
						time.Sleep(cacheDelay)
						instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.WithStatus(otherStatus))
						So(err, ShouldBeNil)
						So(instances, ShouldHaveLength, 1)
						Convey("And selecting instances with status UP should provide none", func() {
							// Ensure that we tolerate a nil option safely.
							instances, err := e.GetInstancesByVIPAddress(vipAddress, false, fargo.ThatAreUp, nil)
							So(err, ShouldBeNil)
							So(instances, ShouldBeEmpty)
						})
					})
				})
			})
		}))
		Convey("When the secure VIP address has one instance", t, withRegisteredInstance(&e, func(_ *fargo.Instance) {
			Convey("requesting the instances by that VIP address should provide that one", func() {
				time.Sleep(cacheDelay)
				// Ensure that we tolerate a nil option safely.
				instances, err := e.GetInstancesByVIPAddress(vipAddress, true, nil)
				So(err, ShouldBeNil)
				So(instances, ShouldHaveLength, 1)
				So(instances[0].SecureVipAddress, ShouldEqual, vipAddress)
			})
		}))
		time.Sleep(cacheDelay)
	}
}

func TestGetMultipleInstancesByVIPAddress(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	cacheDelay := 35 * time.Second
	for _, e.UseJson = range []bool{false, true} {
		Convey("When the VIP address has one instance", t, withRegisteredInstance(&e, func(instance *fargo.Instance) {
			Convey("when the VIP address has two instances", withCustomRegisteredInstance(&e, "TESTAPP2", "i-234567", func(_ *fargo.Instance) {
				Convey("requesting the instances by that VIP address should provide them", func() {
					time.Sleep(cacheDelay)
					vipAddress := "app"
					instances, err := e.GetInstancesByVIPAddress(vipAddress, false)
					So(err, ShouldBeNil)
					So(instances, ShouldHaveLength, 2)
					for _, ins := range instances {
						So(ins.VipAddress, ShouldEqual, vipAddress)
					}
					So(instances[0], ShouldNotEqual, instances[1])
				})
			}))
		}))
		time.Sleep(cacheDelay)
	}
}

func TestRegistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
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
				PortEnabled:      true,
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
			PortEnabled:      true,
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

				Convey("Reregister the TESTAPP instance", func() {
					Convey("Instance reregisters correctly", func() {
						err := e.ReregisterInstance(&i)
						So(err, ShouldBeNil)

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
				})
			})
		})
	}
}

func DontTestDeregistration(t *testing.T) {
	e, _ := fargo.NewConnFromConfigFile("./config_sample/local.gcfg")
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
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
	i := fargo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		PortEnabled:      true,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   fargo.DataCenterInfo{Name: fargo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}
	for _, j := range []bool{false, true} {
		e.UseJson = j
		Convey("Read empty instance metadata", t, func() {
			a, err := e.GetApp("EUREKA")
			So(err, ShouldBeNil)
			i := a.Instances[0]
			_, err = i.Metadata.GetString("SomeProp")
			So(err, ShouldBeNil)
		})
		Convey("Register an instance to TESTAPP", t, func() {
			Convey("Instance registers correctly", func() {
				err := e.RegisterInstance(&i)
				So(err, ShouldBeNil)

				Convey("Read valid instance metadata", func() {
					a, err := e.GetApp("TESTAPP")
					So(err, ShouldBeNil)
					So(len(a.Instances), ShouldBeGreaterThan, 0)
					i := a.Instances[0]
					err = e.AddMetadataString(i, "SomeProp", "AValue")
					So(err, ShouldBeNil)
					v, err := i.Metadata.GetString("SomeProp")
					So(err, ShouldBeNil)
					So(v, ShouldEqual, "AValue")
				})
			})
		})
	}
}
