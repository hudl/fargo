package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestJsonMarshal(t *testing.T) {
	for _, f := range []string{"apps-sample-1-1.json", "apps-sample-1-2.json", "apps-sample-2-2.json"} {
		Convey("Reading .", t, func() {
			blob, err := ioutil.ReadFile("marshal_sample/" + f)

			var v fargo.GetAppsResponseJson
			err = json.Unmarshal(blob, &v)

			if err != nil {
				// Handy dump for debugging funky JSON
				fmt.Printf("v:\n%+v\n", v.Response.Applications)
				for _, app := range v.Response.Applications {
					fmt.Printf("  %+v\n", *app)
					for _, ins := range app.Instances {
						fmt.Printf("    %+v\n", *ins)
					}
				}

				// Print a little more details when there are unmarshalling problems
				switch ute := err.(type) {
				case *json.UnmarshalTypeError:
					fmt.Printf("\nUnmarshalling type error val:%s type:%s: %s\n", ute.Value, ute.Type, err.Error())
					fmt.Printf("UTE:\n%+v\n", ute)
				default:
					fmt.Printf("\nUnmarshalling error: %s\n", err.Error())
				}
			}
			So(err, ShouldBeNil)
		})
	}
}

func TestMetadataMarshal(t *testing.T) {
	Convey("Given an Instance with metadata", t, func() {
		ins := &fargo.Instance{}
		ins.SetMetadataString("key1", "value1")
		ins.SetMetadataString("key2", "value2")

		Convey("When the metadata are marshalled as JSON", func() {
			b, err := json.Marshal(&ins.Metadata)

			Convey("The marshalled JSON should have these values", func() {
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, `{"key1":"value1","key2":"value2"}`)
			})
		})

		Convey("When the metadata are marshalled as XML", func() {
			b, err := xml.Marshal(&ins.Metadata)

			Convey("The marshalled XML should have this value", func() {
				So(err, ShouldBeNil)
				So(string(b), ShouldBeIn,
					"<InstanceMetadata><key1>value1</key1><key2>value2</key2></InstanceMetadata>",
					"<InstanceMetadata><key2>value2</key2><key1>value1</key1></InstanceMetadata>")
			})
		})
	})
}

func TestDataCenterInfoMarshal(t *testing.T) {
	Convey("Given an Instance situated in a data center", t, func() {
		ins := &fargo.Instance{}

		Convey("When the data center name is \"Amazon\"", func() {
			ins.DataCenterInfo.Name = fargo.Amazon
			ins.DataCenterInfo.Metadata.InstanceID = "123"
			ins.DataCenterInfo.Metadata.HostName = "expected.local"

			Convey("When the data center info is marshalled as JSON", func() {
				b, err := json.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled JSON should have these values", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"name":"Amazon","metadata":{"ami-launch-index":"","local-hostname":"","availability-zone":"","instance-id":"123","public-ipv4":"","public-hostname":"","ami-manifest-path":"","local-ipv4":"","hostname":"expected.local","ami-id":"","instance-type":""}}`)

					Convey("The value unmarshalled from JSON should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := json.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						So(d, ShouldResemble, ins.DataCenterInfo)
					})
				})
			})

			Convey("When the data center info is marshalled as XML", func() {
				b, err := xml.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled XML should have this value", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, "<DataCenterInfo><name>Amazon</name><metadata><ami-launch-index></ami-launch-index><local-hostname></local-hostname><availability-zone></availability-zone><instance-id>123</instance-id><public-ipv4></public-ipv4><public-hostname></public-hostname><ami-manifest-path></ami-manifest-path><local-ipv4></local-ipv4><hostname>expected.local</hostname><ami-id></ami-id><instance-type></instance-type></metadata></DataCenterInfo>")

					Convey("The value unmarshalled from XML should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := xml.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						So(d, ShouldResemble, ins.DataCenterInfo)
					})
				})
			})
		})

		Convey("When the data center name is not \"Amazon\"", func() {
			ins.DataCenterInfo.Name = fargo.MyOwn
			ins.DataCenterInfo.AlternateMetadata = map[string]string{
				"instanceId": "123",
				"hostName":   "expected.local",
			}

			Convey("When the data center info is marshalled as JSON", func() {
				b, err := json.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled JSON should have these values", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"name":"MyOwn","metadata":{"hostName":"expected.local","instanceId":"123"}}`)

					Convey("The value unmarshalled from JSON should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := json.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						So(d, ShouldResemble, ins.DataCenterInfo)
					})
				})
			})

			Convey("When the data center info is marshalled as XML", func() {
				b, err := xml.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled XML should have this value", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldBeIn,
						"<DataCenterInfo><name>MyOwn</name><metadata><hostName>expected.local</hostName><instanceId>123</instanceId></metadata></DataCenterInfo>",
						"<DataCenterInfo><name>MyOwn</name><metadata><instanceId>123</instanceId><hostName>expected.local</hostName></metadata></DataCenterInfo>")

					Convey("The value unmarshalled from XML should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := xml.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						So(d, ShouldResemble, ins.DataCenterInfo)
					})
				})
			})
		})
	})
}
