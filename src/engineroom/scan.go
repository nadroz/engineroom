package engineroom

import (
	"github.com/nadroz/azure-sdk-for-go/storage"
	"os"
)

func Scan(queuePrefix string) {
	client := getStorageClient()

	var matchPrefix bool
	if queuePrefix != "" {
		matchPrefix = true
	}

	queueList, err := client.ListQueues(storage.ListQueuesParameters{matchPrefix, queuePrefix})
	if err != nil {
		os.Exit(1)
	}

	var queueNames []string
	for i := range queueList.Queues {
		queueNames = append(queueNames, queueList.Queues[i])
	}
	Fetch(queueNames, false)
}
