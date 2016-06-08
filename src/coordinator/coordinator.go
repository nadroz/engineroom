package coordinator

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"time"
)

type Queue struct {
	Name              string
	Depth             int
	AverageThroughput time.Duration
}

func ReportDepth(queues []Queue) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Queue", "Depth"})
	for i := range queues {
		queue := queues[i]
		table.Append([]string{queue.Name, strconv.Itoa(queue.Depth)})
	}
	table.Render()
}

func ReportMovingAverage(queueChan chan Queue) {
	fmt.Println("Queue\tAverage Wait Time\tQueue Depth")
	for queue := range queueChan {
		fmt.Printf("\r%s\t%s\t%d\r\n", queue.Name, queue.AverageThroughput, queue.Depth)
	}
}
