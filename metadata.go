package fargo

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
	"fmt"
	"github.com/clbanning/x2j"
)

func (im *InstanceMetadata) parse() error {
	if im.parsed != nil {
		return nil
	}
	v := map[string]interface{}{}
	// wrap in a BS xml tag so all metadata tags are pulled
	if len(im.Raw) == 0 {
		log.Warning("Metadata contents has length of 0. This may not be correct")
	}
	log.Debug("Metadata length: ", len(im.Raw), " characters")
	j := append(append([]byte("<d>"), im.Raw...), []byte("</d>")...)
	err := x2j.Unmarshal(j, &v)
	if err != nil {
		log.Error("Error unmarshalling: ", err.Error())
		return fmt.Errorf("Error unmarshalling: ", err.Error())
	}
	parsed := v["d"].(map[string]interface{})
	im.parsed = &parsed
	return nil
}

func (im *InstanceMetadata) getItem(key string) (interface{}, error) {
	err := im.parse()
	if err != nil {
		return "", fmt.Errorf("parsing error: ", err.Error())
	}
	return (*im.parsed)[key], nil
}

func (im *InstanceMetadata) GetString(key string) (string, error) {
	v, err := im.getItem(key)
	return v.(string), err
}

func (im *InstanceMetadata) GetInt(key string) (int, error) {
	v, err := im.getItem(key)
	return v.(int), err
}

func (im *InstanceMetadata) GetFloat32(key string) (float32, error) {
	v, err := im.getItem(key)
	return v.(float32), err
}

func (im *InstanceMetadata) GetFloat64(key string) (float64, error) {
	v, err := im.getItem(key)
	return v.(float64), err
}

func (im *InstanceMetadata) GetBool(key string) (bool, error) {
	v, err := im.getItem(key)
	return v.(bool), err
}
