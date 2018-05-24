package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

func (e *EurekaConnection) generateURL(slugs ...string) string {
	return strings.Join(append([]string{e.SelectServiceURL()}, slugs...), "/")
}

func (e *EurekaConnection) marshal(v interface{}) ([]byte, error) {
	if e.UseJson {
		out, err := json.Marshal(v)
		if err != nil {
			// marshal the JSON *with* indents so it's readable in the error message
			out, _ := json.MarshalIndent(v, "", "    ")
			log.Errorf("Error marshalling JSON value=%v. Error:\"%s\" JSON body=\"%s\"", v, err.Error(), string(out))
			return nil, err
		}
		return out, nil
	} else {
		out, err := xml.Marshal(v)
		if err != nil {
			// marshal the XML *with* indents so it's readable in the error message
			out, _ := xml.MarshalIndent(v, "", "    ")
			log.Errorf("Error marshalling XML value=%v. Error:\"%s\" JSON body=\"%s\"", v, err.Error(), string(out))
			return nil, err
		}
		return out, nil
	}
}

// GetApp returns a single eureka application by name
func (e *EurekaConnection) GetApp(name string) (*Application, error) {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], name)
	reqURL := e.generateURL(slug)
	log.Debugf("Getting app %s from url %s", name, reqURL)
	out, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		log.Errorf("Couldn't get app %s, error: %s", name, err.Error())
		return nil, err
	}
	if rcode == 404 {
		log.Errorf("App %s not found (received 404)", name)
		return nil, AppNotFoundError{specific: name}
	}
	if rcode > 299 || rcode < 200 {
		log.Warningf("Non-200 rcode of %d", rcode)
	}

	var v *Application
	if e.UseJson {
		var r GetAppResponseJson
		err = json.Unmarshal(out, &r)
		v = &r.Application
	} else {
		err = xml.Unmarshal(out, &v)
	}
	if err != nil {
		log.Errorf("Unmarshalling error: %s", err.Error())
		return nil, err
	}

	v.ParseAllMetadata()
	return v, nil
}

func (e *EurekaConnection) readAppInto(app *Application) error {
	tapp, err := e.GetApp(app.Name)
	if err == nil {
		*app = *tapp
	}
	return err
}

// GetApps returns a map of all Applications
func (e *EurekaConnection) GetApps() (map[string]*Application, error) {
	slug := EurekaURLSlugs["Apps"]
	reqURL := e.generateURL(slug)
	log.Debugf("Getting all apps from url %s", reqURL)
	body, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		log.Errorf("Couldn't get apps, error: %s", err.Error())
		return nil, err
	}
	if rcode > 299 || rcode < 200 {
		log.Warningf("Non-200 rcode of %d", rcode)
	}

	var r *GetAppsResponse
	if e.UseJson {
		var rj GetAppsResponseJson
		err = json.Unmarshal(body, &rj)
		r = rj.Response
	} else {
		err = xml.Unmarshal(body, &r)
	}
	if err != nil {
		log.Errorf("Unmarshalling error: %s", err.Error())
		return nil, err
	}

	apps := map[string]*Application{}
	for i, a := range r.Applications {
		apps[a.Name] = r.Applications[i]
	}
	for name, app := range apps {
		log.Debugf("Parsing metadata for app %s", name)
		app.ParseAllMetadata()
	}
	return apps, nil
}

func instanceCount(apps []*Application) int {
	count := 0
	for _, app := range apps {
		count += len(app.Instances)
	}
	return count
}

type instanceQueryOptions struct {
	// predicate guides filtering, indicating whether to retain an instance when it returns true or
	// drop it when it returns false.
	predicate func(*Instance) bool
	// intn behaves like the rand.Rand.Intn function, aiding in randomizing the order of the result
	// sequence when non-nil.
	intn func(int) int
}

// InstanceQueryOption is a customization supplied to instance query functions like
// GetInstancesByVIPAddress to tailor the set of instances returned.
type InstanceQueryOption func(*instanceQueryOptions) error

