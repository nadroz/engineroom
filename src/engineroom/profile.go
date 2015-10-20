package engineroom

import (
	"coordinator"
	"time"
)

//this needs to maintain a buffer of durations... on which to operate
// then send a calcualted average down a channel to the coordinator/reader
//wait 2 sec to peek again
//only recalculate when a dequeue/enqueue occurs
func Profile(queueName string, seconds int) {
	//loop for a time duration... peek the queue
	stopWatch := time.NewTimer(time.Duration(seconds) * time.Second).C
	nameChan := make(chan string)
	resChan := make(chan coordinator.Queue)
	messageChan := peekMessages(nameChan)
	go coordinator.ReportMovingAverage(resChan)

	//do timer
	go func() {
		for {
			select {
			case <-stopWatch:
				close(nameChan)
				return
			default:
				nameChan <- queueName
			}
		}
	}()
	var coll []MessageDuration
	var dur time.Duration
	for message := range messageChan {
		size := len(coll)
		now := time.Now().UTC()
		switch {
		case size == 0:
			ins, _ := time.Parse(time.RFC1123, message.InsertionTime)
			ins = ins.UTC()
			dur = now.Sub(ins)
			coll = append(coll, MessageDuration{message.MessageID, dur})
		case size < 10:
			if coll[len(coll)-1].MessageId != message.MessageID {
				ins, _ := time.Parse(time.RFC1123, message.InsertionTime)
				ins = ins.UTC()
				dur = now.Sub(ins)
				coll = append(coll, MessageDuration{message.MessageID, dur})
			}

		case size == 10:
			avg := doAverage(coll)
			queueColl := []string{queueName}
			currentQ := Count(queueColl, true)
			resChan <- coordinator.Queue{queueName, currentQ[0].Depth, avg}
			coll = coll[1:]
		}
	}
	close(resChan)
}
