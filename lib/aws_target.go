package lib

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
)

func getCredentialValue(loglevel *logrus.Logger) (credValue credentials.Value, err error) {
	homeDir, _ := os.UserHomeDir()
	credsDir := homeDir + "/.aws/credentials"
	if _, err = os.Stat(credsDir); err == nil {
		loglevel.Info("[aws] Get credential values")
		creds := credentials.NewSharedCredentials(credsDir, "default")
		credValue, err = creds.Get()
		if err != nil {
			loglevel.Error("[aws] Cannot get credential values")
			return credValue, err
		}
		return credValue, nil
	}
	return credValue, err
}

func (ts Targets) GetTargetsAWS(loglevel *logrus.Logger) ([]Target, error) {
	credValue, err := getCredentialValue(loglevel)
	if err != nil {
		return ts, err
	}

	resolver := endpoints.DefaultResolver()
	loglevel.Info("Resolve partitions")
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
				// if id == "ap-south-1" || id == "ca-central-1" || id == "ap-east-1" {
				// 	continue
				// }
				sess.Config.Region = aws.String(id)

				loglevel.Infof("[aws] Create new session for get describe instances in region %s", *aws.String(id))
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
				loglevel.Info("[aws] Create list targets")
				// Declared json info
				for _, reservation := range output.Reservations {
					for _, instance := range reservation.Instances {
						t := new(Target)
						t.Labels = make(map[string]string)
						t.Labels["zone"] = *instance.Placement.AvailabilityZone
						t.Labels["hostname"] = *instance.Tags[0].Value
						t.Labels["product_code"] = "ZPTGSN"
						t.Labels["ip"] = *instance.PublicIpAddress
						t.Labels["ip_priv"] = *instance.PrivateIpAddress
						addr := t.Labels["ip"] + ":11011"
						t.Targets = append(t.Targets, addr)
						ts = append(ts, *t)
					}
				}
			}
		}
	}
	return ts, err
}
