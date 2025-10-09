package events

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

func readMessageToStruct(msg *message.Message, v interface{}) error {
	return json.Unmarshal(msg.Payload, v)
}

func createNewProcessedEvent(event proto.Message, metadata map[string]string) ([]*message.Message, error) {
	// turn into json
	jsonPayload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	newMessage := message.NewMessage(watermill.NewUUID(), jsonPayload)
	metadata["event_message_schema"] = fmt.Sprintf("%T", event)
	newMessage.Metadata = metadata
	return []*message.Message{newMessage}, nil
}
