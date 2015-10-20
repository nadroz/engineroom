package main

import (
	"config"
	"coordinator"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/nadroz/azure-sdk-for-go/storage"
	"os"
	"strconv"
	"time"
)

const (
	usage string = `engineroom - an azure message queue client
Usage:
	engineroom count [ -F configFile ] [ -e environment ] <queueName>
	engineroom scan [ -F configFile ] [ -e environment ] ( -a | <queuePrefix> )
	engineroom tp [ -F configFile ] [ -e environment ] <queueName>
	engineroom profile [ -F configFile ] [ -e environment ] <queueName> [<duration>]
Arguments:
	queueName    The name(s) of one or more queues
	queuePrefix  A prefix for filtering which queues to show
	duration     A length of time, in seconds [default: 10]
Options:
	-a           All queues
	-e=env       Azure Storage Services account [default: default]
 	-F=file      Alternate configuration file [default: /usr/local/etc/engineroom/config]
	-h, --help   Show this screen.
	--version    Show version.
The most commonly used commands are:
	count        Prints the number of messages in one or more queues
	scan         Lists the queues in a storage container
	tp           How long for one message to traverse the queue
	profile      Moving average of queue depth and throughput over time
`
	version string = "EngineRoom 0.2"
)

var conf *config.AzureConfig

func main() {
	dict := parse(usage, version)
	doIt(dict)
}

func doIt(dict map[string]interface{}) {

	fmt.Printf("%v\n", dict)

	// load config
	configFile := dict["-F"].(string)
	environment := dict["-e"].(string)

	var err error // Cannot use :=, or we will create a new conf scoped locally
	conf, err = config.LoadAzureConfig(configFile, environment)
	if err != nil {
		fmt.Printf("Failed to load config: %s\n", err)
		return
	}

	if dict["count"].(bool) {
		if dict["<queueName>"] == nil {
			fmt.Println(usage)
			return
		}

		queueName := dict["<queueName>"].(string)
		if queueName == "" {
			os.Exit(1)
		}
		var queueNames []string
		queueNames = append(queueNames, queueName)
		count(queueNames, false)
	}

	if dict["scan"].(bool) {
		queuePrefix := ""
		if !dict["-a"].(bool) {
			if dict["<queuePrefix>"] == nil {
				fmt.Println(usage)
				return
			}

			queuePrefix = dict["<queuePrefix>"].(string)
		}
		scan(queuePrefix)
	}

	if dict["tp"].(bool) {
		if dict["<queueName>"] == nil {
			fmt.Println(usage)
			return
		}

		queueName := dict["<queueName>"].(string)
		measureThroughput(queueName)
	}

	if dict["profile"].(bool) {
		if dict["<queueName>"] == nil {
			fmt.Println(usage)
			return
		}

		queueName := dict["<queueName>"].(string)
		dur, _ := strconv.ParseInt(dict["<duration>"].(string), 10, 64)
		profile(queueName, int(dur))
	}
}

//To do: get a better handle on what docopt.Parse arguments do.
func parse(usage, version string) map[string]interface{} {
	dict, err := docopt.Parse(usage, nil, true, version, false)

	if err != nil {
		os.Exit(1)
	}
	return dict
}

func count(queueNames []string, silent bool) []coordinator.Queue {
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

func scan(queuePrefix string) {
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
	count(queueNames, false)
}

func measureThroughput(queueName string) time.Duration {
	nameChan := make(chan string)
	c := peekMessages(nameChan)

	nameChan <- queueName
	close(nameChan)
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

func peekMessages(queueNames chan string) <-chan storage.PeekMessageResponse {
	client := getStorageClient()
	out := make(chan storage.PeekMessageResponse)
	go func() {
		for name := range queueNames {
			messages, err := client.PeekMessages(name, storage.PeekMessagesParameters{1})
			if err != nil {
				out <- storage.PeekMessageResponse{}
			}

			for i := range messages.QueueMessagesList {
				out <- messages.QueueMessagesList[i]
			}
		}
		close(out)
	}()
	return out
}

//this needs to maintain a buffer of durations... on which to operate
// then send a calcualted average down a channel to the coordinator/reader
//wait 2 sec to peek again
//only recalculate when a dequeue/enqueue occurs
func profile(queueName string, seconds int) {
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
			currentQ := count(queueColl, true)
			resChan <- coordinator.Queue{queueName, currentQ[0].Depth, avg}
			coll = coll[1:]
		}
	}
	close(resChan)
}

func doAverage(durations []MessageDuration) time.Duration {
	var dur time.Duration = 0
	for i := range durations {
		dur += durations[i].ThroughputDuration
	}
	return dur / (time.Duration(len(durations)))
}

func getDuration(start string) time.Duration {
	now := time.Now().UTC()
	ins, err := time.Parse(time.RFC1123, start)
	if err != nil {
		os.Exit(1)
	}
	ins = ins.UTC()
	dif := now.Sub(ins)
	fmt.Println("get duration:")
	fmt.Println(dif)
	return dif
}

func getStorageClient() storage.QueueServiceClient {
	repo, err := storage.NewBasicClient(conf.Name, conf.AccessKey)
	if err != nil {
		os.Exit(1)
	}
	client := repo.GetQueueService()
	return client
}

type MessageDuration struct {
	MessageId          string
	ThroughputDuration time.Duration
}
