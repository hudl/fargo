package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"time"
)

// EurekaUrlSlugs is a map of resource names -> eureka URLs
var EurekaURLSlugs = map[string]string{
	"Apps":      "apps",
	"Instances": "instances",
}

// EurekaConnection is the settings required to make eureka requests
type EurekaConnection struct {
	ServiceUrls    []string
	Timeout        time.Duration
	PollInterval   time.Duration
	PreferSameZone bool
	Retries        int
}

// GetAppsResponse lets us deserialize the eureka/v2/apps response XML
type GetAppsResponse struct {
	VersionDelta int           `xml:"versions__delta"`
	AppsHashCode string        `xml:"apps__hashcode"`
	Applications []Application `xml:"application"`
}

// Application deserializeable from Eureka XML
type Application struct {
	Name      string      `xml:"name"`
	Instances []*Instance `xml:"instance"`
}

// StatusType is an enum of the different statuses allowed by Eureka
type StatusType string

// Supported statuses
const (
	UP       StatusType = "UP"
	DOWN     StatusType = "DOWN"
	STARTING StatusType = "STARTING"
)

// Datacenter names
const (
	Amazon = "Amazon"
	MyOwn  = "MyOwn"
)

// Instance [de]serializeable [to|from] Eureka XML
type Instance struct {
	XMLName          struct{}         `xml:"instance"`
	HostName         string           `xml:"hostName"`
	App              string           `xml:"app"`
	IPAddr           string           `xml:"ipAddr"`
	VipAddress       string           `xml:"vipAddress"`
	SecureVipAddress string           `xml:"secureVipAddress"`
	Status           StatusType       `xml:"status"`
	Port             int              `xml:"port"`
	SecurePort       int              `xml:"securePort"`
	DataCenterInfo   DataCenterInfo   `xml:"dataCenterInfo"`
	LeaseInfo        LeaseInfo        `xml:"leaseInfo"`
	Metadata         InstanceMetadata `xml:"metadata"`
}

// InstanceMetadata represents the eureka metadata, which is arbitrary XML. See
// metadata.go for more info.
type InstanceMetadata struct {
	Raw    []byte                 `xml:",innerxml"`
	parsed map[string]interface{} `xml:"-"`
}

// AmazonMetadataType is information about AZ's, AMI's, and the AWS instance
// <xsd:complexType name="amazonMetdataType">
// from http://docs.amazonwebservices.com/AWSEC2/latest/DeveloperGuide/index.html?AESDG-chapter-instancedata.html
type AmazonMetadataType struct {
	AmiLaunchIndex   string `xml:"ami-launch-index"`
	LocalHostname    string `xml:"local-hostname"`
	AvailabilityZone string `xml:"availability-zone"`
	InstanceID       string `xml:"instance-id"`
	PublicIpv4       string `xml:"public-ipv4"`
	PublicHostname   string `xml:"public-hostname"`
	AmiManifestPath  string `xml:"ami-manifest-path"`
	LocalIpv4        string `xml:"local-ipv4"`
	HostName         string `xml:"hostname"`
	AmiID            string `xml:"ami-id"`
	InstanceType     string `xml:"instance-type"`
}

// DataCenterInfo is only really useful when running in AWS.
type DataCenterInfo struct {
	Name     string             `xml:"name"`
	Metadata AmazonMetadataType `xml:"metadata"`
}

// LeaseInfo tells us about the renewal from Eureka, including how old it is
type LeaseInfo struct {
	RenewalIntervalInSecs int32 `xml:"renewalIntervalInSecs"`
	DurationInSecs        int32 `xml:"durationInSecs"`
	RegistrationTimestamp int64 `xml:"registrationTimestamp"`
	LastRenewalTimestamp  int64 `xml:"lastRenewalTimestamp"`
	EvictionTimestamp     int32 `xml:"evictionTimestamp"`
	ServiceUpTimestamp    int64 `xml:"serviceUpTimestamp"`
}
