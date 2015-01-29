package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"fmt"
	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"testing"
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
	Convey("Given an InstanceMetadata", t, func() {
		ins := &fargo.Instance{}
		ins.SetMetadataString("key1", "value1")
		ins.SetMetadataString("key2", "value2")

		Convey("When the metadata are marshalled", func() {
			b, err := json.Marshal(&ins.Metadata)
			fmt.Printf("(debug info b = %s)", b)

			Convey("The marshalled JSON should have these values", func() {
				So(string(b), ShouldEqual, `{"key1":"value1","key2":"value2"}`)
				So(err, ShouldBeNil)
			})
		})
	})
}
