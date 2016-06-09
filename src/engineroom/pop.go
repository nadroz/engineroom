package engineroom

import (
	"encoding/base64"
	"fmt"
	"github.com/nadroz/azure-sdk-for-go/storage"
)

func Pop(queueName string, count int) {
	client := getStorageClient()
	params := storage.GetMessagesParameters{
		NumOfMessages:     count,
		VisibilityTimeout: 10,
	}

	messages, err := client.GetMessages(queueName, params)
	if err != nil {
		fmt.Printf("Failed to pop message: %s\n", err)
		return
	}

	if len(messages.QueueMessagesList) != count {
		return
	}

	for _, msg := range messages.QueueMessagesList {
		err = client.DeleteMessage(queueName, msg.MessageID, msg.PopReceipt)
		if err != nil {
			fmt.Printf("Failed to pop message: %s\n", err)
			return
		}

		txt, err := base64.StdEncoding.DecodeString(msg.MessageText)
		if err != nil {
			fmt.Printf("Failed to decode message: %s\n", err)
			return
		}

		fmt.Printf("%s\n", txt)
	}
}
