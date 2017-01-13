package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"context"
	"fmt"
	"time"

	"github.com/hudl/fargo"
)

func ExampleEurekaConnection_ScheduleVIPAddressUpdates_manual() {
	e := makeConnection()
	done := make(chan struct{})
	time.AfterFunc(2*time.Minute, func() {
		close(done)
	})
	vipAddress := "my_vip"
	// We only care about those instances that are available to receive requests.
	updates, err := e.ScheduleVIPAddressUpdates(vipAddress, false, true, done, fargo.ThatAreUp, fargo.Shuffled)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Monitoring VIP address %q.\n", vipAddress)
	for update := range updates {
		if update.Err != nil {
			fmt.Printf("Most recent request for VIP address %q's instances failed: %v\n", vipAddress, update.Err)
			continue
		}
		fmt.Printf("VIP address %q has %d instances available.\n", vipAddress, len(update.Instances))
	}
	fmt.Printf("Done monitoring VIP address %q.\n", vipAddress)
}

func ExampleEurekaConnection_ScheduleVIPAddressUpdates_context() {
	e := makeConnection()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	vipAddress := "my_vip"
	// Look for instances that are in trouble.
	updates, err := e.ScheduleVIPAddressUpdates(vipAddress, false, true, ctx.Done(), fargo.WithStatus(fargo.DOWN), fargo.WithStatus(fargo.OUTOFSERVICE))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Monitoring VIP address %q.\n", vipAddress)
	for update := range updates {
		if update.Err != nil {
			fmt.Printf("Most recent request for VIP address %q's instances failed: %v\n", vipAddress, update.Err)
			continue
		}
		fmt.Printf("VIP address %q has %d instances in trouble.\n", vipAddress, len(update.Instances))
	}
	fmt.Printf("Done monitoring VIP address %q.\n", vipAddress)
}

func ExampleEurekaConnection_ScheduleSecureVIPAddressUpdates_context() {
	e := makeConnection()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	svipAddress := "my_vip"
	updates, err := e.ScheduleVIPAddressUpdates(svipAddress, true, true, ctx.Done(), fargo.ThatAreUp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Monitoring secure VIP address %q.\n", svipAddress)
	for update := range updates {
		if update.Err != nil {
			fmt.Printf("Most recent request for secure VIP address %q's instances failed: %v\n", svipAddress, update.Err)
			continue
		}
		fmt.Printf("Secure VIP address %q has %d instances.\n", svipAddress, len(update.Instances))
	}
	fmt.Printf("Done monitoring secure VIP address %q.\n", svipAddress)
}