func retainIfStatusIs(status StatusType, o *instanceQueryOptions) {
	if prev := o.predicate; prev != nil {
		o.predicate = func(instance *Instance) bool {
			return prev(instance) || instance.Status == status
		}
	} else {
		o.predicate = func(instance *Instance) bool {
			return instance.Status == status
		}
	}
}

// WithStatus restricts the set of instances returned to only those with the given status.
//
// Supplying multiple options produced by this function applies their logical disjunction.
func WithStatus(status StatusType) InstanceQueryOption {
	return func(o *instanceQueryOptions) error {
		if len(status) == 0 {
			return errors.New("invalid instance status")
		}
		retainIfStatusIs(status, o)
		return nil
	}
}

// ThatAreUp restricts the set of instances returned to only those with status UP.
//
// Combining this function with the options produced by WithStatus applies their logical
// disjunction.
func ThatAreUp(o *instanceQueryOptions) error {
	retainIfStatusIs(UP, o)
	return nil
}

// Shuffled requests randomizing the order of the sequence of instances returned, using the default
// shared rand.Source.
func Shuffled(o *instanceQueryOptions) error {
	o.intn = rand.Intn
	return nil
}

// ShuffledWith requests randomizing the order of the sequence of instances returned, using the
// supplied source of random numbers.
func ShuffledWith(r *rand.Rand) InstanceQueryOption {
	return func(o *instanceQueryOptions) error {
		o.intn = r.Intn
		return nil
	}
}

func shuffleInstances(instances []*Instance, intn func(int) int) {
	count := len(instances)
	if count < 2 {
		return
	}
	if intn(2) == 0 {
		instances[1], instances[0] = instances[0], instances[1]
	}
	for i := 2; i != count; i++ {
		if j := intn(i + 1); j != i {
			instances[i], instances[j] = instances[j], instances[i]
		}
	}
}

// filterInstances returns a filtered subset of the supplied sequence of instances, retaining only those
// instances for which the supplied predicate function returns true. It returns the retained instances in
// the same order the occurred in the input sequence. Note that the returned slice may share storage with
// the input sequence.
//
// The filtering algorithm is arguably baroque, in the interest of efficiency: namely, eliminating
// allocation and copying when we can avoid it. We only need to allocate and copy elements of the
// input sequence when the result sequence contains at least two nonadjacent subsequences of the
// input sequence. That is, if the predicate is, say, retaining only instances with status "UP", we
// can avoid copying elements and instead return a subsequence of the input sequence in the
// following cases (where "U" indicates an instance with status "UP", "d" with status "DOWN"):
//
//   ∙ No instances are "UP"
//     |dddd|
//
//   ∙ A single contiguous run of instances are "UP", preceded or followed by a possibly empty contiguous
//     sequence of instances that are not "UP"
//
//     |UUUU|
//     |ddUU|
//     |UUdd|
//     |dUUd|
//
// Conversely, in the following cases, no contiguous subsequence of the input sequence captures the
// set of "UP" instances:
//
//   ∙ Two or more contiguous runs of instances that are "UP" are interrupted by runs of instances
//     that are not "UP"
//
//     |UUdU|
//     |UddU|
//
// There, it's necessary to copy the "UP" instances to a fresh sequence in order to collapse them
// over the intervening "DOWN" instances.
//
// A high-level sketch of the algorithm:
//
//   Find a subsequence to retain, then try to find a second one.
//   If there is a second one, switch to copying elements to a fresh sequence to retain them.
//   Otherwise, return the lone retained subsequence, if any.
//
// In more detail:
//
//   Find the first element of the sequence to retain.
//   If there are none, return an empty sequence.
//   Otherwise
//     Note the first dropped sequence as range [0,firstBegin).
//     Find the next element to drop, looking for the end of the first subsequence to retain.
//     If there are none, the retained sequence runs through the end; return the subsequence [firstBegin,end).
//     Otherwise
//       Note the first retained sequence as range [firstBegin,firstEnd).
//       Note the second dropped sequence as range [firstEnd,secondBegin).
//       Allocate a fresh array to collect the two or more retained sequences we've found.
//       Copy the first retained sequence into the array.
//       Copy the first element at the start of second retained sequence into the array.
//       Continue collecting any remaining retained elements into the array.
//       Return the populated subsequence of the array.
//
// The algorithm evaluates the predicate exactly once for each element of the input sequence.
func filterInstances(instances []*Instance, pred func(*Instance) bool) []*Instance {
	for firstBegin, instance := range instances {
		if !pred(instance) {
			continue
		}
		// We found the first item to keep. Where is the next item to drop?
		for firstEnd, count := firstBegin+1, len(instances); firstEnd != count; firstEnd++ {
			if !pred(instances[firstEnd]) {
				// We found the first range of items to keep, followed by at least one to drop.
				for secondBegin := firstEnd + 1; secondBegin != count; secondBegin++ {
					if instance := instances[secondBegin]; pred(instance) {
						// We found at least one other item to keep, so we'll have to concatenate the first range with the rest.
						filtered := make([]*Instance, firstEnd-firstBegin+1, count-firstBegin-(secondBegin-firstEnd))
						filtered[copy(filtered, instances[firstBegin:firstEnd])] = instance
						for _, instance := range instances[secondBegin+1:] {
							if pred(instance) {
								filtered = append(filtered, instance)
							}
						}
						return filtered
					}
				}
				return instances[firstBegin:firstEnd]
			}
		}
		return instances[firstBegin:]
	}
	return nil
}

