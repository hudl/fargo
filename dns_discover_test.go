package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestGetNXDomain(t *testing.T) {
	Convey("Given nonexistent domain nxd.local.", t, func() {
		resp, _, err := findTXT("nxd.local.")
		So(err, ShouldNotBeNil)
		So(len(resp), ShouldEqual, 0)
	})
}

func TestGetNetflixTestDomain(t *testing.T) {
	Convey("Given domain txt.us-east-1.discoverytest.netflix.net.", t, func() {
		// TODO: use a mock DNS server to eliminate dependency on netflix
		// keeping their discoverytest domain up
		resp, ttl, err := findTXT("txt.us-east-1.discoverytest.netflix.net.")
		So(err, ShouldBeNil)
		So(ttl, ShouldEqual, 60*time.Second)
		So(len(resp), ShouldEqual, 3)
		Convey("And the contents are zone records", func() {
			expected := map[string]bool{
				"us-east-1c.us-east-1.discoverytest.netflix.net": true,
				"us-east-1d.us-east-1.discoverytest.netflix.net": true,
				"us-east-1e.us-east-1.discoverytest.netflix.net": true,
			}
			for _, item := range resp {
				_, ok := expected[item]
				So(ok, ShouldEqual, true)
			}
			Convey("And the zone records contain instances", func() {
				for _, record := range resp {
					servers, _, err := findTXT("txt." + record + ".")
					So(err, ShouldBeNil)
					So(len(servers) >= 1, ShouldEqual, true)
					// servers should be EC2 DNS names
					So(servers[0][0:4], ShouldEqual, "ec2-")
				}
			})
		})
	})
	Convey("Autodiscover discoverytest.netflix.net.", t, func() {
		servers, ttl, err := discoverDNS("discoverytest.netflix.net", 7001, "")
		So(ttl, ShouldEqual, 60*time.Second)
		So(err, ShouldBeNil)
		So(len(servers), ShouldEqual, 6)
		Convey("Servers discovered should all be EC2 DNS names", func() {
			for _, s := range servers {
				So(s[0:11], ShouldEqual, "http://ec2-")
			}
		})
	})
}
