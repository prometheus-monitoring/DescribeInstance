package lib

import (
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
)

func (ts Targets) GetTargetsGCP(logLevel *logrus.Logger) ([]Target, error) {
	filters := [...]string{
		"status = RUNNING",
	}

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return ts, err
	}
	computeService, err := compute.New(c)
	logLevel.Info("[gcp] Get list zones on project zingplayinternational-097")
	zoneListCall := computeService.Zones.List("zingplayinternational-097")
	zoneList, err := zoneListCall.Do()
	if err != nil {
		return ts, nil
	}
	for _, zone := range zoneList.Items {
		instanceListCall := computeService.Instances.List("zingplayinternational-097", zone.Name)
		instanceListCall.Filter(strings.Join(filters[:], " "))
		logLevel.Infof("[gcp] Get list instances in zone %s", zone.Name)
		instanceList, err := instanceListCall.Do()
		if err != nil {
			logLevel.Error(err)
			continue
		}
		logLevel.Info("[gcp] Create list targets")
		for _, instance := range instanceList.Items {
			t := new(Target)
			t.Labels = make(map[string]string)
			t.Labels["zone"] = zone.Name
			t.Labels["instance"] = instance.Name
			t.Labels["product_code"] = "ZPTGSN"
			t.Labels["ip"] = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
			t.Labels["ip_priv"] = instance.NetworkInterfaces[0].NetworkIP
			addr := t.Labels["ip"] + ":11011"
			t.Targets = append(t.Targets, addr)
			ts = append(ts, *t)
		}
	}
	return ts, err
}
