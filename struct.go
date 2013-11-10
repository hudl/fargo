package eugo

type EurekaUrls struct {
	Apps      string
	Instances string
}

type EurekaConnection struct {
	Port    string
	Address string
	Proto   string // either http or https
	Urls    EurekaUrls
}

func NewConn(proto, address, port string) (c EurekaConnection) {
	c.Proto = proto
	c.Address = address
	c.Port = port
	c.Urls = EurekaUrls{
		Apps:      "eureka/v2/apps",
		Instances: "eureka/v2/instances",
	}
	return c
}

type GetAppsResponse struct {
	VersionDelta int           `xml:"versions__delta"`
	AppsHashCode string        `xml:"apps__hashcode"`
	Applications []Application `xml:"application"`
}

type Application struct {
	Name      string     `xml:"name"`
	Instances []Instance `xml:"instance"`
}

type StatusType string

const (
	UP       StatusType = "UP"
	DOWN     StatusType = "DOWN"
	STARTING StatusType = "STARTING"
)

const (
	Amazon = "Amazon"
	MyOwn  = "MyOwn"
)

type Instance struct {
	XMLName          struct{}       `xml:"instance"`
	HostName         string         `xml:"hostName"`
	App              string         `xml:"app"`
	IpAddr           string         `xml:"ipAddr"`
	VipAddress       string         `xml:"vipAddress"`
	SecureVipAddress string         `xml:"secureVipAddress"`
	Status           StatusType     `xml:"status"`
	Port             int            `xml:"port"`
	SecurePort       int            `xml:"securePort"`
	DataCenterInfo   DataCenterInfo `xml:"dataCenterInfo"`
	LeaseInfo        LeaseInfo      `xml:"leaseInfo"`
	//Metadata         AppMetadataType `xml:"appMetadataType"`
}

type AppMetadataType map[string]string

type AmazonMetadataType struct {
	// <xsd:complexType name="amazonMetdataType">
	// from http://docs.amazonwebservices.com/AWSEC2/latest/DeveloperGuide/index.html?AESDG-chapter-instancedata.html
	AmiLaunchIndex   string `xml:"ami-launch-index"`
	LocalHostname    string `xml:"local-hostname"`
	AvailabilityZone string `xml:"availability-zone"`
	InstanceId       string `xml:"instance-id"`
	PublicIpv4       string `xml:"public-ipv4"`
	PublicHostname   string `xml:"public-hostname"`
	AmiManifestPath  string `xml:"ami-manifest-path"`
	LocalIpv4        string `xml:"local-ipv4"`
	HostName         string `xml:"hostname"`
	AmiId            string `xml:"ami-id"`
	InstanceType     string `xml:"instance-type"`
}

type DataCenterInfo struct {
	Name     string             `xml:"name"`
	Metadata AmazonMetadataType `xml:"metadata"`
}

type LeaseInfo struct {
	RenewalIntervalInSecs int32 `xml:"renewalIntervalInSecs"`
	DurationInSecs        int32 `xml:"durationInSecs"`
	RegistrationTimestamp int64 `xml:"registrationTimestamp"`
	LastRenewalTimestamp  int64 `xml:"lastRenewalTimestamp"`
	EvictionTimestamp     int32 `xml:"evictionTimestamp"`
	ServiceUpTimestamp    int64 `xml:"serviceUpTimestamp"`
}
