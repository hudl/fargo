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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func postXml(url string, reqBody []byte) ([]byte, int, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		log.Error("Could not create POST %s with body %s Error: %s", url, string(reqBody), err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXml(req)
	if err != nil {
		log.Error("Could not complete POST %s with body %s Error: %s", url, string(reqBody), err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func getXml(url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Could not create POST %s with Error: %s", url, err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXml(req)
	if err != nil {
		log.Error("Could not complete POST %s with Error: %s", url, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func reqXml(req *http.Request) ([]byte, int, error) {

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")

	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failure reading request body Error: %s", err.Error())
		return nil, -1, err
	}
	// At this point we're done and shit worked, simply return the bytes
	return body, resp.StatusCode, nil
}

func (e *EurekaConnection) GetApp(name string) (Application, error) {
	url := fmt.Sprintf("%s/%s/%s", e.SelectServiceUrl(), EurekaUrlSlugs["Apps"], name)
	log.Debug("Getting app %s from url %s", name, url)
	out, rcode, err := getXml(url)
	if err != nil {
		log.Error("Couldn't get XML. Error: %s", err.Error())
		return Application{}, err
	}
	var v Application
	err = xml.Unmarshal(out, &v)
	if err != nil {
		log.Error("Unmarshalling error. Error: %s", err.Error())
		return Application{}, err
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}
	return v, nil
}

func (e *EurekaConnection) GetApps() (map[string]Application, error) {
	url := fmt.Sprintf("%s/%s", e.SelectServiceUrl(), EurekaUrlSlugs["Apps"])
	log.Debug("Getting all apps from url %s", url)
	out, rcode, err := getXml(url)
	if err != nil {
		log.Error("Couldn't get XML.", err.Error())
		return nil, err
	}
	var v GetAppsResponse
	err = xml.Unmarshal(out, &v)
	if err != nil {
		log.Error("Unmarshalling error", err.Error())
		return nil, err
	}
	apps := map[string]Application{}
	for _, app := range v.Applications {
		apps[app.Name] = app
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}
	return apps, nil
}

func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	url := fmt.Sprintf("%s/%s/%s", e.SelectServiceUrl(), EurekaUrlSlugs["Apps"], ins.App)
	log.Debug("Registering instance with url %s", url)
	_, rcode, err := getXml(url + "/" + ins.HostName)
	if err != nil {
		log.Error("Failed check if Instance=%s exists in App=%s. Error=\"%s\"",
			ins.HostName, ins.App, err.Error())
		return err
	}
	if rcode == 200 {
		log.Notice("Instance=%s already exists in App=%s aborting registration", ins.HostName, ins.App)
		return nil
	} else {
		log.Notice("Instance=%s not yet registered with App=%s. Registering.", ins.HostName, ins.App)
	}

	out, err := xml.Marshal(ins)
	if err != nil {
		// marshal the xml *with* indents so it's readable if there's an error
		out, _ := xml.MarshalIndent(ins, "", "    ")
		log.Error("Error marshalling XML Instance=%s App=%s. Error:\"%s\" XML body=\"%s\"", err.Error(), ins.HostName, ins.App, string(out))
		return err
	}
	body, rcode, err := postXml(url, out)
	if err != nil {
		log.Error("Could not complete registration Error: ", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d registering Instance=%s App=%s Body=\"%s\"", rcode, ins.HostName, ins.App, string(body))
		return errors.New(fmt.Sprintf("HTTP returned %d possible failure creating instance\n", rcode))
	}

	body, rcode, err = getXml(url + "/" + ins.HostName)
	xml.Unmarshal(body, ins)
	return nil
}

func (e *EurekaConnection) HeartBeatInstance(ins *Instance) error {
	url := fmt.Sprintf("%s/%s/%s/%s", e.SelectServiceUrl(), EurekaUrlSlugs["Apps"], ins.App, ins.HostName)
	log.Debug("Sending heartbeat with url %s", url)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Error("Could not create request for heartbeat. Error: %s", err.Error())
		return err
	}
	_, rcode, err := reqXml(req)
	if err != nil {
		log.Error("Error sending heartbeat for Instance=%s App=%s error: %s", ins.HostName, ins.App, err.Error())
		return err
	}
	if rcode != 200 {
		log.Error("Sending heartbeat for Instance=%s App=%s returned code %d\n", ins.HostName, ins.App, rcode)
		return errors.New(fmt.Sprintf("Error, heartbeat returned code %d\n", rcode))
	}
	return nil
}
