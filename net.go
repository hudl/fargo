package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

func (e *EurekaConnection) generateURL(slugs ...string) string {
	return strings.Join(append([]string{e.SelectServiceURL()}, slugs...), "/")
}

// GetApp returns a single eureka application by name
func (e *EurekaConnection) GetApp(name string) (Application, error) {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	reqURL := e.generateURL(slug)
	log.Debug("Getting app %s from url %s", name, reqURL)
	out, rcode, err := getXML(reqURL)
	if err != nil {
		log.Error("Couldn't get XML. Error: %s", err.Error())
		return Application{}, err
	}
	if rcode == 404 {
		log.Error("application %s not found (received 404)", name)
		return Application{}, AppNotFoundError{specific: name}
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
	v.ParseAllMetadata()
	return v, nil
}

// GetApps returns a map of all Applications
func (e *EurekaConnection) GetApps() (map[string]*Application, error) {
	slug := EurekaURLSlugs["Apps"]
	reqURL := e.generateURL(slug)
	log.Debug("Getting all apps from url %s", reqURL)
	out, rcode, err := getXML(reqURL)
	if err != nil {
		log.Error("Couldn't get XML: " + err.Error())
		return nil, err
	}
	var v GetAppsResponse
	err = xml.Unmarshal(out, &v)
	if err != nil {
		log.Error("Unmarshalling error: " + err.Error())
		return nil, err
	}
	apps := map[string]*Application{}
	for i, a := range v.Applications {
		apps[a.Name] = &v.Applications[i]
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}
	for name, app := range apps {
		log.Debug("Parsing metadata for Application=%s", name)
		app.ParseAllMetadata()
	}
	return apps, nil
}

// AddMetadataString to a given instance. Is immediately sent to Eureka server.
func (e EurekaConnection) AddMetadataString(ins *Instance, key, value string) error {
	slug := fmt.Sprintf("%s/%s/%s/metadata", EurekaURLSlugs["Apps"], ins.App, ins.HostName)
	reqURL := e.generateURL(slug)

	params := map[string]string{key: value}
	if ins.Metadata.parsed == nil {
		ins.Metadata.parsed = map[string]interface{}{}
	}
	ins.Metadata.parsed[key] = value

	log.Debug("Updating instance metadata url=%s metadata=%s", reqURL, params)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Error("Could not complete update with Error: ", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warning("HTTP returned %d updating metadata Instance=%s App=%s Body=\"%s\"", rcode, ins.HostName, ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure updating instance metadata ", rcode)
	}
	return nil
}

// RegisterInstance will register the given Instance with eureka if it is not already registered,
// but DOES NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)
	log.Debug("Registering instance with url %s", reqURL)
	_, rcode, err := getXML(reqURL + "/" + ins.HostName)
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
	return e.ReregisterInstance(ins)
}

// ReregisterInstance will register the given Instance with eureka but DOES
// NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) ReregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)
	out, err := xml.Marshal(ins)
	if err != nil {
		// marshal the xml *with* indents so it's readable if there's an error
		out, _ := xml.MarshalIndent(ins, "", "    ")
		log.Error("Error marshalling XML Instance=%s App=%s. Error:\"%s\" XML body=\"%s\"", err.Error(), ins.HostName, ins.App, string(out))
		return err
	}
	body, rcode, err := postXML(reqURL, out)
	if err != nil {
		log.Error("Could not complete registration Error: ", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d registering Instance=%s App=%s Body=\"%s\"", rcode, ins.HostName, ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure registering instance\n", rcode)
	}

	body, rcode, err = getXML(reqURL + "/" + ins.HostName)
	xml.Unmarshal(body, ins)
	return nil
}

// DeregisterInstance will register the given Instance with eureka but DOES
// NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) DeregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.HostName)
	reqURL := e.generateURL(slug)
	log.Debug("Deregistering instance with url %s", reqURL)

	rcode, err := deleteReq(reqURL)
	if err != nil {
		log.Error("Could not complete deregistration Error: ", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d deregistering Instance=%s App=%s", rcode, ins.HostName, ins.App)
		return fmt.Errorf("http returned %d possible failure deregistering instance\n", rcode)
	}

	return nil
}

// UpdateInstanceStatus updates the status of a given instance with eureka.
func (e EurekaConnection) UpdateInstanceStatus(ins *Instance, status StatusType) error {
	slug := fmt.Sprintf("%s/%s/%s/status", EurekaURLSlugs["Apps"], ins.App, ins.HostName)
	reqURL := e.generateURL(slug)

	params := map[string]string{"value": string(status)}

	log.Debug("Updating instance status url=%s value=%s", reqURL, status)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Error("Could not complete update with Error: ", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warning("HTTP returned %d updating status Instance=%s App=%s Body=\"%s\"", rcode, ins.HostName, ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure updating instance status ", rcode)
	}
	return nil
}

// HeartBeatInstance sends a single eureka heartbeat. Does not continue sending
// heartbeats. Errors if the response is not 200.
func (e *EurekaConnection) HeartBeatInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.HostName)
	reqURL := e.generateURL(slug)
	log.Debug("Sending heartbeat with url %s", reqURL)
	req, err := http.NewRequest("PUT", reqURL, nil)
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
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	reqURL := e.generateURL(slug)
	log.Debug("Getting app %s from url %s", name, reqURL)
	out, rcode, err := getXML(reqURL)
	if err != nil {
		log.Error("Couldn't get XML. Error: %s", err.Error())
		return err
	}
	oldInstances := app.Instances
	app.Instances = []*Instance{}
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
