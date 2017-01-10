package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

func intFromJSONNumberOrString(jv interface{}, description string) (int, error) {
	switch v := jv.(type) {
	case float64:
		return int(v), nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unexpected %s: %[2]v (type %[2]T)", description, jv)
	}
}

// UnmarshalJSON is a custom JSON unmarshaler for GetAppsResponse to deal with
// sometimes non-wrapped Application arrays when there is only a single Application item.
func (r *GetAppsResponse) UnmarshalJSON(b []byte) error {
	marshalLog.Debugf("GetAppsResponse.UnmarshalJSON b:%s\n", string(b))
	resolveDelta := func(d interface{}) (int, error) {
		return intFromJSONNumberOrString(d, "versions delta")
	}

	// Normal array case
	type getAppsResponse GetAppsResponse
	auxArray := struct {
		*getAppsResponse
		VersionsDelta interface{} `json:"versions__delta"`
	}{
		getAppsResponse: (*getAppsResponse)(r),
	}
	var err error
	if err = json.Unmarshal(b, &auxArray); err == nil {
		marshalLog.Debugf("GetAppsResponse.UnmarshalJSON array:%+v\n", auxArray)
		r.VersionsDelta, err = resolveDelta(auxArray.VersionsDelta)
		return err
	}

	// Bogus non-wrapped case
	auxSingle := struct {
		Application   *Application `json:"application"`
		AppsHashcode  string       `json:"apps__hashcode"`
		VersionsDelta interface{}  `json:"versions__delta"`
	}{}
	if err := json.Unmarshal(b, &auxSingle); err != nil {
		return err
	}
	marshalLog.Debugf("GetAppsResponse.UnmarshalJSON single:%+v\n", auxSingle)
	if r.VersionsDelta, err = resolveDelta(auxSingle.VersionsDelta); err != nil {
		return err
	}
	r.Applications = make([]*Application, 1, 1)
	r.Applications[0] = auxSingle.Application
	r.AppsHashcode = auxSingle.AppsHashcode
	return nil
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
		marshalLog.Debugf("Application.UnmarshalJSON aa:%+v\n", aa)
		*a = Application(aa)
		return nil
	}

	// Bogus non-wrapped case
	var as applicationSingle
	if err = json.Unmarshal(b, &as); err == nil {
		marshalLog.Debugf("Application.UnmarshalJSON as:%+v\n", as)
		a.Name = as.Name
		a.Instances = make([]*Instance, 1, 1)
		a.Instances[0] = as.Instance
		return nil
	}
	return err
}

func stringAsBool(s string) bool {
	return s == "true"
}

// UnmarshalJSON is a custom JSON unmarshaler for Instance, transcribing the two composite port
// specifications up to top-level fields.
func (i *Instance) UnmarshalJSON(b []byte) error {
	// Preclude recursive calls to MarshalJSON.
	type instance Instance
	// inboundJSONFormatPort describes an instance's network port, including whether its registrant
	// considers the port to be enabled or disabled.
	//
	// Example JSON encoding:
	//
	//   Eureka versions 1.2.1 and prior:
	//     "port":{"@enabled":"true", "$":"7101"}
	//
	//   Eureka version 1.2.2 and later:
	//     "port":{"@enabled":"true", "$":7101}
	//
	// Note that later versions of Eureka write the port number as a JSON number rather than as a
	// decimal-formatted string. We accept it as either an integer or a string. Strangely, the
	// "@enabled" field remains a string.
	type inboundJSONFormatPort struct {
		Number  interface{} `json:"$"`
		Enabled bool        `json:"@enabled,string"`
	}
	aux := struct {
		*instance
		Port       inboundJSONFormatPort `json:"port"`
		SecurePort inboundJSONFormatPort `json:"securePort"`
	}{
		instance: (*instance)(i),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	resolvePort := func(port interface{}) (int, error) {
		return intFromJSONNumberOrString(port, "port number")
	}
	var err error
	if i.Port, err = resolvePort(aux.Port.Number); err != nil {
		return err
	}
	i.PortEnabled = aux.Port.Enabled
	if i.SecurePort, err = resolvePort(aux.SecurePort.Number); err != nil {
		return err
	}
	i.SecurePortEnabled = aux.SecurePort.Enabled
	return nil
}

// MarshalJSON is a custom JSON marshaler for Instance, adapting the top-level raw port values to
// the composite port specifications.
func (i *Instance) MarshalJSON() ([]byte, error) {
	// Preclude recursive calls to MarshalJSON.
	type instance Instance
	// outboundJSONFormatPort describes an instance's network port, including whether its registrant
	// considers the port to be enabled or disabled.
	//
	// Example JSON encoding:
	//
	//   "port":{"@enabled":"true", "$":"7101"}
	//
	// Note that later versions of Eureka write the port number as a JSON number rather than as a
	// decimal-formatted string. We emit the port number as a string, not knowing the Eureka
	// server's version. Strangely, the "@enabled" field remains a string.
	type outboundJSONFormatPort struct {
		Number  int  `json:"$,string"`
		Enabled bool `json:"@enabled,string"`
	}
	aux := struct {
		*instance
		Port       outboundJSONFormatPort `json:"port"`
		SecurePort outboundJSONFormatPort `json:"securePort"`
	}{
		(*instance)(i),
		outboundJSONFormatPort{i.Port, i.PortEnabled},
		outboundJSONFormatPort{i.SecurePort, i.SecurePortEnabled},
	}
	return json.Marshal(&aux)
}

// xmlFormatPort describes an instance's network port, including whether its registrant considers
// the port to be enabled or disabled.
//
// Example XML encoding:
//
//     <port enabled="true">7101</port>
type xmlFormatPort struct {
	Number  int  `xml:",chardata"`
	Enabled bool `xml:"enabled,attr"`
}

// UnmarshalXML is a custom XML unmarshaler for Instance, transcribing the two composite port
// specifications up to top-level fields.
func (i *Instance) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type instance Instance
	aux := struct {
		*instance
		Port       xmlFormatPort `xml:"port"`
		SecurePort xmlFormatPort `xml:"securePort"`
	}{
		instance: (*instance)(i),
	}
	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}
	i.Port = aux.Port.Number
	i.PortEnabled = aux.Port.Enabled
	i.SecurePort = aux.SecurePort.Number
	i.SecurePortEnabled = aux.SecurePort.Enabled
	return nil
}

