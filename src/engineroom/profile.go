package engineroom

import (
	"coordinator"
	"fmt"
	// "os"
	// "strings"
	"time"
)

//this needs to maintain a buffer of durations... on which to operate
// then send a calcualted average down a channel to the coordinator/reader
//wait 2 sec to peek again
//only recalculate when a dequeue/enqueue occurs
func Profile(queueName string, duration time.Duration, bufferSize int) {
	//loop for a time duration... peek the queue
	pollTimer := time.NewTimer(duration).C
	pollTicker := time.NewTicker(time.Duration(100) * time.Millisecond).C
	nameChan := make(chan string)
	resChan := make(chan coordinator.QueueStatus)
	defer close(resChan)
	messageChan := peekMessages(nameChan)
	go coordinator.ReportMovingAverage(resChan)

	//do timer
	go func() {
		ellipseCount := 0
		for {
			select {
			case <-pollTimer:
				close(nameChan)
				fmt.Printf("\ntime up after %v\n", duration)
				return
			case <-pollTicker:
				fmt.Fprintf(os.Stderr, "Polling queues%s%s\r", strings.Repeat(".", ellipseCount), strings.Repeat(" ", 4-ellipseCount))
				nameChan <- queueName
				ellipseCount++
				ellipseCount %= 4
			}
		}
	}()

	var durations []MessageDuration
	lastId := ""
	for message := range messageChan {
		if lastId == message.MessageID {
			continue
		}
		lastId = message.MessageID

		now := time.Now().UTC()
		ins, _ := time.Parse(time.RFC1123, message.InsertionTime)
		current := MessageDuration{message.MessageID, now.Sub(ins.UTC())}

		durations = append(durations, current)
		if len(durations) >= bufferSize {
			avg := averageDuration(durations)
			current := Fetch([]string{queueName}, true)
			resChan <- coordinator.QueueStatus{queueName, current[0].Depth, avg}
			durations = durations[1:]
		}
	}
}
