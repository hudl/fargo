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
