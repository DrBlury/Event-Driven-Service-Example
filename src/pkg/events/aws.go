package events

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	amazonsns "github.com/aws/aws-sdk-go-v2/service/sns"
	amazonsqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	transport "github.com/aws/smithy-go/endpoints"
)

var (
	awsDefaultConfigLoader  = awsconfig.LoadDefaultConfig
	snsTopicResolverFactory = sns.NewGenerateArnTopicResolver
	snsPublisherFactory     = func(cfg sns.PublisherConfig, logger watermill.LoggerAdapter) (message.Publisher, error) {
		return sns.NewPublisher(cfg, logger)
	}
	snsSubscriberFactory = func(cfg sns.SubscriberConfig, sqsCfg sqs.SubscriberConfig, logger watermill.LoggerAdapter) (message.Subscriber, error) {
		return sns.NewSubscriber(cfg, sqsCfg, logger)
	}
)

func (s *Service) createAWSConfig(ctx context.Context) *aws.Config {
	cfg, err := awsDefaultConfigLoader(ctx)
	if err != nil {
		s.Logger.Error("Failed to load AWS default config", err, watermill.LogFields{"cfg": cfg})
		panic(err)
	}
	if s.Conf != nil && s.Conf.AWSRegion != "" {
		s.Logger.Info("Setting AWS region from config", watermill.LogFields{"region": s.Conf.AWSRegion})
		cfg.Region = s.Conf.AWSRegion
	}

	return &cfg
}

func (s *Service) createAwsPublisher(logger watermill.LoggerAdapter, cfg *aws.Config) {
	var accountID, region string
	if s.Conf != nil {
		accountID = s.Conf.AWSAccountID
		region = s.Conf.AWSRegion
	}

	accountID = strings.Trim(accountID, "\"' ")
	if accountID == "" {
		if s.Conf != nil && s.Conf.AWSEndpoint != "" {
			accountID = "000000000000"
			s.Logger.Info("AWS account ID empty; using LocalStack default", watermill.LogFields{"accountID": accountID})
		}
	}
	if len(accountID) != 12 {
		if s.Conf != nil && s.Conf.AWSEndpoint != "" {
			s.Logger.Info("Invalid AWS account ID; falling back to LocalStack default", watermill.LogFields{"accountID": accountID})
			accountID = "000000000000"
		}
	}

	s.Logger.Info("Create AWS Publisher",
		watermill.LogFields{
			"accountID": accountID,
			"region":    region,
		})

	topicResolver, err := snsTopicResolverFactory(accountID, region)
	if err != nil {
		s.Logger.Error("Failed to create SNS topic resolver", err, watermill.LogFields{
			"accountID": accountID,
			"region":    region,
		})
		panic(err)
	}

	publisherConfig := sns.PublisherConfig{
		TopicResolver: topicResolver,
		AWSConfig:     *cfg,
		Marshaler:     sns.DefaultMarshalerUnmarshaler{},
		OptFns: []func(o *amazonsns.Options){
			func(o *amazonsns.Options) {
				if s.Conf != nil && s.Conf.AWSEndpoint != "" {
					parsedURL, err := url.Parse(s.Conf.AWSEndpoint)
					if err != nil {
						s.Logger.Error("Failed to parse AWS endpoint", err, watermill.LogFields{"endpoint": s.Conf.AWSEndpoint})
						panic(err)
					}
					o.BaseEndpoint = aws.String(parsedURL.String())
				}
			},
		},
	}

	publisher, err := snsPublisherFactory(
		publisherConfig,
		logger,
	)
	if err != nil {
		panic(err)
	}

	s.Publisher = publisher
}

func (s *Service) createAwsSubscriber(logger watermill.LoggerAdapter, cfg *aws.Config) {
	var accountID, region string
	if s.Conf != nil {
		accountID = s.Conf.AWSAccountID
		region = s.Conf.AWSRegion
	}

	if len(accountID) != 12 {
		if s.Conf != nil && s.Conf.AWSEndpoint != "" {
			s.Logger.Info("Invalid AWS account ID; falling back to LocalStack default", watermill.LogFields{"accountID": accountID})
			accountID = "000000000000"
		}
	}

	topicResolver, err := snsTopicResolverFactory(accountID, region)
	if err != nil {
		panic(err)
	}

	name := "subscriber"

	var snsOpts []func(*amazonsns.Options)
	var sqsOpts []func(*amazonsqs.Options)
	snsOpts, sqsOpts = addEndpointResolver(cfg, snsOpts, sqsOpts)

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

	subscriber, err := snsSubscriberFactory(
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

	s.Subscriber = subscriber
}

func addEndpointResolver(cfg *aws.Config, snsOpts []func(*amazonsns.Options), sqsOpts []func(*amazonsqs.Options)) ([]func(*amazonsns.Options), []func(*amazonsqs.Options)) {
	if cfg != nil && cfg.BaseEndpoint != nil && *cfg.BaseEndpoint != "" {
		parsedURL, err := url.Parse(*cfg.BaseEndpoint)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse BaseEndpoint: %v", err))
		}
		snsOpts = []func(*amazonsns.Options){
			amazonsns.WithEndpointResolverV2(sns.OverrideEndpointResolver{
				Endpoint: transport.Endpoint{
					URI: *parsedURL,
				},
			}),
		}
		sqsOpts = []func(*amazonsqs.Options){
			amazonsqs.WithEndpointResolverV2(sqs.OverrideEndpointResolver{
				Endpoint: transport.Endpoint{
					URI: *parsedURL,
				},
			}),
		}
	}
	return snsOpts, sqsOpts
}
