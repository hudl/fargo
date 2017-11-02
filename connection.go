package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SelectServiceURL gets a eureka instance based on the connection's load
// balancing scheme.
// TODO: Make this not just pick a random one.
func (e *EurekaConnection) SelectServiceURL() string {
	if e.discoveryTtl == nil {
		e.discoveryTtl = make(chan struct{}, 1)
	}
	if e.DNSDiscovery && len(e.discoveryTtl) == 0 {
		servers, ttl, err := discoverDNS(e.DiscoveryZone, e.ServicePort, e.ServerURLBase)
		if err != nil {
			return choice(e.ServiceUrls)
		}
		e.discoveryTtl <- struct{}{}
		time.AfterFunc(ttl, func() {
			// At the end of the timeout, empty the channel so that the next
			// SelectServiceURL call will refresh the DNS info
			<-e.discoveryTtl
		})
		e.ServiceUrls = servers
	}
	return choice(e.ServiceUrls)
}

func choice(options []string) string {
	if len(options) == 0 {
		log.Fatal("There are no ServiceUrls to choose from, bailing out")
	}
	return options[rand.Int()%len(options)]
}

// NewConnFromConfigFile sets up a connection object based on a config in
// specified path
func NewConnFromConfigFile(location string) (c EurekaConnection, err error) {
	cfg, err := ReadConfig(location)
	if err != nil {
		log.Errorf("Problem reading config %s error: %s", location, err.Error())
		return c, err
	}
	return NewConnFromConfig(cfg), nil
}

// NewConnFromConfig will, given a Config struct, return a connection based on
// those options
func NewConnFromConfig(conf Config) (c EurekaConnection) {
	c.ServiceUrls = conf.Eureka.ServiceUrls
	c.ServicePort = conf.Eureka.ServerPort
	if len(c.ServiceUrls) == 0 && len(conf.Eureka.ServerDNSName) > 0 {
		c.ServiceUrls = []string{conf.Eureka.ServerDNSName}
	}
	c.Timeout = time.Duration(conf.Eureka.ConnectTimeoutSeconds) * time.Second
	c.PollInterval = time.Duration(conf.Eureka.PollIntervalSeconds) * time.Second
	c.PreferSameZone = conf.Eureka.PreferSameZone
	if conf.Eureka.UseDNSForServiceUrls {
		log.Warning("UseDNSForServiceUrls is an experimental option")
		c.DNSDiscovery = true
		c.DiscoveryZone = conf.Eureka.DNSDiscoveryZone
		c.ServerURLBase = conf.Eureka.ServerURLBase
	}
	return c
}

// NewConn is a default connection with just a list of ServiceUrls. Most basic
// way to make a new connection. Generally only if you know what you're doing
// and are going to do the configuration yourself some other way.
func NewConn(address ...string) (e EurekaConnection) {
	e.ServiceUrls = address
	return e
}

// UpdateApp creates a goroutine that continues to keep an application updated
// with its status in Eureka.
func (e *EurekaConnection) UpdateApp(app *Application) {
	go func() {
		for {
			log.Noticef("Updating app %s", app.Name)
			err := e.readAppInto(app)
			if err != nil {
				log.Errorf("Failure updating %s in goroutine", app.Name)
			}
			<-time.After(time.Duration(e.PollInterval) * time.Second)
		}
	}()
}

// AppUpdate is the outcome of an attempt to get a fresh snapshot of a Eureka
// application's state, together with an error that may have occurred in that
// attempt. If the Err field is nil, the App field will be non-nil.
type AppUpdate struct {
	App *Application
	Err error
}

func exchangeAppEvery(d time.Duration, produce func() (*Application, error), consume func(*Application, error), done <-chan struct{}) {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-done:
			return
		case <-t.C:
			app, err := produce()
			consume(app, err)
		}
	}
}

// ScheduleAppUpdates starts polling for updates to the Eureka application with
// the given name, using the connection's configured polling interval as its
// period. It sends the outcome of each update attempt to the returned channel,
// and continues until the supplied done channel is either closed or has a value
// available. Once done sending updates to the returned channel, it closes it.
//
// If await is true, it sends at least one application update outcome to the
// returned channel before returning.
func (e *EurekaConnection) ScheduleAppUpdates(name string, await bool, done <-chan struct{}) <-chan AppUpdate {
	produce := func() (*Application, error) {
		return e.GetApp(name)
	}
	c := make(chan AppUpdate, 1)
	if await {
		app, err := produce()
		c <- AppUpdate{app, err}
	}
	consume := func(app *Application, err error) {
		// Drop attempted sends when the consumer hasn't received the last buffered update.
		select {
		case c <- AppUpdate{app, err}:
		default:
		}
	}
	go func() {
		defer close(c)
		exchangeAppEvery(e.PollInterval, produce, consume, done)
	}()
	return c
}

// An AppSource holds a periodically updated copy of a Eureka application.
type AppSource struct {
	m    sync.RWMutex
	app  *Application
	done chan<- struct{}
}

// NewAppSource returns a new AppSource that offers a periodically updated copy
// of the Eureka application with the given name, using the connection's
// configured polling interval as its period.
//
// If await is true, it waits for the first application update to complete
// before returning, though it's possible that that first update attempt could
// fail, so that a subsequent call to Latest would return nil and CopyLatestTo
// would return false.
func (e *EurekaConnection) NewAppSource(name string, await bool) *AppSource {
	done := make(chan struct{})
	s := &AppSource{
		done: done,
	}
	produce := func() (*Application, error) {
		return e.GetApp(name)
	}
	if await {
		if app, err := produce(); err == nil {
			s.app = app
		}
	}
	consume := func(app *Application, err error) {
		s.m.Lock()
		s.app = app
		s.m.Unlock()
	}
	go exchangeAppEvery(e.PollInterval, produce, consume, done)
	return s
}

// Latest returns the most recently acquired Eureka application, if any. If the
// most recent update attempt failed, or if no update attempt has yet to
// complete, it returns nil.
func (s *AppSource) Latest() *Application {
	if s == nil {
		return nil
	}
	s.m.RLock()
	defer s.m.RUnlock()
	return s.app
}

// CopyLatestTo copies the most recently acquired Eureka application to dst, if
// any, and returns true if such an application was available. If no preceding
// update attempt had succeeded, such that no application is available to be
// copied, it returns false.
func (s *AppSource) CopyLatestTo(dst *Application) bool {
	if s == nil {
		return false
	}
	s.m.RLock()
	defer s.m.RUnlock()
	if s.app == nil {
		return false
	}
	*dst = *s.app
	return true
}

// Stop turns off an AppSource, so that it will no longer attempt to update its
// latest application.
//
// It is safe to call Latest or CopyLatestTo on a stopped source.
func (s *AppSource) Stop() {
	if s == nil {
		return
	}
	// Allow multiple calls to Stop by precluding repeated attempts to close an
	// already closed channel.
	s.m.Lock()
	defer s.m.Unlock()
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
}
