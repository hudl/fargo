package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"time"
)

// EurekaURLSlugs is a map of resource names->Eureka URLs.
var EurekaURLSlugs = map[string]string{
	"Apps":                        "apps",
	"Instances":                   "instances",
	"InstancesByVIPAddress":       "vips",
	"InstancesBySecureVIPAddress": "svips",
}

// EurekaConnection is the settings required to make Eureka requests.
type EurekaConnection struct {
	ServiceUrls    []string
	ServicePort    int
	ServerURLBase  string
	Timeout        time.Duration
	PollInterval   time.Duration
	PreferSameZone bool
	Retries        int
	DNSDiscovery   bool
	DiscoveryZone  string
	discoveryTtl   chan struct{}
	UseJson        bool
}

// GetAppsResponseJson lets us deserialize the eureka/v2/apps response JSON—a wrapped GetAppsResponse.
type GetAppsResponseJson struct {
	Response *GetAppsResponse `json:"applications"`
}

// GetAppsResponse lets us deserialize the eureka/v2/apps response XML.
type GetAppsResponse struct {
	Applications  []*Application `xml:"application" json:"application"`
	AppsHashcode  string         `xml:"apps__hashcode" json:"apps__hashcode"`
	VersionsDelta int            `xml:"versions__delta" json:"versions__delta"`
}

// GetAppResponseJson wraps an Application for deserializing from Eureka JSON.
type GetAppResponseJson struct {
	Application Application `json:"application"`
}

// Application deserializeable from Eureka XML.
type Application struct {
	Name      string      `xml:"name" json:"name"`
	Instances []*Instance `xml:"instance" json:"instance"`
}

// StatusType is an enum of the different statuses allowed by Eureka.
type StatusType string

// Supported statuses
const (
	UP           StatusType = "UP"
	DOWN         StatusType = "DOWN"
	STARTING     StatusType = "STARTING"
	OUTOFSERVICE StatusType = "OUT_OF_SERVICE"
	UNKNOWN      StatusType = "UNKNOWN"
)

// Datacenter names
const (
	Amazon = "Amazon"
	MyOwn  = "MyOwn"
)

// RegisterInstanceJson lets us serialize the eureka/v2/apps/<ins> request JSON—a wrapped Instance.
type RegisterInstanceJson struct {
	Instance *Instance `json:"instance"`
}

// Instance [de]serializeable [to|from] Eureka [XML|JSON].
type Instance struct {
	InstanceId       string `xml:"instanceId" json:"instanceId"`
	HostName         string `xml:"hostName" json:"hostName"`
	App              string `xml:"app" json:"app"`
	IPAddr           string `xml:"ipAddr" json:"ipAddr"`
	VipAddress       string `xml:"vipAddress" json:"vipAddress"`
	SecureVipAddress string `xml:"secureVipAddress" json:"secureVipAddress"`

	Status           StatusType `xml:"status" json:"status"`
	Overriddenstatus StatusType `xml:"overriddenstatus" json:"overriddenstatus"`

	Port              int  `xml:"-" json:"-"`
	PortEnabled       bool `xml:"-" json:"-"`
	SecurePort        int  `xml:"-" json:"-"`
	SecurePortEnabled bool `xml:"-" json:"-"`

	HomePageUrl    string `xml:"homePageUrl" json:"homePageUrl"`
	StatusPageUrl  string `xml:"statusPageUrl" json:"statusPageUrl"`
	HealthCheckUrl string `xml:"healthCheckUrl" json:"healthCheckUrl"`

	CountryId      int64          `xml:"countryId" json:"countryId"`
	DataCenterInfo DataCenterInfo `xml:"dataCenterInfo" json:"dataCenterInfo"`

	LeaseInfo LeaseInfo        `xml:"leaseInfo" json:"leaseInfo"`
	Metadata  InstanceMetadata `xml:"metadata" json:"metadata"`

	UniqueID func(i Instance) string `xml:"-" json:"-"`
}

// InstanceMetadata represents the eureka metadata, which is arbitrary XML.
// See metadata.go for more info.
type InstanceMetadata struct {
	Raw    []byte `xml:",innerxml" json:"-"`
	parsed map[string]interface{}
}

// AmazonMetadataType is information about AZ's, AMI's, and the AWS instance.
// <xsd:complexType name="amazonMetdataType">
// from http://docs.amazonwebservices.com/AWSEC2/latest/DeveloperGuide/index.html?AESDG-chapter-instancedata.html
type AmazonMetadataType struct {
	AmiLaunchIndex   string `xml:"ami-launch-index" json:"ami-launch-index"`
	LocalHostname    string `xml:"local-hostname" json:"local-hostname"`
	AvailabilityZone string `xml:"availability-zone" json:"availability-zone"`
	InstanceID       string `xml:"instance-id" json:"instance-id"`
	PublicIpv4       string `xml:"public-ipv4" json:"public-ipv4"`
	PublicHostname   string `xml:"public-hostname" json:"public-hostname"`
	AmiManifestPath  string `xml:"ami-manifest-path" json:"ami-manifest-path"`
	LocalIpv4        string `xml:"local-ipv4" json:"local-ipv4"`
	HostName         string `xml:"hostname" json:"hostname"`
	AmiID            string `xml:"ami-id" json:"ami-id"`
	InstanceType     string `xml:"instance-type" json:"instance-type"`
}

// DataCenterInfo indicates which type of data center hosts this instance
// and conveys details about the instance's environment.
type DataCenterInfo struct {
	// Name indicates which type of data center hosts this instance.
	Name string
	// Class indicates the Java class name representing this structure in the Eureka server,
	// noted only when encoding communication with JSON.
	//
	// When registering an instance, if the name is neither "Amazon" nor "MyOwn", this field's
	// value is used. Otherwise, a suitable default value will be supplied to the server. This field
	// is available for specifying custom data center types other than the two built-in ones, for
	// which no suitable default value could be known.
	Class string
	// Metadata provides details specific to an Amazon data center,
	// populated and honored when the Name field's value is "Amazon".
	Metadata AmazonMetadataType
	// AlternateMetadata provides details specific to a data center other than Amazon,
	// populated and honored when the Name field's value is not "Amazon".
	AlternateMetadata map[string]string
}

// LeaseInfo tells us about the renewal from Eureka, including how old it is.
type LeaseInfo struct {
	RenewalIntervalInSecs int32 `xml:"renewalIntervalInSecs" json:"renewalIntervalInSecs"`
	DurationInSecs        int32 `xml:"durationInSecs" json:"durationInSecs"`
	RegistrationTimestamp int64 `xml:"registrationTimestamp" json:"registrationTimestamp"`
	LastRenewalTimestamp  int64 `xml:"lastRenewalTimestamp" json:"lastRenewalTimestamp"`
	EvictionTimestamp     int64 `xml:"evictionTimestamp" json:"evictionTimestamp"`
	ServiceUpTimestamp    int64 `xml:"serviceUpTimestamp" json:"serviceUpTimestamp"`
}
