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