func filterInstancesInApps(apps []*Application, pred func(*Instance) bool) []*Instance {
	switch len(apps) {
	case 0:
		return nil
	case 1:
		return filterInstances(apps[0].Instances, pred)
	default:
		instances := make([]*Instance, 0, instanceCount(apps))
		for _, app := range apps {
			for _, instance := range app.Instances {
				if pred(instance) {
					instances = append(instances, instance)
				}
			}
		}
		return instances
	}
}

func (e *EurekaConnection) getInstancesByVIPAddress(addr string, secure bool, opts instanceQueryOptions) ([]*Instance, error) {
	var slug string
	if secure {
		slug = EurekaURLSlugs["InstancesBySecureVIPAddress"]
	} else {
		slug = EurekaURLSlugs["InstancesByVIPAddress"]
	}
	reqURL := e.generateURL(slug, addr)
	log.Debugf("Getting instances for VIP address %q from URL %s", addr, reqURL)
	body, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		return nil, err
	}
	if rcode != http.StatusOK {
		return nil, &unsuccessfulHTTPResponse{rcode, "unable to retrieve instances by VIP address"}
	}
	var r *GetAppsResponse
	if e.UseJson {
		var rj GetAppsResponseJson
		err = json.Unmarshal(body, &rj)
		r = rj.Response
	} else {
		err = xml.Unmarshal(body, &r)
	}
	if err != nil {
		log.Errorf("Unmarshalling error: %s", err.Error())
		return nil, err
	}
	var instances []*Instance
	if pred := opts.predicate; pred != nil {
		instances = filterInstancesInApps(r.Applications, pred)
	} else {
		switch len(r.Applications) {
		case 0:
		case 1:
			instances = r.Applications[0].Instances
		default:
			instances = make([]*Instance, instanceCount(r.Applications))
			base := 0
			for _, app := range r.Applications {
				base += copy(instances[base:], app.Instances)
			}
		}
	}
	if intn := opts.intn; intn != nil {
		shuffleInstances(instances, intn)
	}
	return instances, nil
}

func mergeInstanceQueryOptions(defaults instanceQueryOptions, opts []InstanceQueryOption) (instanceQueryOptions, error) {
	for _, o := range opts {
		if o != nil {
			if err := o(&defaults); err != nil {
				return instanceQueryOptions{}, err
			}
		}
	}
	return defaults, nil
}

