package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/xml"
	"github.com/hudl/fargo"
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
)

func TestGetInt(t *testing.T) {
	Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		Convey("With metadata", func() {
			metadata := new(fargo.InstanceMetadata)
			instance.Metadata = *metadata
			Convey("That has a single integer value", func() {
				key := "d"
				value := 1
				metadata.Raw = []byte("<" + key + ">" + strconv.Itoa(value) + "</" + key + ">")
				Convey("GetInt should return that value", func() {
					actualValue, err := metadata.GetInt(key)
					So(err, ShouldBeNil)
					So(actualValue, ShouldEqual, value)
				})
			})
		})
	})
}

func TestGetFloat(t *testing.T) {
	Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		Convey("With metadata", func() {
			metadata := new(fargo.InstanceMetadata)
			instance.Metadata = *metadata
			Convey("That has a float64 value", func() {
				key := "d"
				value := 1.9
				metadata.Raw = []byte("<" + key + ">" + strconv.FormatFloat(value, 'f', -1, 64) + "</" + key + ">")
				Convey("GetFloat64 should return that value", func() {
					actualValue, err := metadata.GetFloat64(key)
					So(err, ShouldBeNil)
					So(actualValue, ShouldEqual, value)
				})
			})
			Convey("That has a float32 value", func() {
				key := "d"
				value := 1.9
				metadata.Raw = []byte("<" + key + ">" + strconv.FormatFloat(value, 'f', -1, 32) + "</" + key + ">")
				Convey("GetFloat32 should return that value", func() {
					actualValue, err := metadata.GetFloat32(key)
					So(err, ShouldBeNil)
					So(actualValue, ShouldEqual, float32(1.9))
				})
			})
		})
	})
}

func TestSerializeMeta(t *testing.T) {
	Convey("Given an instance", t, func() {
		instance := new(fargo.Instance)
		Convey("With metadata", func() {
			instance.SetMetadataString("test", "value")
			Convey("Serializing results in correct JSON", func() {
				b, err := instance.Metadata.MarshalJSON()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "{\"test\":\"value\"}")
			})
			Convey("Serializing results in correct XML", func() {
				b, err := xml.Marshal(instance.Metadata)
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "<InstanceMetadata><test>value</test></InstanceMetadata>")
			})
			Convey("Blank metadata results in blank XML", func() {
				metadata := new(fargo.InstanceMetadata)
				b, err := xml.Marshal(metadata)
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "<InstanceMetadata></InstanceMetadata>")
			})
		})
	})
}
