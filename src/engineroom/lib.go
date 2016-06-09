package engineroom

import (
	"config"
	"github.com/nadroz/azure-sdk-for-go/storage"
	"os"
	"time"
)

var conf *config.AzureConfig

type MessageDuration struct {
	MessageId          string
	ThroughputDuration time.Duration
}

func LoadAzureConfig(configFile, environment string) error {
	var err error // Cannot use :=, or we will create a new conf scoped locally
	conf, err = config.LoadAzureConfig(configFile, environment)

	return err
}

func getStorageClient() storage.QueueServiceClient {
	repo, err := storage.NewBasicClient(conf.Name, conf.AccessKey)
	if err != nil {
		os.Exit(1)
	}
	client := repo.GetQueueService()
	return client
}

func averageDuration(durations []MessageDuration) time.Duration {
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
				continue
			}

			for i := range messages.QueueMessagesList {
				out <- messages.QueueMessagesList[i]
			}
		}
		close(out)
	}()
	return out
}
