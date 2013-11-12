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

// EurekaUrlSlugs is a map of resource names -> eureka URLs
var EurekaURLSlugs = map[string]string{
	"Apps":      "apps",
	"Instances": "instances",
}

// EurekaConnection is the settings required to make eureka requests
type EurekaConnection struct {
	ServiceUrls    []string
	Timeout        int
	PollInterval   int
	PreferSameZone bool
}

// GetAppsResponse lets us deserialize the eureka/v2/apps response XML
type GetAppsResponse struct {
	VersionDelta int           `xml:"versions__delta"`
	AppsHashCode string        `xml:"apps__hashcode"`
	Applications []Application `xml:"application"`
}

// Application deserializeable from Eureka XML
type Application struct {
	Name      string     `xml:"name"`
	Instances []Instance `xml:"instance"`
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
	XMLName          struct{}       `xml:"instance"`
	HostName         string         `xml:"hostName"`
	App              string         `xml:"app"`
	IPAddr           string         `xml:"ipAddr"`
	VipAddress       string         `xml:"vipAddress"`
	SecureVipAddress string         `xml:"secureVipAddress"`
	Status           StatusType     `xml:"status"`
	Port             int            `xml:"port"`
	SecurePort       int            `xml:"securePort"`
	DataCenterInfo   DataCenterInfo `xml:"dataCenterInfo"`
	LeaseInfo        LeaseInfo      `xml:"leaseInfo"`
	//Metadata         AppMetadataType `xml:"appMetadataType"`
}

// AppMetadataType is extra properties attachable to Eureka Instances
// TODO: Actually serialize this
type AppMetadataType map[string]string

// AmazonMetadataType is information about AZ's, AMI's, and the AWS instance
type AmazonMetadataType struct {
	// <xsd:complexType name="amazonMetdataType">
	// from http://docs.amazonwebservices.com/AWSEC2/latest/DeveloperGuide/index.html?AESDG-chapter-instancedata.html
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
