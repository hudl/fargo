package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"bytes"
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

func portsEqual(actual, expected *fargo.Instance) {
	Convey("For the insecure port", func() {
		So(actual.Port, ShouldEqual, expected.Port)
		So(actual.PortEnabled, ShouldEqual, expected.PortEnabled)

		Convey("For the secure port", func() {
			So(actual.SecurePort, ShouldEqual, expected.SecurePort)
			So(actual.SecurePortEnabled, ShouldEqual, expected.SecurePortEnabled)
		})
	})
}

func jsonEncodedInstanceHasPortsEqualTo(b []byte, expected *fargo.Instance) {
	Convey("Reading them back should yield the equivalent value", func() {
		var decoded fargo.Instance
		err := json.Unmarshal(b, &decoded)
		So(err, ShouldBeNil)
		portsEqual(&decoded, expected)
	})
}

func xmlEncodedInstanceHasPortsEqualTo(b []byte, expected *fargo.Instance) {
	Convey("And reading them back should yield the equivalent value", func() {
		var decoded fargo.Instance
		err := xml.Unmarshal(b, &decoded)
		So(err, ShouldBeNil)
		portsEqual(&decoded, expected)
	})
}

func TestPortsMarshal(t *testing.T) {
	Convey("Given an Instance with only the insecure port enabled", t, func() {
		ins := fargo.Instance{
			Port:        80,
			PortEnabled: true,
		}

		Convey("When the ports are marshalled as JSON", func() {
			b, err := json.Marshal(&ins)

			Convey("The marshalled JSON should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `,"port":{"$":"80","@enabled":"true"}`)
				So(s, ShouldContainSubstring, `,"securePort":{"$":"0","@enabled":"false"}`)

				jsonEncodedInstanceHasPortsEqualTo(b, &ins)

				Convey("When the Eureka server is version 1.22 or later", func() {
					jsonEncodedInstanceHasPortsEqualTo(bytes.Replace(b, []byte(`"80"`), []byte("80"), -1), &ins)
				})
			})
		})

		Convey("When the ports are marshalled as XML", func() {
			b, err := xml.Marshal(&ins)

			Convey("The marshalled XML should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `<port enabled="true">80</port>`)
				So(s, ShouldContainSubstring, `<securePort enabled="false">0</securePort>`)

				xmlEncodedInstanceHasPortsEqualTo(b, &ins)
			})
		})
	})
	Convey("Given an Instance with only the secure port enabled", t, func() {
		ins := fargo.Instance{
			SecurePort:        443,
			SecurePortEnabled: true,
		}

		Convey("When the ports are marshalled as JSON", func() {
			b, err := json.Marshal(&ins)

			Convey("The marshalled JSON should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `,"port":{"$":"0","@enabled":"false"}`)
				So(s, ShouldContainSubstring, `,"securePort":{"$":"443","@enabled":"true"}`)

				jsonEncodedInstanceHasPortsEqualTo(b, &ins)

				Convey("When the Eureka server is version 1.22 or later", func() {
					jsonEncodedInstanceHasPortsEqualTo(bytes.Replace(b, []byte(`"443"`), []byte("443"), -1), &ins)
				})
			})
		})

		Convey("When the ports are marshalled as XML", func() {
			b, err := xml.Marshal(&ins)

			Convey("The marshalled XML should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `<port enabled="false">0</port>`)
				So(s, ShouldContainSubstring, `<securePort enabled="true">443</securePort>`)

				xmlEncodedInstanceHasPortsEqualTo(b, &ins)
			})
		})
	})
	Convey("Given an Instance with only the both ports enabled", t, func() {
		ins := fargo.Instance{
			Port:              80,
			PortEnabled:       true,
			SecurePort:        443,
			SecurePortEnabled: true,
		}

		Convey("When the ports are marshalled as JSON", func() {
			b, err := json.Marshal(&ins)

			Convey("The marshalled JSON should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `,"port":{"$":"80","@enabled":"true"}`)
				So(s, ShouldContainSubstring, `,"securePort":{"$":"443","@enabled":"true"}`)

				jsonEncodedInstanceHasPortsEqualTo(b, &ins)

				Convey("When the Eureka server is version 1.22 or later", func() {
					b = bytes.Replace(b, []byte(`"80"`), []byte("80"), -1)
					b = bytes.Replace(b, []byte(`"443"`), []byte("443"), -1)
					jsonEncodedInstanceHasPortsEqualTo(b, &ins)
				})
			})
		})

		Convey("When the ports are marshalled as XML", func() {
			b, err := xml.Marshal(&ins)

			Convey("The marshalled XML should have these values", func() {
				So(err, ShouldBeNil)
				s := string(b)
				So(s, ShouldContainSubstring, `<port enabled="true">80</port>`)
				So(s, ShouldContainSubstring, `<securePort enabled="true">443</securePort>`)

				xmlEncodedInstanceHasPortsEqualTo(b, &ins)
			})
		})
	})
}

