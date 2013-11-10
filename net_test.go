package eugo_test

import (
	"github.com/ryansb/eugo"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestGetAllApps(c *C) {
	e := eugo.NewConn("http", "127.0.0.1", "8080")
	a, _ := e.GetApps()
	c.Assert(a["EUREKA"].Instances[0].HostName, Equals, "localhost.localdomain")
	c.Assert(a["EUREKA"].Instances[0].IpAddr, Equals, "127.0.0.1")
}

func (s *S) TestGetAppInstances(c *C) {
	e := eugo.NewConn("http", "127.0.0.1", "8080")
	a, _ := e.GetApp("EUREKA")
	c.Assert(a.Instances[0].HostName, Equals, "localhost.localdomain")
	c.Assert(a.Instances[0].IpAddr, Equals, "127.0.0.1")
}

func (s *S) TestRegisterFakeInstance(c *C) {
	e := eugo.NewConn("http", "127.0.0.1", "8080")
	i := eugo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		App:              "TESTAPP",
		IpAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   eugo.DataCenterInfo{Name: eugo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           eugo.UP,
	}
	err := e.RegisterInstance(&i)
	c.Assert(err, IsNil)
}

func (s *S) TestCheckin(c *C) {
	e := eugo.NewConn("http", "127.0.0.1", "8080")
	i := eugo.Instance{
		HostName:         "i-123456",
		Port:             9090,
		App:              "TESTAPP",
		IpAddr:           "127.0.0.10",
		VipAddress:       "127.0.0.10",
		DataCenterInfo:   eugo.DataCenterInfo{Name: eugo.MyOwn},
		SecureVipAddress: "127.0.0.10",
		Status:           eugo.UP,
	}
	err := e.HeartBeatInstance(&i)
	c.Assert(err, IsNil)
}
