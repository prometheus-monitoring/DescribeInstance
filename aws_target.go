package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getCredentialValue() credentials.Value {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	credsDir := homeDir + "/.aws/credentials"
	creds := credentials.NewSharedCredentials(credsDir, "default")
	credValue, err := creds.Get()
	if err != nil {
		log.Fatal(err.Error())
	}
	return credValue
}

func GetTargetsAWS(list []target) []target {
	credValue := getCredentialValue()
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentialsFromCreds(credValue)},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, p := range partitions {
		if p.ID() == "aws" {
			for id, _ := range p.Regions() {
				// if id == "ap-south-1" || id == "ca-central-1" || id == "ap-east-1" {
				// 	continue
				// }
				sess.Config.Region = aws.String(id)
				ec2Svc := ec2.New(sess)
				params := &ec2.DescribeInstancesInput{
					Filters: []*ec2.Filter{
						{
							Name:   aws.String("instance-state-name"),
							Values: []*string{aws.String("running")},
						},
					},
				}
				output, err := ec2Svc.DescribeInstances(params)
				if err != nil {
					log.Fatal(err.Error())
				}

				for _, reservation := range output.Reservations {
					for _, instance := range reservation.Instances {
						// var info instanceInfo
						var t target
						t.Labels["zone"] = *instance.Placement.AvailabilityZone
						t.Labels["hostname"] = *instance.Tags[0].Value
						t.Labels["ip"] = *instance.PublicIpAddress
						t.Labels["ip_priv"] = *instance.PrivateIpAddress
						// info.Zone = *instance.Placement.AvailabilityZone
						// info.Hostname = *instance.Tags[0].Value
						// info.IPprivate = *instance.PrivateIpAddress
						// info.IPpublic = *instance.PublicIpAddress
						addr := t.Labels["ip"] + ":11011"
						t.Targets = append(t.Targets, addr)
						list = append(list, t)
					}
				}
			}
		}
	}
	return list
}