func collectInstanceQueryOptions(opts []InstanceQueryOption) (instanceQueryOptions, error) {
	return mergeInstanceQueryOptions(instanceQueryOptions{}, opts)
}

// GetInstancesByVIPAddress returns the set of instances registered with the given VIP address,
// selecting either an insecure or secure VIP address with the given name, potentially filtered
// per the constraints supplied as options.
//
// NB: The VIP address is case-sensitive, and must match the address used at registration time.
func (e *EurekaConnection) GetInstancesByVIPAddress(addr string, secure bool, opts ...InstanceQueryOption) ([]*Instance, error) {
	options, err := collectInstanceQueryOptions(opts)
	if err != nil {
		return nil, err
	}
	return e.getInstancesByVIPAddress(addr, secure, options)
}

// InstanceSetUpdate is the outcome of an attempt to get a fresh snapshot of a Eureka VIP address's
// set of instances, together with an error that may have occurred in that attempt. If the Err field
// is nil, the Instances field will be populated—though possibly with an empty set.
type InstanceSetUpdate struct {
	Instances []*Instance
	Err       error
}

func exchangeInstancesEvery(d time.Duration, produce func() ([]*Instance, error), consume func([]*Instance, error), done <-chan struct{}) {
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		select {
		case <-done:
			return
		case <-t.C:
			instances, err := produce()
			consume(instances, err)
		}
	}
}

func scheduleInstanceUpdates(d time.Duration, produce func() ([]*Instance, error), await bool, done <-chan struct{}) <-chan InstanceSetUpdate {
	c := make(chan InstanceSetUpdate, 1)
	if await {
		instances, err := produce()
		c <- InstanceSetUpdate{instances, err}
	}
	consume := func(instances []*Instance, err error) {
		// Drop attempted sends when the consumer hasn't received the last buffered update.
		select {
		case c <- InstanceSetUpdate{instances, err}:
		default:
		}
	}
	go func() {
		defer close(c)
		exchangeInstancesEvery(d, produce, consume, done)
	}()
	return c
}

func (e *EurekaConnection) scheduleVIPAddressUpdates(addr string, secure bool, await bool, done <-chan struct{}, opts instanceQueryOptions) <-chan InstanceSetUpdate {
	produce := func() ([]*Instance, error) {
		return e.getInstancesByVIPAddress(addr, secure, opts)
	}
	return scheduleInstanceUpdates(e.PollInterval, produce, await, done)
}

// ScheduleVIPAddressUpdates starts polling for updates to the set of instances registered with the
// given Eureka VIP address, selecting either an insecure or secure VIP address with the given name,
// potentially filtered per the constraints supplied as options, using the connection's configured
// polling interval as its period. It sends the outcome of each update attempt to the returned
// channel, and continues until the supplied done channel is either closed or has a value available.
// Once done sending updates to the returned channel, it closes it.
//
// If await is true, it sends at least one instance set update outcome to the returned channel
// before returning.
//
// It returns an error if any of the supplied options are invalid, precluding it from scheduling the
// intended updates.
func (e *EurekaConnection) ScheduleVIPAddressUpdates(addr string, secure bool, await bool, done <-chan struct{}, opts ...InstanceQueryOption) (<-chan InstanceSetUpdate, error) {
	options, err := collectInstanceQueryOptions(opts)
	if err != nil {
		return nil, err
	}
	return e.scheduleVIPAddressUpdates(addr, secure, await, done, options), nil
}

func (e *EurekaConnection) makeInstanceProducerForApp(name string, opts []InstanceQueryOption) (func() ([]*Instance, error), error) {
	options, err := collectInstanceQueryOptions(opts)
	if err != nil {
		return nil, err
	}
	predicate := options.predicate
	intn := options.intn
	return func() ([]*Instance, error) {
		app, err := e.GetApp(name)
		if err != nil {
			return nil, err
		}
		instances := app.Instances
		if instances != nil {
			if predicate != nil {
				instances = filterInstances(instances, predicate)
			}
			if intn != nil {
				shuffleInstances(instances, intn)
			}
		}
		return instances, nil
	}, nil
}

