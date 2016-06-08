package engineroom

import (
	"fmt"
	"os"
	"time"
)

func MeasureThroughput(queueName string) time.Duration {
	nameChan := make(chan string)
	c := peekMessages(nameChan)
	nameChan <- queueName
	defer close(nameChan)

	message := <-c
	now := time.Now().UTC()
	ins, err := time.Parse(time.RFC1123, message.InsertionTime)
	if err != nil {
		os.Exit(1)
	}

	ins = ins.UTC()
	dif := now.Sub(ins)
	fmt.Println(dif)
	return dif
}
