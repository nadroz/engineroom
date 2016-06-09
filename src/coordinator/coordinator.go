package coordinator

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"time"
)

type QueueStatus struct {
	Name    string
	Depth   int
	Latency time.Duration
}

func ReportDepth(queues []QueueStatus) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Queue", "Depth"})
	for i := range queues {
		queue := queues[i]
		table.Append([]string{queue.Name, strconv.Itoa(queue.Depth)})
	}
	table.Render()
}

func ReportMovingAverage(queueChan chan QueueStatus) {
	fmt.Printf("Queue\tLatency (s)\tDepth\tTimestamp\n")
	for queue := range queueChan {
		fmt.Printf("%s\t%f\t%d\t%s\n", queue.Name, queue.Latency.Seconds(), queue.Depth, time.Now().Format("03:04:05.000000"))
	}
}
