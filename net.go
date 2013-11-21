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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/pmylund/go-cache"
	"net/http"
	"strings"
	"time"
)

// expire cached items after 30 seconds, cleanup every 10
var eurekaCache = cache.New(30*time.Second, 10*time.Second)

func (e *EurekaConnection) generateUrl(slugs ...string) string {
	return strings.Join(append([]string{e.SelectServiceURL()}, slugs...), "/")
}

// GetMetaData fills in the "MetadataMap" field on a given Instance This is
// here because the golang XML unmarshalling doesn't handle arbitrary XML well.
func (e *EurekaConnection) GetMetadata(i *Instance) error {
	var slug string
	if i.DataCenterInfo.Name == Amazon {
		slug = fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], i.App, i.DataCenterInfo.Metadata.InstanceID)
	} else if i.DataCenterInfo.Name == MyOwn {
		slug = fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], i.App, i.HostName)
	}
	url := e.generateUrl(slug)
	out, rcode, err := getJSON(url)
	if err != nil {
		log.Error("Couldn't get JSON. Error: %s", err.Error())
		return err
	}
	if rcode == 404 {
		log.Error("instance %s/%s not found (received 404)", i.App, i.HostName)
		return fmt.Errorf("instance %s/%s not found (received 404)", i.App, i.HostName)
	}
	v := map[string]interface{}{}
	err = json.Unmarshal(out, &v)
	if err != nil {
		log.Error("Couldn't decode JSON. Error: %s", err.Error())
		return err
	}
	i.MetadataMap = v["instance"].(map[string]interface{})["metadata"].(map[string]interface{})
	return nil
}

// GetApp returns a single eureka application by name. This may be cached.
func (e *EurekaConnection) GetApp(name string) (Application, error) {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	url := e.generateUrl(slug)
	cachedApp, found := eurekaCache.Get(slug)
	if found {
		log.Notice("Got %s from cache", url)
		return cachedApp.(Application), nil
	}
	log.Debug("Getting app %s from url %s", name, url)
	out, rcode, err := getXML(url)
	if err != nil {
		log.Error("Couldn't get XML. Error: %s", err.Error())
		return Application{}, err
	}
	if rcode == 404 {
		log.Error("application %s not found (received 404)", name)
		return Application{}, fmt.Errorf("application %s not found (received 404)", name)
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
	eurekaCache.Set(slug, v, 0)
	return v, nil
}

// GetApps returns a map of all Applications. Note: May be cached
func (e *EurekaConnection) GetApps() (map[string]Application, error) {
	slug := EurekaURLSlugs["Apps"]
	url := e.generateUrl(slug)
	cachedApps, found := eurekaCache.Get(slug)
	if found {
		log.Notice("Got %s from cache", url)
		return cachedApps.(map[string]Application), nil
	}
	log.Debug("Getting all apps from url %s", url)
	out, rcode, err := getXML(url)
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
	eurekaCache.Set(slug, apps, 0)
	return apps, nil
}

// RegisterInstance will register the relevant Instance with eureka but DOES
// NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	url := e.generateUrl(slug)
	log.Debug("Registering instance with url %s", url)
	_, rcode, err := getXML(url + "/" + ins.HostName)
	if err != nil {
		log.Error("Failed check if Instance=%s exists in App=%s. Error=\"%s\"",
			ins.HostName, ins.App, err.Error())
		return err
	}
	if rcode == 200 {
		log.Notice("Instance=%s already exists in App=%s aborting registration", ins.HostName, ins.App)
		return nil
	}
	log.Notice("Instance=%s not yet registered with App=%s. Registering.", ins.HostName, ins.App)

	out, err := xml.Marshal(ins)
	if err != nil {
		// marshal the xml *with* indents so it's readable if there's an error
		out, _ := xml.MarshalIndent(ins, "", "    ")
		log.Error("Error marshalling XML Instance=%s App=%s. Error:\"%s\" XML body=\"%s\"", err.Error(), ins.HostName, ins.App, string(out))
		return err
	}
	body, rcode, err := postXML(url, out)
	if err != nil {
		log.Error("Could not complete registration Error: ", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d registering Instance=%s App=%s Body=\"%s\"", rcode, ins.HostName, ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure creating instance\n", rcode)
	}

	body, rcode, err = getXML(url + "/" + ins.HostName)
	xml.Unmarshal(body, ins)
	return nil
}

// HeartBeatInstance sends a single eureka heartbeat. Does not continue sending
// heartbeats. Errors if the response is not 200.
func (e *EurekaConnection) HeartBeatInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.HostName)
	url := e.generateUrl(slug)
	log.Debug("Sending heartbeat with url %s", url)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Error("Could not create request for heartbeat. Error: %s", err.Error())
		return err
	}
	_, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Error sending heartbeat for Instance=%s App=%s error: %s", ins.HostName, ins.App, err.Error())
		return err
	}
	if rcode != 200 {
		log.Error("Sending heartbeat for Instance=%s App=%s returned code %d\n", ins.HostName, ins.App, rcode)
		return fmt.Errorf("heartbeat returned code %d\n", rcode)
	}
	return nil
}

func (e *EurekaConnection) readAppInto(name string, app *Application) error {
	//TODO: This should probably use the cache, but it's only called at PollInterval anyways
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	url := e.generateUrl(slug)
	log.Debug("Getting app %s from url %s", name, url)
	out, rcode, err := getXML(url)
	if err != nil {
		log.Error("Couldn't get XML. Error: %s", err.Error())
		return err
	}
	oldInstances := app.Instances
	app.Instances = []Instance{}
	err = xml.Unmarshal(out, app)
	if err != nil {
		app.Instances = oldInstances
		log.Error("Unmarshalling error. Error: %s", err.Error())
		return err
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}
	return nil
}
