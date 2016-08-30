package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"encoding/xml"
	"io"
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
	marshalLog.Debugf("GetAppsResponse.UnmarshalJSON b:%s\n", string(b))
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
	marshalLog.Debugf("Application.UnmarshalJSON b:%s\n", string(b))
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
		log.Warningf("Invalid port error: %s", err.Error())
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

// startLocalName creates a start-tag of an XML element with the given local name and no namespace name.
func startLocalName(local string) xml.StartElement {
	return xml.StartElement{Name: xml.Name{Space: "", Local: local}}
}

// MarshalXML is a custom XML marshaler for InstanceMetadata.
func (i InstanceMetadata) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}

	if i.parsed != nil {
		for key, value := range i.parsed {
			t := startLocalName(key)
			tokens = append(tokens, t, xml.CharData(value.(string)), xml.EndElement{Name: t.Name})
		}
	}
	tokens = append(tokens, xml.EndElement{Name: start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	return e.Flush()
}

type metadataMap map[string]string

// MarshalXML is a custom XML marshaler for metadataMap, mapping each metadata name/value pair to a
// correspondingly named XML element with the pair's value as character data content.
func (m metadataMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for k, v := range m {
		if err := e.EncodeElement(v, startLocalName(k)); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

// UnmarshalXML is a custom XML unmarshaler for metadataMap, mapping each XML element's name and
// character data content to a corresponding metadata name/value pair.
func (m metadataMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	for {
		t, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if k, ok := t.(xml.StartElement); ok {
			if err := d.DecodeElement(&v, &k); err != nil {
				return err
			}
			m[k.Name.Local] = v
		}
	}
	return nil
}

func metadataValue(i DataCenterInfo) interface{} {
	if i.Name == Amazon {
		return i.Metadata
	}
	return metadataMap(i.AlternateMetadata)
}

var (
	startName     = startLocalName("name")
	startMetadata = startLocalName("metadata")
)

// MarshalXML is a custom XML marshaler for DataCenterInfo, writing either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i DataCenterInfo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if err := e.EncodeElement(i.Name, startName); err != nil {
		return err
	}
	if err := e.EncodeElement(metadataValue(i), startMetadata); err != nil {
		return err
	}

	return e.EncodeToken(start.End())
}

type preliminaryDataCenterInfo struct {
	Name     string      `xml:"name" json:"name"`
	Metadata metadataMap `xml:"metadata" json:"metadata"`
}

func bindValue(dst *string, src map[string]string, k string) bool {
	if v, ok := src[k]; ok {
		*dst = v
		return true
	}
	return false
}

func populateAmazonMetadata(dst *AmazonMetadataType, src map[string]string) {
	bindValue(&dst.AmiLaunchIndex, src, "ami-launch-index")
	bindValue(&dst.LocalHostname, src, "local-hostname")
	bindValue(&dst.AvailabilityZone, src, "availability-zone")
	bindValue(&dst.InstanceID, src, "instance-id")
	bindValue(&dst.PublicIpv4, src, "public-ipv4")
	bindValue(&dst.PublicHostname, src, "public-hostname")
	bindValue(&dst.AmiManifestPath, src, "ami-manifest-path")
	bindValue(&dst.LocalIpv4, src, "local-ipv4")
	bindValue(&dst.HostName, src, "hostname")
	bindValue(&dst.AmiID, src, "ami-id")
	bindValue(&dst.InstanceType, src, "instance-type")
}

func adaptDataCenterInfo(dst *DataCenterInfo, src preliminaryDataCenterInfo) {
	dst.Name = src.Name
	if src.Name == Amazon {
		populateAmazonMetadata(&dst.Metadata, src.Metadata)
	} else {
		dst.AlternateMetadata = src.Metadata
	}
}

// UnmarshalXML is a custom XML unmarshaler for DataCenterInfo, populating either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i *DataCenterInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p := preliminaryDataCenterInfo{
		Metadata: make(map[string]string, 11),
	}
	if err := d.DecodeElement(&p, &start); err != nil {
		return err
	}
	adaptDataCenterInfo(i, p)
	return nil
}

// MarshalJSON is a custom JSON marshaler for DataCenterInfo, writing either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i DataCenterInfo) MarshalJSON() ([]byte, error) {
	type named struct {
		Name string `json:"name"`
	}
	if i.Name == Amazon {
		return json.Marshal(struct {
			named
			Metadata AmazonMetadataType `json:"metadata"`
		}{named{i.Name}, i.Metadata})
	}
	return json.Marshal(struct {
		named
		Metadata map[string]string `json:"metadata"`
	}{named{i.Name}, i.AlternateMetadata})
}

// UnmarshalJSON is a custom JSON unmarshaler for DataCenterInfo, populating either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i *DataCenterInfo) UnmarshalJSON(b []byte) error {
	p := preliminaryDataCenterInfo{
		Metadata: make(map[string]string, 11),
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	adaptDataCenterInfo(i, p)
	return nil
}
