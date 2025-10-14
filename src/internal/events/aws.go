package events

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	amazonsns "github.com/ThreeDotsLabs/watermill-aws/sns"
	amazonsqs "github.com/ThreeDotsLabs/watermill-aws/sqs"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func (s *Service) createAwsPublisher(ctx context.Context, logger watermill.LoggerAdapter) {
	// Build AWS config (uses default credential chain)
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	pub, err := amazonsns.NewPublisher(amazonsns.PublisherConfig{
		TopicResolver: amazonsns.GenerateArnTopicResolver{},
		AWSConfig:     cfg,
		Marshaler:     amazonsns.DefaultMarshalerUnmarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}
	s.Publisher = pub
}

func (s *Service) createAwsSubscriber(ctx context.Context, logger watermill.LoggerAdapter) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	sqsSub, err := amazonsqs.NewSubscriber(amazonsqs.SubscriberConfig{
		AWSConfig:   cfg,
		Unmarshaler: amazonsqs.DefaultMarshalerUnmarshaler{},
	}, logger)
	if err != nil {
		panic(err)
	}

	s.Subscriber = sqsSub
}
