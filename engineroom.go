package main

import (
    "fmt"
    "github.com/docopt/docopt-go"
	"github.com/nadroz/azure-sdk-for-go/storage"
	"os"
)

func main(){
usage := `engineroom - an azure message queue client
Usage:
  engineroom count [<queueName>]
  engineroom scan [ -a ] [<queuePrefix>]
  
  engineroom -h | --help
  engineroom --version
Arguments:
  container     	The name of the container to query
  blobspec      	A reference to one or more blobs (e.g. "mycontainer/foo", "mycontainer/")
  blobpath			The path of a blob (e.g. "mycontainer/foo.txt")
Options:
  -a                all queues
  -h, --help     	Show this screen.
  --version     	Show version.
The most commonly used commands are:
   ls         	Lists containers and blobs
   get          Downloads a blob
   put          Uploads a blob
   tree         Prints the contents of a container as a tree
   rm           Deletes a blob
`

	dict := parse(usage, "EngineRoom 0.0")
	doIt(dict)
}

func doIt(dict map[string]interface{}){
	if dict["count"].(bool){
			queueName := dict["<queueName>"].(string)
			if queueName == ""{
				os.Exit(1)
			}
			count(queueName)
		}

	if dict["scan"].(bool){
		queuePrefix := ""
		if !dict["-a"].(bool){
			queuePrefix = dict["<queuePrefix>"].(string)
		}
		scan(queuePrefix)
	}
}

//To do: get a better handle on what docopt.Parse arguments do.
func parse(usage, version string) map[string]interface{} {
	dict, err := docopt.Parse(usage, nil, true, version, false)
	
	if err != nil{
		os.Exit(1)
	}
	return dict
}

func count(queueName string){
	queue := getStorageClient()

	messages, err := queue.GetMessages(queueName, storage.GetMessagesParameters{10, 5})
	if err != nil{
			os.Exit(1)
		}
	fmt.Printf("%s: %d\n", queueName, len(messages.QueueMessagesList))
}

func scan(queuePrefix string){
	queue := getStorageClient()
	
	var matchPrefix bool
	if queuePrefix != ""{
		matchPrefix = true
	}
	
	queueList, err := queue.ListQueues(storage.ListQueuesParameters{matchPrefix, queuePrefix})
	if err != nil{
		fmt.Println("ListQueues bombed!")
		os.Exit(1)
	}

	fmt.Println("Prefix: " + queueList.Prefix)
	for i := (len(queueList.Queues) -1); i > 0; i-- {
		count(queueList.Queues[i])	
	}
}

func getStorageClient() storage.QueueServiceClient {
	//should these be more configurable?
	const account = "yourAccount"
	const key = "yourKey"

	repo, err:= storage.NewBasicClient(account, key)
	if err != nil{
			os.Exit(1)
		}
	queue := repo.GetQueueService()
	return queue
}
