package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"strconv"
)

// Temporary structs used for GetAppsResponse unmarshalling
type getAppsResponseArray GetAppsResponse
type getAppsResponseSingle struct {
	Application   *Application `json:"application"`
	AppsHashcode  string       `json:"apps__hashcode"`
	VersionsDelta int          `json:"versions__delta"`
}

// UnmarshalJSON is a custom JSON unmarshaler for GetAppsResponse to deal with
// sometimes non-wrapped Application arrays when there is only a single Application item.
func (r *GetAppsResponse) UnmarshalJSON(b []byte) error {
	marshalLog.Debug("GetAppsResponse.UnmarshalJSON b:%s\n", string(b))
	var err error

	// Normal array case
	var ra getAppsResponseArray
	if err = json.Unmarshal(b, &ra); err == nil {
		marshalLog.Debug("GetAppsResponse.UnmarshalJSON ra:%+v\n", ra)
		*r = GetAppsResponse(ra)
		return nil
	}
	// Bogus non-wrapped case
	var rs getAppsResponseSingle
	if err = json.Unmarshal(b, &rs); err == nil {
		marshalLog.Debug("GetAppsResponse.UnmarshalJSON rs:%+v\n", rs)
		r.Applications = make([]*Application, 1, 1)
		r.Applications[0] = rs.Application
		r.AppsHashcode = rs.AppsHashcode
		r.VersionsDelta = rs.VersionsDelta
		return nil
	}
	return err
}

// Temporary structs used for Application unmarshalling
type applicationArray Application
type applicationSingle struct {
	Name     string    `json:"name"`
	Instance *Instance `json:"instance"`
}

// UnmarshalJSON is a custom JSON unmarshaler for Application to deal with
// sometimes non-wrapped Instance array when there is only a single Instance item.
func (a *Application) UnmarshalJSON(b []byte) error {
	marshalLog.Debug("Application.UnmarshalJSON b:%s\n", string(b))
	var err error

	// Normal array case
	var aa applicationArray
	if err = json.Unmarshal(b, &aa); err == nil {
		marshalLog.Debug("Application.UnmarshalJSON aa:%+v\n", aa)
		*a = Application(aa)
		return nil
	}

	// Bogus non-wrapped case
	var as applicationSingle
	if err = json.Unmarshal(b, &as); err == nil {
		marshalLog.Debug("Application.UnmarshalJSON as:%+v\n", as)
		a.Name = as.Name
		a.Instances = make([]*Instance, 1, 1)
		a.Instances[0] = as.Instance
		return nil
	}
	return err
}

type instance Instance

// UnmarshalJSON is a custom JSON unmarshaler for Instance to deal with the
// different Port encodings between XML and JSON. Here we copy the values from the JSON
// Port struct into the simple XML int field.
func (i *Instance) UnmarshalJSON(b []byte) error {
	var err error
	var ii instance
	if err = json.Unmarshal(b, &ii); err == nil {
		marshalLog.Debug("Instance.UnmarshalJSON ii:%+v\n", ii)
		*i = Instance(ii)
		i.Port = parsePort(ii.PortJ.Number)
		i.SecurePort = parsePort(ii.SecurePortJ.Number)
		return nil
	}
	return err
}

func parsePort(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Warning("Invalid port error: %s", err.Error())
	}
	return n
}

// UnmarshalJSON is a custom JSON unmarshaler for InstanceMetadata to handle squirreling away
// the raw JSON for later parsing.
func (i *InstanceMetadata) UnmarshalJSON(b []byte) error {
	i.Raw = b
	// TODO(cq) could actually parse Raw here, and in a parallel UnmarshalXML as well.
	return nil
}

// MarshalJSON is a custom JSON marshaler for InstanceMetadata.
func (i *InstanceMetadata) MarshalJSON() ([]byte, error) {
	if i.parsed != nil {
		return json.Marshal(i.parsed)
	}

	if i.Raw == nil {
		i.Raw = []byte("{}")
	}

	return i.Raw, nil
}
