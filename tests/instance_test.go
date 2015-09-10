package fargo_test

// MIT Licensed (see README.md)

import (
	"fmt"
	"testing"

	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInstanceID(t *testing.T) {
	i := fargo.Instance{
		HostName:         "i-6543",
		Port:             9090,
		App:              "TESTAPP",
		IPAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		SecureVipAddress: "127.0.0.10",
		Status:           fargo.UP,
	}

	Convey("Given an instance with DataCenterInfo.Name set to Amazon", t, func() {
		i.DataCenterInfo = fargo.DataCenterInfo{Name: fargo.Amazon}

		Convey("When UniqueID function has NOT been set", func() {
			i.UniqueID = nil

			Convey("And InstanceID has been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = "EXPECTED-ID"

				Convey("Id() should return the provided InstanceID", func() {
					So(i.Id(), ShouldEqual, "EXPECTED-ID")
				})
			})

			Convey("And InstanceID has NOT been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = ""

				Convey("Id() should return an empty string", func() {
					So(i.Id(), ShouldEqual, "")
				})
			})
		})

		Convey("When UniqueID function has been set", func() {
			i.UniqueID = func(i fargo.Instance) string {
				return fmt.Sprintf("%s:%d", i.App, 123)
			}

			Convey("And InstanceID has been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = "UNEXPECTED"

				Convey("Id() should return the ID that is provided by UniqueID", func() {
					So(i.Id(), ShouldEqual, "TESTAPP:123")
				})
			})

			Convey("And InstanceID has not been set in AmazonMetadata", func() {
				i.DataCenterInfo.Metadata.InstanceID = ""

				Convey("Id() should return the ID that is provided by UniqueID", func() {
					So(i.Id(), ShouldEqual, "TESTAPP:123")
				})
			})
		})
	})

	Convey("Given an instance with DataCenterInfo.Name set to MyOwn", t, func() {
		i.DataCenterInfo = fargo.DataCenterInfo{Name: fargo.MyOwn}

		Convey("When UniqueID function has NOT been set", func() {
			i.UniqueID = nil

			Convey("Id() should return the host name", func() {
				So(i.Id(), ShouldEqual, "i-6543")
			})
		})

		Convey("When UniqueID function has been set", func() {
			i.Metadata.Raw = []byte(`{"instanceId": "unique-id"}`)
			i.UniqueID = func(i fargo.Instance) string {
				if id, err := i.Metadata.GetString("instanceId"); err == nil {
					return fmt.Sprintf("%s:%s", i.HostName, id)
				}
				return i.HostName
			}

			Convey("Id() should return the ID that is provided by UniqueID", func() {
				So(i.Id(), ShouldEqual, "i-6543:unique-id")
			})
		})
	})
}
