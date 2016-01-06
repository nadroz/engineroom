package engineroom

import (
	"encoding/base64"
	"fmt"
	"github.com/nadroz/azure-sdk-for-go/storage"
)

func Peek(queueName string) {
	client := getStorageClient()
	params := storage.PeekMessagesParameters{
		NumOfMessages: 1,
	}

	messages, err := client.PeekMessages(queueName, params)
	if err != nil {
		fmt.Printf("Failed to peek messages: %s\n", err)
		return
	}

	if len(messages.QueueMessagesList) != 1 {
		return
	}

	msg := messages.QueueMessagesList[0]

	txt, err := base64.StdEncoding.DecodeString(msg.MessageText)

	if err != nil {
		fmt.Printf("Failed to decode message: %s\n", err)
		return
	}

	fmt.Printf("%s\n", txt)
}
