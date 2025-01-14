package ec2

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/tailwarden/komiser/models"
	. "github.com/tailwarden/komiser/providers"
)

func ElasticIps(ctx context.Context, client ProviderClient) ([]Resource, error) {
	config := ec2.DescribeAddressesInput{}
	resources := make([]Resource, 0)
	ec2Client := ec2.NewFromConfig(*client.AWSClient)

	stsClient := sts.NewFromConfig(*client.AWSClient)
	stsOutput, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return resources, err
	}

	accountId := stsOutput.Account

	for {
		output, err := ec2Client.DescribeAddresses(ctx, &config)
		if err != nil {
			return resources, err
		}

		for _, elasticIps := range output.Addresses {
			tags := make([]Tag, 0)
			for _, tag := range elasticIps.Tags {
				tags = append(tags, Tag{
					Key:   *tag.Key,
					Value: *tag.Value,
				})
			}

			resourceArn := fmt.Sprintf("arn:aws:ec2:%s:%s:elastic-ip/%s", client.AWSClient.Region, *accountId, *elasticIps.AllocationId)

			resources = append(resources, Resource{
				Provider:   "AWS",
				Account:    client.Name,
				Service:    "Elastic IP",
				Region:     client.AWSClient.Region,
				ResourceId: resourceArn,
				Cost:       0,
				Name:       *elasticIps.AllocationId,
				FetchedAt:  time.Now(),
				Tags:       tags,
				Link:       fmt.Sprintf("https:/%s.console.aws.amazon.com/ec2/home?region=%s#ElasticIpDetails:AllocationId=%s", client.AWSClient.Region, client.AWSClient.Region, *elasticIps.AllocationId),
			})
		}

		log.WithFields(log.Fields{
			"provider":  "AWS",
			"account":   client.Name,
			"region":    client.AWSClient.Region,
			"service":   "Elastic IP",
			"resources": len(resources),
		}).Info("Fetched resources")
		return resources, nil
	}
}
