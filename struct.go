package eurekago

type EurekaUrls struct {
	apps      string "/eureka/v2/apps"
	instances string "/eureka/v2/instances"
}

type EurekaConnection struct {
	Port    string
	Address string
	Proto   string // either http or https
}

type EurekaApps struct {
	Applications []EurekaApp `json:"applications"`
}

type EurekaApp struct{}
type EurekaInstances struct{}
type EurekaInstance struct{}
type EurekaVipAddr struct{}
type EurekaSVipAddr struct{}