// ScheduleAppInstanceUpdates starts polling for updates to the set of instances from the Eureka
// application with the given name, potentially filtered per the constraints supplied as options,
// using the connection's configured polling interval as its period. It sends the outcome of each
// update attempt to the returned channel, and continues until the supplied done channel is either
// closed or has a value available. Once done sending updates to the returned channel, it closes it.
//
// If await is true, it sends at least one instance set update outcome to the returned channel
// before returning.
//
// It returns an error if any of the supplied options are invalid, precluding it from scheduling the
// intended updates.
func (e *EurekaConnection) ScheduleAppInstanceUpdates(name string, await bool, done <-chan struct{}, opts ...InstanceQueryOption) (<-chan InstanceSetUpdate, error) {
	produce, err := e.makeInstanceProducerForApp(name, opts)
	if err != nil {
		return nil, err
	}
	return scheduleInstanceUpdates(e.PollInterval, produce, await, done), nil
}

// An InstanceSetSource holds a periodically updated set of instances registered with Eureka.
type InstanceSetSource struct {
	m         sync.RWMutex
	instances []*Instance
	done      chan<- struct{}
}

func (e *EurekaConnection) newInstanceSetSourceFor(produce func() ([]*Instance, error), await bool) *InstanceSetSource {
	done := make(chan struct{})
	s := &InstanceSetSource{
		done: done,
	}
	// NB: If an application contained no instances, such that it either lacked the "instance" field
	// entirely or had it present but with a "null" value, or none of the present instances
	// satisfied the filtering predicate, then it's possible that the slice returned by
	// getInstancesByVIPAddress (or similar) will be nil. Make it possible to discern when we've
	// received at least one update in Latest by never storing a nil value for a successful update.
	if await {
		if instances, err := produce(); err == nil {
			if instances != nil {
				s.instances = instances
			} else {
				s.instances = []*Instance{}
			}
		}
	}
	consume := func(instances []*Instance, err error) {
		var latest []*Instance
		if err == nil {
			if instances != nil {
				latest = instances
			} else {
				latest = []*Instance{}
			}
		}
		s.m.Lock()
		s.instances = latest
		s.m.Unlock()
	}
	go exchangeInstancesEvery(e.PollInterval, produce, consume, done)
	return s
}

func (e *EurekaConnection) newInstanceSetSourceForVIPAddress(addr string, secure bool, await bool, opts instanceQueryOptions) *InstanceSetSource {
	produce := func() ([]*Instance, error) {
		return e.getInstancesByVIPAddress(addr, secure, opts)
	}
	return e.newInstanceSetSourceFor(produce, await)
}

// NewInstanceSetSourceForVIPAddress returns a new InstantSetSource that offers a periodically
// updated set of instances registered with the given Eureka VIP address, selecting either an
// insecure or secure VIP address with the given name, potentially filtered per the constraints
// supplied as options, using the connection's configured polling interval as its period.
//
// If await is true, it waits for the first instance set update to complete before returning, though
// it's possible that that first update attempt could fail, so that a subsequent call to Latest
// would return nil.
//
// It returns an error if any of the supplied options are invalid, precluding it from scheduling the
// intended updates.
func (e *EurekaConnection) NewInstanceSetSourceForVIPAddress(addr string, secure bool, await bool, opts ...InstanceQueryOption) (*InstanceSetSource, error) {
	options, err := collectInstanceQueryOptions(opts)
	if err != nil {
		return nil, err
	}
	return e.newInstanceSetSourceForVIPAddress(addr, secure, await, options), nil
}

