package main

import (
	"log"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

func GetTargets(list []target) []target {
	filters := [...]string{
		"status = RUNNING",
	}

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	zoneListCall := computeService.Zones.List("zingplayinternational-097")
	zoneList, err := zoneListCall.Do()
	if err != nil {
		log.Fatal(err)
	} else {
		for _, zone := range zoneList.Items {
			instanceListCall := computeService.Instances.List("zingplayinternational-097", zone.Name)
			instanceListCall.Filter(strings.Join(filters[:], " "))
			instanceList, err := instanceListCall.Do()
			if err != nil {
				log.Fatal(err)
			} else {
				for _, instance := range instanceList.Items {
					t := new(target)
					t.Labels = make(map[string]string)
					t.Labels["zone"] = zone.Name
					t.Labels["hostname"] = instance.Name
					t.Labels["ip"] = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
					t.Labels["ip_priv"] = instance.NetworkInterfaces[0].NetworkIP
					addr := t.Labels["ip"] + ":11011"
					t.Targets = append(t.Targets, addr)
					list = append(list, *t)
				}
			}
		}
	}
	return list
}
