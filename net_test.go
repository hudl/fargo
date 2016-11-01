package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func instancePredicateFrom(t *testing.T, opts ...InstanceQueryOption) func(*Instance) bool {
	var mergedOptions instanceQueryOptions
	for _, o := range opts {
		if err := o(&mergedOptions); err != nil {
			t.Fatal(err)
		}
	}
	if pred := mergedOptions.predicate; pred != nil {
		return pred
	}
	t.Fatal("no predicate available")
	panic("unreachable")
}

type countingSource struct {
	callCount uint
	seed      int64
}

func (s *countingSource) Int63() int64 {
	s.callCount++
	return s.seed
}

func (s *countingSource) Seed(seed int64) {
	s.seed = seed
}

func (s *countingSource) Reset() {
	s.callCount = 0
}

func TestInstanceQueryOptions(t *testing.T) {
	Convey("A status predicate", t, func() {
		Convey("mandates a nonempty status", func() {
			var opts instanceQueryOptions
			err := WithStatus("")(&opts)
			So(err, ShouldNotBeNil)
			So(opts.predicate, ShouldBeNil)
		})
		matchesStatus := func(pred func(*Instance) bool, status StatusType) bool {
			return pred(&Instance{Status: status})
		}
		Convey("matches a single status", func() {
			var opts instanceQueryOptions
			desiredStatus := UNKNOWN
			err := WithStatus(desiredStatus)(&opts)
			So(err, ShouldBeNil)
			pred := opts.predicate
			So(pred, ShouldNotBeNil)
			So(matchesStatus(pred, desiredStatus), ShouldBeTrue)
			for _, status := range []StatusType{UP, DOWN, STARTING, OUTOFSERVICE} {
				So(status, ShouldNotEqual, desiredStatus)
				So(matchesStatus(pred, status), ShouldBeFalse)
			}
		})
		Convey("matches a set of states", func() {
			var opts instanceQueryOptions
			desiredStates := []StatusType{DOWN, OUTOFSERVICE}
			for _, status := range desiredStates {
				err := WithStatus(status)(&opts)
				So(err, ShouldBeNil)
			}
			pred := opts.predicate
			So(pred, ShouldNotBeNil)
			for _, status := range desiredStates {
				So(matchesStatus(pred, status), ShouldBeTrue)
			}
			for _, status := range []StatusType{UP, STARTING, UNKNOWN} {
				So(desiredStates, ShouldNotContain, status)
				So(matchesStatus(pred, status), ShouldBeFalse)
			}
		})
	})
	Convey("A shuffling directive", t, func() {
		Convey("using the global Rand instance", func() {
			var opts instanceQueryOptions
			err := Shuffled()(&opts)
			So(err, ShouldBeNil)
			So(opts.intn, ShouldNotBeNil)
			So(opts.intn(1), ShouldEqual, 0)
		})
		Convey("using a specific Rand instance", func() {
			source := countingSource{}
			var opts instanceQueryOptions
			err := ShuffledWith(rand.New(&source))(&opts)
			So(err, ShouldBeNil)
			So(opts.intn, ShouldNotBeNil)
			So(source.callCount, ShouldEqual, 0)
			So(opts.intn(2), ShouldEqual, 0)
			So(source.callCount, ShouldEqual, 1)
		})
	})
}

func TestFilterInstancesInApps(t *testing.T) {
	Convey("A predicate should preserve only those instances", t, func() {
		Convey("with status UP", func() {
			areUp := instancePredicateFrom(t, ThatAreUp())
			Convey("from an empty set of applications", func() {
				So(filterInstancesInApps(nil, areUp), ShouldBeEmpty)
			})
			Convey("from a single application with no instances", func() {
				So(filterInstancesInApps([]*Application{
					&Application{},
				}, areUp), ShouldBeEmpty)
			})
			Convey("from a single application with one DOWN instance", func() {
				So(filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{&Instance{Status: DOWN}},
					},
				}, areUp), ShouldBeEmpty)
			})
			Convey("from a single application with one UP instance", func() {
				instance := &Instance{Status: UP}
				filtered := filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{instance},
					},
				}, areUp)
				So(filtered, ShouldHaveLength, 1)
				So(filtered, ShouldContain, instance)
			})
			Convey("from a single application with multiple instances", func() {
				upInstance := &Instance{Status: UP}
				justHasUpInstance := func(instances ...*Instance) {
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: instances,
						},
					}, areUp)
					So(filtered, ShouldHaveLength, 1)
					So(filtered, ShouldContain, upInstance)
				}
				downInstance := &Instance{Status: DOWN}
				Convey("with UP instance first", func() {
					justHasUpInstance(upInstance, downInstance)
				})
				Convey("with UP instance last", func() {
					justHasUpInstance(downInstance, upInstance)
				})
				Convey("with multiple UP instances", func() {
					secondUpInstance := &Instance{Status: UP}
					thirdUpInstance := &Instance{Status: UP}
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, downInstance, secondUpInstance, thirdUpInstance, &Instance{Status: OUTOFSERVICE}},
						},
					}, areUp)
					So(filtered, ShouldHaveLength, 3)
					So(filtered, ShouldContain, upInstance)
					So(filtered, ShouldContain, secondUpInstance)
					So(filtered, ShouldContain, thirdUpInstance)
				})
			})
			Convey("from multiple applications", func() {
				firstUpInstance := &Instance{Status: UP}
				secondUpInstance := &Instance{Status: UP}
				filtered := filterInstancesInApps([]*Application{
					&Application{
						Instances: []*Instance{firstUpInstance, &Instance{Status: OUTOFSERVICE}},
					},
					&Application{},
					&Application{
						Instances: []*Instance{&Instance{Status: DOWN}, secondUpInstance},
					},
					&Application{
						Instances: []*Instance{&Instance{Status: UNKNOWN}},
					},
				}, areUp)
				So(filtered, ShouldHaveLength, 2)
				So(filtered, ShouldContain, firstUpInstance)
				So(filtered, ShouldContain, secondUpInstance)
			})
		})
		Convey("with status matching any of those designated", func() {
			upInstance := &Instance{Status: UP}
			downInstance := &Instance{Status: DOWN}
			startingInstance := &Instance{Status: STARTING}
			outOfServiceInstance := &Instance{Status: OUTOFSERVICE}
			pred := instancePredicateFrom(t, WithStatus(DOWN), WithStatus(OUTOFSERVICE))
			Convey("from a single application", func() {
				Convey("with no matching instances", func() {
					So(filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, startingInstance},
						},
					}, pred), ShouldBeEmpty)
				})
				Convey("with two matching instances", func() {
					filtered := filterInstancesInApps([]*Application{
						&Application{
							Instances: []*Instance{upInstance, downInstance, startingInstance, outOfServiceInstance},
						},
					}, pred)
					So(filtered, ShouldHaveLength, 2)
					So(filtered, ShouldContain, downInstance)
					So(filtered, ShouldContain, outOfServiceInstance)
				})
			})
		})
	})
}
