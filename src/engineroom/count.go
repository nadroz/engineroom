package engineroom

import (
	"coordinator"
	"os"
)

func Count(queueNames []string, silent bool) []coordinator.Queue {
	client := getStorageClient()

	var queues []coordinator.Queue

	for i := range queueNames {
		queueName := queueNames[i]
		depth, err := client.GetQueueDepth(queueName)
		if err != nil {
			os.Exit(1)
		}
		var queue coordinator.Queue
		queue.Name = queueName
		queue.Depth = int(depth)
		queues = append(queues, queue)
	}
	if silent != true {
		coordinator.ReportDepth(queues)
	}
	return queues
}
