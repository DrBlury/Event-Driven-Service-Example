package events

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	amazonsns "github.com/aws/aws-sdk-go-v2/service/sns"
	amazonsqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/samber/lo"
)

func (s *Service) createAWSConfig(ctx context.Context, logger watermill.LoggerAdapter) *aws.Config {
	var cfg aws.Config
	// if an endpoint override is configured, use AnonymousCredentials and endpoint resolver
	if s.Conf != nil && s.Conf.AWSEndpoint != "" {
		cfg = aws.Config{
			Credentials: aws.AnonymousCredentials{},
			Region:      s.Conf.AWSRegion,
		}
	} else {
		// default SDK configuration
		c, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}
		cfg = c
		// ensure region from config if set
		if s.Conf != nil && s.Conf.AWSRegion != "" {
			cfg.Region = s.Conf.AWSRegion
		}
	}
	return &cfg
}

func (s *Service) createAwsPublisher(ctx context.Context, logger watermill.LoggerAdapter, cfg *aws.Config) *sns.Publisher {
	topicResolver, err := sns.NewGenerateArnTopicResolver(s.Conf.AWSAccountID, s.Conf.AWSRegion)
	if err != nil {
		panic(err)
	}

	publisherConfig := sns.PublisherConfig{
		TopicResolver: topicResolver,
		AWSConfig:     *cfg,
		Marshaler:     sns.DefaultMarshalerUnmarshaler{},
	}

	publisher, err := sns.NewPublisher(
		publisherConfig,
		logger,
	)
	if err != nil {
		panic(err)
	}

	return publisher
}

func (s *Service) createAwsSubscriber(ctx context.Context, logger watermill.LoggerAdapter, cfg *aws.Config) *sns.Subscriber {
	topicResolver, err := sns.NewGenerateArnTopicResolver(s.Conf.AWSAccountID, s.Conf.AWSRegion)
	if err != nil {
		panic(err)
	}

	name := "subscriber"

	snsOpts := []func(*amazonsns.Options){
		amazonsns.WithEndpointResolverV2(sns.OverrideEndpointResolver{
			Endpoint: transport.Endpoint{
				URI: *lo.Must(url.Parse(*cfg.BaseEndpoint)),
			},
		}),
	}

	sqsOpts := []func(*amazonsqs.Options){
		amazonsqs.WithEndpointResolverV2(sqs.OverrideEndpointResolver{
			Endpoint: transport.Endpoint{
				URI: *lo.Must(url.Parse(*cfg.BaseEndpoint)),
			},
		}),
	}

	subscriberConfig := sns.SubscriberConfig{
		AWSConfig: aws.Config{
			Credentials: aws.AnonymousCredentials{},
		},
		OptFns:        snsOpts,
		TopicResolver: topicResolver,
		GenerateSqsQueueName: func(ctx context.Context, snsTopic sns.TopicArn) (string, error) {
			topic, err := sns.ExtractTopicNameFromTopicArn(snsTopic)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%v-%v", topic, name), nil
		},
	}

	subscriber, err := sns.NewSubscriber(
		subscriberConfig,
		sqs.SubscriberConfig{
			AWSConfig: *cfg,
			OptFns:    sqsOpts,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	return subscriber
}
