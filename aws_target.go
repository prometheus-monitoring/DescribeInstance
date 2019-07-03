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

func GetTargetsAWS(list []target) []target {
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	// sess.Config.Region = aws.String("ap-south-1")

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
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
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentialsFromCreds(credValue)},
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, p := range partitions {
		if p.ID() == "aws" {
			for id, _ := range p.Regions() {
				if id == "ap-south-1" || id == "ca-central-1" || id == "ap-east-1" {
					continue
				}
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
						var info instanceInfo
						var t target
						info.Zone = *instance.Placement.AvailabilityZone
						info.Hostname = *instance.Tags[0].Value
						info.IPprivate = *instance.PrivateIpAddress
						info.IPpublic = *instance.PublicIpAddress
						addr := info.IPpublic + ":11011"
						t.Targets = append(t.Targets, addr)
						t.Labels = info
						list = append(list, t)
					}
				}
			}
		}
	}
	return list
}
