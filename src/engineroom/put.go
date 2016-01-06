package engineroom

import (
	"encoding/base64"
	"fmt"
	"github.com/nadroz/azure-sdk-for-go/storage"
)

func Put(queueName, message string) {
	client := getStorageClient()
	params := storage.PutMessageParameters{}
	bytes := []byte(message)
	body := base64.StdEncoding.EncodeToString(bytes)

	client.PutMessage(queueName, body, params)

	fmt.Printf("Put message:\n%s\n", message)
}