// startLocalName creates a start-tag of an XML element with the given local name and no namespace name.
func startLocalName(local string) xml.StartElement {
	return xml.StartElement{Name: xml.Name{Space: "", Local: local}}
}

// MarshalXML is a custom XML marshaler for Instance, adapting the top-level raw port values to
// the composite port specifications.
func (i *Instance) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type instance Instance
	aux := struct {
		*instance
		Port       xmlFormatPort `xml:"port"`
		SecurePort xmlFormatPort `xml:"securePort"`
	}{
		instance:   (*instance)(i),
		Port:       xmlFormatPort{i.Port, i.PortEnabled},
		SecurePort: xmlFormatPort{i.SecurePort, i.SecurePortEnabled},
	}
	return e.EncodeElement(&aux, startLocalName("instance"))
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

func metadataValue(i *DataCenterInfo) interface{} {
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
func (i *DataCenterInfo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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
	Class    string      `xml:"-" json:"@class"`
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

func adaptDataCenterInfo(dst *DataCenterInfo, src *preliminaryDataCenterInfo) {
	dst.Name = src.Name
	dst.Class = src.Class
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
	adaptDataCenterInfo(i, &p)
	return nil
}

// MarshalJSON is a custom JSON marshaler for DataCenterInfo, writing either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i *DataCenterInfo) MarshalJSON() ([]byte, error) {
	type named struct {
		Name  string `json:"name"`
		Class string `json:"@class"`
	}
	if i.Name == Amazon {
		return json.Marshal(struct {
			named
			Metadata AmazonMetadataType `json:"metadata"`
		}{
			named{i.Name, "com.netflix.appinfo.AmazonInfo"},
			i.Metadata,
		})
	}
	class := "com.netflix.appinfo.MyDataCenterInfo"
	if i.Name != MyOwn {
		class = i.Class
	}
	return json.Marshal(struct {
		named
		Metadata map[string]string `json:"metadata,omitempty"`
	}{
		named{i.Name, class},
		i.AlternateMetadata,
	})
}

func jsonValueAsString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.f", v)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}, map[string]interface{}:
		// Don't bother trying to decode these.
		return ""
	case nil:
		return ""
	default:
		panic("type of unexpected value")
	}
}

// UnmarshalJSON is a custom JSON unmarshaler for DataCenterInfo, populating either Metadata or AlternateMetadata
// depending on the type of data center indicated by the Name.
func (i *DataCenterInfo) UnmarshalJSON(b []byte) error {
	// The Eureka server will mistakenly convert metadata values that look like numbers to JSON numbers.
	// Convert them back to strings.
	aux := struct {
		*preliminaryDataCenterInfo
		PreliminaryMetadata map[string]interface{} `json:"metadata"`
	}{
		PreliminaryMetadata: make(map[string]interface{}, 11),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	metadata := make(map[string]string, len(aux.PreliminaryMetadata))
	for k, v := range aux.PreliminaryMetadata {
		metadata[k] = jsonValueAsString(v)
	}
	aux.Metadata = metadata
	adaptDataCenterInfo(i, aux.preliminaryDataCenterInfo)
	return nil
}
