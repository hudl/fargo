package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (e *EurekaConnection) generateURL(slugs ...string) string {
	return strings.Join(append([]string{e.SelectServiceURL()}, slugs...), "/")
}

func (e *EurekaConnection) marshal(v interface{}) ([]byte, error) {
	if e.UseJson {
		out, err := json.Marshal(v)
		if err != nil {
			// marshal the JSON *with* indents so it's readable in the error message
			out, _ := json.MarshalIndent(v, "", "    ")
			log.Error("Error marshalling JSON value=%v. Error:\"%s\" JSON body=\"%s\"", v, err.Error(), string(out))
			return nil, err
		}
		return out, nil
	} else {
		out, err := xml.Marshal(v)
		if err != nil {
			// marshal the XML *with* indents so it's readable in the error message
			out, _ := xml.MarshalIndent(v, "", "    ")
			log.Error("Error marshalling XML value=%v. Error:\"%s\" JSON body=\"%s\"", v, err.Error(), string(out))
			return nil, err
		}
		return out, nil
	}
}

// GetApp returns a single eureka application by name
func (e *EurekaConnection) GetApp(name string) (*Application, error) {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	reqURL := e.generateURL(slug)
	log.Debug("Getting app %s from url %s", name, reqURL)
	out, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		log.Error("Couldn't get app %s, error: %s", name, err.Error())
		return nil, err
	}
	if rcode == 404 {
		log.Error("App %s not found (received 404)", name)
		return nil, AppNotFoundError{specific: name}
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}

	var v *Application
	if e.UseJson {
		var r GetAppResponseJson
		err = json.Unmarshal(out, &r)
		v = &r.Application
	} else {
		err = xml.Unmarshal(out, &v)
	}
	if err != nil {
		log.Error("Unmarshalling error: %s", err.Error())
		return nil, err
	}

	v.ParseAllMetadata()
	return v, nil
}

func (e *EurekaConnection) readAppInto(app *Application) error {
	tapp, err := e.GetApp(app.Name)
	if err == nil {
		*app = *tapp
	}
	return err
}

// GetApps returns a map of all Applications
func (e *EurekaConnection) GetApps() (map[string]*Application, error) {
	slug := EurekaURLSlugs["Apps"]
	reqURL := e.generateURL(slug)
	log.Debug("Getting all apps from url %s", reqURL)
	body, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		log.Error("Couldn't get apps, error: %s", err.Error())
		return nil, err
	}
	if rcode > 299 || rcode < 200 {
		log.Warning("Non-200 rcode of %d", rcode)
	}

	var r *GetAppsResponse
	if e.UseJson {
		var rj GetAppsResponseJson
		err = json.Unmarshal(body, &rj)
		r = rj.Response
	} else {
		err = xml.Unmarshal(body, &r)
	}
	if err != nil {
		log.Error("Unmarshalling error: %s", err.Error())
		return nil, err
	}

	apps := map[string]*Application{}
	for i, a := range r.Applications {
		apps[a.Name] = r.Applications[i]
	}
	for name, app := range apps {
		log.Debug("Parsing metadata for app %s", name)
		app.ParseAllMetadata()
	}
	return apps, nil
}

// RegisterInstance will register the given Instance with eureka if it is not already registered,
// but DOES NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)
	log.Debug("Registering instance with url %s", reqURL)
	_, rcode, err := getBody(reqURL+"/"+ins.Id(), e.UseJson)
	if err != nil {
		log.Error("Failed check if Instance=%s exists in app=%s, error: %s",
			ins.Id(), ins.App, err.Error())
		return err
	}
	if rcode == 200 {
		log.Notice("Instance=%s already exists in App=%s, aborting registration", ins.Id(), ins.App)
		return nil
	}
	log.Notice("Instance=%s not yet registered with App=%s, registering.", ins.Id(), ins.App)
	return e.ReregisterInstance(ins)
}

