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
					var info instanceInfo
					var t target
					// info.Project = "zingplayinternational-097"
					info.Zone = zone.Name
					info.Hostname = instance.Name
					info.IPprivate = instance.NetworkInterfaces[0].NetworkIP
					info.IPpublic = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
					addr := info.IPpublic + ":11011"
					t.Targets = append(t.Targets, addr)
					t.Labels = info
					list = append(list, t)
				}
			}
		}
	}
	return list
}
