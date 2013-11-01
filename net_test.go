package eurekago_test

import (
	. "github.com/ryansb/eureka-go"
	. "launchpad.net/gocheck"
)

type S struct{}

var _ = Suite(S{})

func (s *S) TestBadConnect(c *C) {
	e := EurekaConnection{Port: "9090", Address: "127.0.0.1"}
	a, err := e.GetApps()
	c.Assert(err, NotNil)
}