// ReregisterInstance will register the given Instance with eureka but DOES
// NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) ReregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)

	var out []byte
	var err error
	if e.UseJson {
		ins.PortJ.Number = strconv.Itoa(ins.Port)
		ins.SecurePortJ.Number = strconv.Itoa(ins.SecurePort)
		out, err = e.marshal(&RegisterInstanceJson{ins})
	} else {
		out, err = e.marshal(ins)
	}

	body, rcode, err := postBody(reqURL, out, e.UseJson)
	if err != nil {
		log.Error("Could not complete registration, error: %s", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d registering Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure registering instance\n", rcode)
	}

	// read back our registration to pick up eureka-supplied values
	e.readInstanceInto(ins)

	return nil
}

// GetInstance gets an Instance from eureka given its app and instanceid.
func (e *EurekaConnection) GetInstance(app, insId string) (*Instance, error) {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], app, insId)
	reqURL := e.generateURL(slug)
	log.Debug("Getting instance with url %s", reqURL)
	body, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		return nil, err
	}
	if rcode != 200 {
		return nil, fmt.Errorf("Error getting instance, rcode = %d", rcode)
	}
	var ins *Instance
	if e.UseJson {
		var ij RegisterInstanceJson
		err = json.Unmarshal(body, &ij)
		ins = ij.Instance
	} else {
		err = xml.Unmarshal(body, &ins)
	}
	return ins, err
}

func (e *EurekaConnection) readInstanceInto(ins *Instance) error {
	tins, err := e.GetInstance(ins.App, ins.Id())
	if err == nil {
		tins.UniqueID = ins.UniqueID
		*ins = *tins
	}
	return err
}

// DeregisterInstance will deregister the given Instance from eureka. This is good practice
// to do before exiting or otherwise going off line.
func (e *EurekaConnection) DeregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)
	log.Debug("Deregistering instance with url %s", reqURL)

	rcode, err := deleteReq(reqURL)
	if err != nil {
		log.Error("Could not complete deregistration, error: %s", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warning("HTTP returned %d deregistering Instance=%s App=%s", rcode, ins.Id(), ins.App)
		return fmt.Errorf("http returned %d possible failure deregistering instance\n", rcode)
	}

	return nil
}

// AddMetadataString to a given instance. Is immediately sent to Eureka server.
func (e EurekaConnection) AddMetadataString(ins *Instance, key, value string) error {
	slug := fmt.Sprintf("%s/%s/%s/metadata", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)

	params := map[string]string{key: value}

	log.Debug("Updating instance metadata url=%s metadata=%s", reqURL, params)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Error("Could not complete update, error: %s", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warning("HTTP returned %d updating metadata Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure updating instance metadata ", rcode)
	}
	ins.SetMetadataString(key, value)
	return nil
}

// UpdateInstanceStatus updates the status of a given instance with eureka.
func (e EurekaConnection) UpdateInstanceStatus(ins *Instance, status StatusType) error {
	slug := fmt.Sprintf("%s/%s/%s/status", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)

	params := map[string]string{"value": string(status)}

	log.Debug("Updating instance status url=%s value=%s", reqURL, status)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Error("Could not complete update, error: ", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warning("HTTP returned %d updating status Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return fmt.Errorf("http returned %d possible failure updating instance status ", rcode)
	}
	return nil
}

// HeartBeatInstance sends a single eureka heartbeat. Does not continue sending
// heartbeats. Errors if the response is not 200.
func (e *EurekaConnection) HeartBeatInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)
	log.Debug("Sending heartbeat with url %s", reqURL)
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		log.Error("Could not create request for heartbeat, error: %s", err.Error())
		return err
	}
	_, rcode, err := netReq(req)
	if err != nil {
		log.Error("Error sending heartbeat for Instance=%s App=%s, error: %s", ins.Id(), ins.App, err.Error())
		return err
	}
	if rcode != 200 {
		log.Error("Sending heartbeat for Instance=%s App=%s returned code %d", ins.Id(), ins.App, rcode)
		return fmt.Errorf("heartbeat returned code %d\n", rcode)
	}
	return nil
}

func (i *Instance) Id() string {
	if i.UniqueID != nil {
		return i.UniqueID(*i)
	}

	if i.DataCenterInfo.Name == "Amazon" {
		return i.DataCenterInfo.Metadata.InstanceID
	}

	return i.HostName
}