// NewInstanceSetSourceForApp returns a new InstantSetSource that offers a periodically updated set
// of instances from the Eureka application with the given name, potentially filtered per the
// constraints supplied as options, using the connection's configured polling interval as its
// period.
//
// If await is true, it waits for the first instance set update to complete before returning, though
// it's possible that that first update attempt could fail, so that a subsequent call to Latest
// would return nil.
//
// It returns an error if any of the supplied options are invalid, precluding it from scheduling the
// intended updates.
func (e *EurekaConnection) NewInstanceSetSourceForApp(name string, await bool, opts ...InstanceQueryOption) (*InstanceSetSource, error) {
	produce, err := e.makeInstanceProducerForApp(name, opts)
	if err != nil {
		return nil, err
	}
	return e.newInstanceSetSourceFor(produce, await), nil
}

// Latest returns the most recently acquired set of Eureka instances, if any. If the most recent
// update attempt failed, or if no update attempt has yet to complete, it returns nil.
//
// Note that if the most recent update attempt was successful but resulted in no instances, it
// returns a non-nil empty slice.
func (s *InstanceSetSource) Latest() []*Instance {
	if s == nil {
		return nil
	}
	s.m.RLock()
	defer s.m.RUnlock()
	return s.instances
}

// Stop turns off an InstantSetSource, so that it will no longer attempt to update its latest set of
// Eureka instances.
//
// It is safe to call Latest or CopyLatestTo on a stopped source.
func (s *InstanceSetSource) Stop() {
	if s == nil {
		return
	}
	// Allow multiple calls to Stop by precluding repeated attempts to close an already closed
	// channel.
	s.m.Lock()
	defer s.m.Unlock()
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
}

// RegisterInstance will register the given Instance with eureka if it is not already registered,
// but DOES NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) RegisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)
	log.Debugf("Registering instance with url %s", reqURL)
	_, rcode, err := getBody(reqURL+"/"+ins.Id(), e.UseJson)
	if err != nil {
		log.Errorf("Failed check if Instance=%s exists in app=%s, error: %s",
			ins.Id(), ins.App, err.Error())
		return err
	}
	if rcode == http.StatusOK {
		log.Noticef("Instance=%s already exists in App=%s, aborting registration", ins.Id(), ins.App)
		return nil
	}
	log.Noticef("Instance=%s not yet registered with App=%s, registering.", ins.Id(), ins.App)
	return e.ReregisterInstance(ins)
}

// ReregisterInstance will register the given Instance with eureka but DOES
// NOT automatically send heartbeats. See HeartBeatInstance for that
// functionality
func (e *EurekaConnection) ReregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s", EurekaURLSlugs["Apps"], ins.App)
	reqURL := e.generateURL(slug)

	var out []byte
	var err error
	if e.UseJson {
		out, err = e.marshal(&RegisterInstanceJson{ins})
	} else {
		out, err = e.marshal(ins)
	}
	if err != nil {
		return err
	}

	body, rcode, err := postBody(reqURL, out, e.UseJson)
	if err != nil {
		log.Errorf("Could not complete registration, error: %s", err.Error())
		return err
	}
	if rcode != 204 {
		log.Warningf("HTTP returned %d registering Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return &unsuccessfulHTTPResponse{rcode, "possible failure registering instance"}
	}

	// read back our registration to pick up eureka-supplied values
	e.readInstanceInto(ins)

	return nil
}

// GetInstance gets an Instance from eureka given its app and instanceid.
func (e *EurekaConnection) GetInstance(app, insId string) (*Instance, error) {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], app, insId)
	reqURL := e.generateURL(slug)
	log.Debugf("Getting instance with url %s", reqURL)
	body, rcode, err := getBody(reqURL, e.UseJson)
	if err != nil {
		return nil, err
	}
	if rcode != http.StatusOK {
		return nil, &unsuccessfulHTTPResponse{rcode, "unable to retrieve instance"}
	}
	var ins *Instance
	if e.UseJson {
		var ij RegisterInstanceJson
		err = json.Unmarshal(body, &ij)
		ins = ij.Instance
	} else {
		err = xml.Unmarshal(body, &ins)
	}
	return ins, err
}

