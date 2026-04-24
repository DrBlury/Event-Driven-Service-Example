package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	nethttp "net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	HeaderUUID     = "X-Watermill-UUID"
	HeaderMetadata = "X-Watermill-Metadata"
)

var ErrHTTPTransportUnsupported = errors.New("protoflow http transport is not supported in this repository")

type MarshalMessageFunc func(topic string, msg *message.Message) (*nethttp.Request, error)

type UnmarshalMessageFunc func(topic string, req *nethttp.Request) (*message.Message, error)

type PublisherConfig struct {
	MarshalMessageFunc                MarshalMessageFunc
	Client                            *nethttp.Client
	DoNotLogResponseBodyOnServerError bool
}

type SubscriberConfig struct {
	UnmarshalMessageFunc UnmarshalMessageFunc
}

type Publisher struct{}

type Subscriber struct{}

func DefaultMarshalMessageFunc(url string, msg *message.Message) (*nethttp.Request, error) {
	req, err := nethttp.NewRequest(nethttp.MethodPost, url, bytes.NewReader(msg.Payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderUUID, msg.UUID)

	metadataJSON, err := json.Marshal(msg.Metadata)
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderMetadata, string(metadataJSON))

	return req, nil
}

func DefaultUnmarshalMessageFunc(_ string, req *nethttp.Request) (*message.Message, error) {
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(req.Header.Get(HeaderUUID), payload)
	metadata := req.Header.Get(HeaderMetadata)
	if metadata == "" {
		return msg, nil
	}

	if err := json.Unmarshal([]byte(metadata), &msg.Metadata); err != nil {
		return nil, err
	}

	return msg, nil
}

func NewPublisher(PublisherConfig, watermill.LoggerAdapter) (*Publisher, error) {
	return nil, ErrHTTPTransportUnsupported
}

func (Publisher) Publish(string, ...*message.Message) error {
	return ErrHTTPTransportUnsupported
}

func (Publisher) Close() error {
	return nil
}

func NewSubscriber(string, SubscriberConfig, watermill.LoggerAdapter) (*Subscriber, error) {
	return nil, ErrHTTPTransportUnsupported
}

func (*Subscriber) Subscribe(context.Context, string) (<-chan *message.Message, error) {
	return nil, ErrHTTPTransportUnsupported
}

func (*Subscriber) Close() error {
	return nil
}

func (*Subscriber) StartHTTPServer() error {
	return ErrHTTPTransportUnsupported
}