func TestMetadataMarshal(t *testing.T) {
	Convey("Given an Instance with metadata", t, func() {
		ins := fargo.Instance{}
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
		ins := fargo.Instance{}

		Convey("When the data center name is \"Amazon\"", func() {
			ins.DataCenterInfo.Name = fargo.Amazon
			ins.DataCenterInfo.Class = "ignored"
			ins.DataCenterInfo.Metadata.InstanceID = "123"
			ins.DataCenterInfo.Metadata.HostName = "expected.local"

			Convey("When the data center info is marshalled as JSON", func() {
				b, err := json.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled JSON should have these values", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"name":"Amazon","@class":"com.netflix.appinfo.AmazonInfo","metadata":{"ami-launch-index":"","local-hostname":"","availability-zone":"","instance-id":"123","public-ipv4":"","public-hostname":"","ami-manifest-path":"","local-ipv4":"","hostname":"expected.local","ami-id":"","instance-type":""}}`)

					Convey("The value unmarshalled from JSON should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := json.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						expected := ins.DataCenterInfo
						expected.Class = "com.netflix.appinfo.AmazonInfo"
						So(d, ShouldResemble, expected)
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
						expected := ins.DataCenterInfo
						expected.Class = ""
						So(d, ShouldResemble, expected)
					})
				})
			})
		})

		Convey("When the data center name is not \"Amazon\"", func() {
			ins.DataCenterInfo.Name = fargo.MyOwn
			ins.DataCenterInfo.Class = "ignored"
			ins.DataCenterInfo.AlternateMetadata = map[string]string{
				"instanceId": "123",
				"hostName":   "expected.local",
			}

			Convey("When the data center info has no class specified and is marshalled as JSON", func() {
				b, err := json.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled JSON should have these values", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"name":"MyOwn","@class":"com.netflix.appinfo.MyDataCenterInfo","metadata":{"hostName":"expected.local","instanceId":"123"}}`)

					Convey("The value unmarshalled from JSON should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := json.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						expected := ins.DataCenterInfo
						expected.Class = "com.netflix.appinfo.MyDataCenterInfo"
						So(d, ShouldResemble, expected)
					})
				})
			})

			Convey("When the data center info has both a custom name and class specified and is marshalled as JSON", func() {
				ins.DataCenterInfo.Name = "Custom"
				ins.DataCenterInfo.Class = "custom"
				b, err := json.Marshal(&ins.DataCenterInfo)

				Convey("The marshalled JSON should have these values", func() {
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, `{"name":"Custom","@class":"custom","metadata":{"hostName":"expected.local","instanceId":"123"}}`)

					Convey("The value unmarshalled from JSON should have the same values as the original", func() {
						d := fargo.DataCenterInfo{}
						err := json.Unmarshal(b, &d)

						So(err, ShouldBeNil)
						So(d, ShouldResemble, ins.DataCenterInfo)

						Convey("Even if the server translates strings to other types", func() {
							translated := bytes.Replace(b, []byte(`"123"`), []byte("123"), 1)

							d := fargo.DataCenterInfo{}
							err := json.Unmarshal(translated, &d)

							So(err, ShouldBeNil)
							So(d, ShouldResemble, ins.DataCenterInfo)
						})
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
						expected := ins.DataCenterInfo
						expected.Class = ""
						So(d, ShouldResemble, expected)
					})
				})
			})
		})
	})
}
