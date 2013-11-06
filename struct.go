package eurekago

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

type EurekaApps struct {
	App EurekaApp `json:"application"`
}

type EurekaApp struct {
	Name            string `json:"name"`
	EurekaInstances `json:"instance"`
}

type EurekaInstances struct {
	Instance EurekaInstance `json:"instance"`
}

type EurekaInstance struct {
	LastUpdatedTimestamp          int64                  `json:"lastUpdatedTimestamp"`
	LastDirtyTimestamp            int64                  `json:"lastDirtyTimestamp"`
	CountryId                     int16                  `json:"countryId"`
	SecurePort                    EurekaPort             `json:"securePort"`
	Port                          EurekaPort             `json:"port"`
	Status                        string                 `json:"status"`
	IpAddr                        string                 `json:"ipAddr"`
	VipAddr                       string                 `json:"vipAddress"`
	App                           string                 `json:"app"`
	HostName                      string                 `json:"hostName"`
	DataCenterInfo                EurekaDataCenter       `json:"dataCenterInfo"`
	LeaseInfo                     EurekaLease            `json:"leaseInfo"`
	Metadata                      EurekaInstanceMetadata `json:"metadata"`
	HomePageUrl                   string                 `json:"homePageUrl"`
	StatusPageUrl                 string                 `json:"statusPageUrl"`
	HealthCheckUrl                string                 `json:"healthCheckUrl"`
	IsCoordinatingDiscoveryServer string                 `json:"isCoordinatingDiscoveryServer"`
}

type EurekaInstanceMetadata struct{}

type EurekaDataCenter struct {
	Name  string `json:"name"`
	Class string `json:"@class"`
}

type EurekaLease struct {
	ServiceUpTimestamp    int64 `json:"serviceUpTimestamp"`
	EvictionTimestamp     int64 `json:"evictionTimestamp"`
	LastRenewalTimestamp  int64 `json:"lastRenewalTimestamp"`
	RegistrationTimestamp int64 `json:"registrationTimestamp"`
	DurationInSecs        int   `json:"durationInSecs"`
	RenewalIntervalInSecs int   `json:"renewalIntervalInSecs"`
}
type EurekaPort struct {
	Port    string `json:"$"`
	Enabled bool   `json:"@enabled"`
}
type EurekaVipAddr struct{}
type EurekaSVipAddr struct{}
