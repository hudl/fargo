package fargo_test

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"fmt"
	"time"

	"github.com/hudl/fargo"
)

func ExampleInstanceSetSource_Latest_outcomes() {
	e := makeConnection()
	vipAddress := "my_vip"
	source, err := e.NewInstanceSetSourceForVIPAddress(vipAddress, false, false, fargo.ThatAreUp)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer source.Stop()
	time.Sleep(30 * time.Second)
	if instances := source.Latest(); instances != nil {
		fmt.Printf("VIP address %q has %d instances available.\n", vipAddress, len(instances))
	}
	time.Sleep(time.Minute)
	switch instances := source.Latest(); {
	case instances == nil:
		fmt.Printf("Unsure whether any instances for VIP address %q are available.\n", vipAddress)
	case len(instances) == 0:
		fmt.Printf("No instances for VIP address %q are available.\n", vipAddress)
	default:
		fmt.Printf("VIP address %q has %d instances available.\n", vipAddress, len(instances))
	}
}

func ExampleInstanceSetSource_Latest_compare() {
	e := makeConnection()
	svipAddress := "my_vip"
	source, err := e.NewInstanceSetSourceForVIPAddress(svipAddress, true, true, fargo.WithStatus(fargo.DOWN), fargo.WithStatus(fargo.OUTOFSERVICE))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer source.Stop()
	var troubled []*fargo.Instance
	for remaining := 10; ; {
		if instances := source.Latest(); instances != nil {
			fmt.Printf("VIP address %q has %d troubled instances.", svipAddress, len(instances))
			switch diff := len(instances) - len(troubled); {
			case diff > 0:
				fmt.Printf(" (%d more than last time)\n", diff)
			case diff < 0:
				fmt.Printf(" (%d fewer than last time)\n", diff)
			default:
				fmt.Println()
			}
			troubled = instances
		}
		remaining--
		if remaining == 0 {
			break
		}
		time.Sleep(30 * time.Second)
	}
}

func ExampleInstanceSetSource_Latest_app() {
	e := makeConnection()
	name := "my_app"
	source, err := e.NewInstanceSetSourceForApp(name, true, fargo.ThatAreUp, fargo.Shuffled)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer source.Stop()
	for remaining := 10; ; {
		if instances := source.Latest(); len(instances) > 0 {
			instance := instances[0]
			// Assume the insecure port is enabled.
			fmt.Printf("Chose service %s:%d for application %q.\n", instance.IPAddr, instance.Port, name)
		}
		remaining--
		if remaining == 0 {
			break
		}
		time.Sleep(30 * time.Second)
	}
}