func (e *EurekaConnection) readInstanceInto(ins *Instance) error {
	tins, err := e.GetInstance(ins.App, ins.Id())
	if err == nil {
		tins.UniqueID = ins.UniqueID
		*ins = *tins
	}
	return err
}

// DeregisterInstance will deregister the given Instance from eureka. This is good practice
// to do before exiting or otherwise going off line.
func (e *EurekaConnection) DeregisterInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)
	log.Debugf("Deregistering instance with url %s", reqURL)

	rcode, err := deleteReq(reqURL)
	if err != nil {
		log.Errorf("Could not complete deregistration, error: %s", err.Error())
		return err
	}
	// Eureka promises to return HTTP status code upon deregistration success, but fargo used to accept status code 204
	// here instead. Accommodate both for backward compatibility with any fake or proxy Eureka stand-ins.
	if rcode != http.StatusOK && rcode != http.StatusNoContent {
		log.Warningf("HTTP returned %d deregistering Instance=%s App=%s", rcode, ins.Id(), ins.App)
		return &unsuccessfulHTTPResponse{rcode, "possible failure deregistering instance"}
	}

	return nil
}

// AddMetadataString to a given instance. Is immediately sent to Eureka server.
func (e EurekaConnection) AddMetadataString(ins *Instance, key, value string) error {
	slug := fmt.Sprintf("%s/%s/%s/metadata", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)

	params := map[string]string{key: value}

	log.Debugf("Updating instance metadata url=%s metadata=%s", reqURL, params)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Errorf("Could not complete update, error: %s", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warningf("HTTP returned %d updating metadata Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return &unsuccessfulHTTPResponse{rcode, "possible failure updating instance metadata"}
	}
	ins.SetMetadataString(key, value)
	return nil
}

// UpdateInstanceStatus updates the status of a given instance with eureka.
func (e EurekaConnection) UpdateInstanceStatus(ins *Instance, status StatusType) error {
	slug := fmt.Sprintf("%s/%s/%s/status", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)

	params := map[string]string{"value": string(status)}

	log.Debugf("Updating instance status url=%s value=%s", reqURL, status)
	body, rcode, err := putKV(reqURL, params)
	if err != nil {
		log.Error("Could not complete update, error: ", err.Error())
		return err
	}
	if rcode < 200 || rcode >= 300 {
		log.Warningf("HTTP returned %d updating status Instance=%s App=%s Body=\"%s\"", rcode,
			ins.Id(), ins.App, string(body))
		return &unsuccessfulHTTPResponse{rcode, "possible failure updating instance status"}
	}
	return nil
}

// HeartBeatInstance sends a single eureka heartbeat. Does not continue sending
// heartbeats. Errors if the response is not 200.
func (e *EurekaConnection) HeartBeatInstance(ins *Instance) error {
	slug := fmt.Sprintf("%s/%s/%s", EurekaURLSlugs["Apps"], ins.App, ins.Id())
	reqURL := e.generateURL(slug)
	log.Debugf("Sending heartbeat with url %s", reqURL)
	req, err := http.NewRequest("PUT", reqURL, nil)
	if err != nil {
		log.Errorf("Could not create request for heartbeat, error: %s", err.Error())
		return err
	}
	_, rcode, err := netReq(req)
	if err != nil {
		log.Errorf("Error sending heartbeat for Instance=%s App=%s, error: %s", ins.Id(), ins.App, err.Error())
		return err
	}
	if rcode != http.StatusOK {
		log.Errorf("Sending heartbeat for Instance=%s App=%s returned code %d", ins.Id(), ins.App, rcode)
		return &unsuccessfulHTTPResponse{rcode, "heartbeat failed"}
	}
	return nil
}

func (i *Instance) Id() string {
	if i.InstanceId != "" {
		return i.InstanceId
	}
	if i.UniqueID != nil {
		return i.UniqueID(*i)
	}

	if i.DataCenterInfo.Name == "Amazon" {
		return i.DataCenterInfo.Metadata.InstanceID
	}

	return i.HostName
}
