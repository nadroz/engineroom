package engineroom

import (
	"coordinator"
	"os"
)

func Fetch(queueNames []string, silent bool) []coordinator.QueueStatus {
	client := getStorageClient()

	var queues []coordinator.QueueStatus

	for i := range queueNames {
		queueName := queueNames[i]
		depth, err := client.GetQueueDepth(queueName)
		if err != nil {
			os.Exit(1)
		}
		var queue coordinator.QueueStatus
		queue.Name = queueName
		queue.Depth = int(depth)
		queues = append(queues, queue)
	}
	if silent != true {
		coordinator.ReportDepth(queues)
	}
	return queues
}
