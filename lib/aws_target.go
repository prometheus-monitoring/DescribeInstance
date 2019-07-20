package lib

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

func getCredentialValue(logLevel *logrus.Logger) (credValue credentials.Value, err error) {
	homeDir, _ := os.UserHomeDir()
	credsDir := homeDir + "/.aws/credentials"
	if _, err = os.Stat(credsDir); err == nil {
		logLevel.Info("[aws] Get credential values")
		creds := credentials.NewSharedCredentials(credsDir, "default")
		credValue, err = creds.Get()
		if err != nil {
			logLevel.Error("[aws] Cannot get credential values")
			return credValue, err
		}
	}
	return credValue, err
}

func (ts Targets) GetTargetsAWS(logLevel *logrus.Logger) ([]Target, error) {
	credValue, err := getCredentialValue(logLevel)
	if err == nil {
		resolver := endpoints.DefaultResolver()
		logLevel.Info("[aws] Resolve partitions")
		partitions := resolver.(endpoints.EnumPartitions).Partitions()
		sess, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentialsFromCreds(credValue)},
		)
		if err != nil {
			return ts, err
		}

		for _, p := range partitions {
			if p.ID() == "aws" {
				for id, _ := range p.Regions() {
					if id == "ap-east-1" {
						continue
					}
					sess.Config.Region = aws.String(id)

					logLevel.Infof("[aws] Create new session for get describe instances in region %s", *aws.String(id))
					ec2Svc := ec2.New(sess)

					// Get only instances with status running
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
						return ts, err
					}
					logLevel.Info("[aws] Create list targets")
					// Declared json info
					for _, reservation := range output.Reservations {
						for _, instance := range reservation.Instances {
							t := new(Target)
							if strings.Contains(*instance.Tags[0].Value, "AutoScaling") {
								continue
							}
							t.Labels = make(map[string]string)
							// t.Labels["zone"] = *instance.Placement.AvailabilityZone
							t.Labels["instance"] = *instance.Tags[0].Value
							t.Labels["product_code"] = "ZPTGSN"
							t.Labels["ip"] = *instance.PublicIpAddress
							t.Labels["ip_priv"] = *instance.PrivateIpAddress
							addr := t.Labels["ip"] + ":11011"
							t.Addrs = append(t.Addrs, addr)
							ts = append(ts, *t)
						}
					}
				}
			}
		}
	}
	return ts, err
}
