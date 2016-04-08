package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"time"
)

func ExampleEurekaConnection_ScheduleAppUpdates(e *EurekaConnection) {
	done := make(chan struct{})
	time.AfterFunc(2*time.Minute, func() {
		close(done)
	})
	name := "my_app"
	fmt.Printf("Monitoring application %q.\n", name)
	for update := range e.ScheduleAppUpdates(name, true, done) {
		if update.Err != nil {
			fmt.Printf("Most recent request for application %q failed: %v\n", name, update.Err)
			continue
		}
		fmt.Printf("Application %q has %d instances.\n", name, len(update.App.Instances))
	}
	fmt.Printf("Done monitoring application %q.\n", name)
}

func ExampleAppSource_Latest(e *EurekaConnection) {
	name := "my_app"
	source := e.NewAppSource(name, false)
	defer source.Stop()
	time.Sleep(30 * time.Second)
	if app := source.Latest(); app != nil {
		fmt.Printf("Application %q has %d instances\n.", name, len(app.Instances))
	}
	time.Sleep(time.Minute)
	if app := source.Latest(); app == nil {
		fmt.Printf("No application named %q is available.\n", name)
	}
}

func ExampleAppSource_CopyLatestTo(e *EurekaConnection) {
	name := "my_app"
	source := e.NewAppSource(name, true)
	defer source.Stop()
	var app Application
	if !source.CopyLatestTo(&app) {
		fmt.Printf("No application named %q is available.\n", name)
	}
	time.Sleep(time.Minute)
	if source.CopyLatestTo(&app) {
		fmt.Printf("Application %q has %d instances\n.", name, len(app.Instances))
	}
}
