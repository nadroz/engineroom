package main

import (
	"engineroom"
	"errors"
	"fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	usage string = `azq - an azure message queue client
Usage:
	azq count [ -F configFile ] [ -e environment ] <queueName>
	azq scan [ -F configFile ] [ -e environment ] ( -a | <queuePrefix> )
	azq tp [ -F configFile ] [ -e environment ] <queueName>
	azq profile [ -F configFile ] [ -e environment ] <queueName> [<duration>] [<buffer>]
	azq put [ -F configFile ] [ -e environment ] <queueName> (- | <message>)
	azq peek [ -F configFile ] [ -e environment ] <queueName>
	azq pop [ -F configFile ] [ -e environment ] <queueName>
Arguments:
	queueName    The name(s) of one or more queues
	queuePrefix  A prefix for filtering which queues to show
	duration     A length of time, in seconds [default: 10]
	message		 An arbitrary text message to be placed in the queue
Options:
	-a           All queues
	-e=env       Azure Storage Services account [default: default]
 	-F=file      Alternate configuration file [default: /usr/local/etc/azq/config]
	-h, --help   Show this screen.
	--version    Show version.
The most commonly used commands are:
	count        Prints the number of messages in one or more queues
	scan         Lists the queues in a storage container
	tp           How long for one message to traverse the queue
	profile      Moving average of queue depth and throughput over time
`
	version string = "EngineRoom 0.3.1"
)

func main() {
	dict := parse(usage, version)
	doIt(dict)
}

func doIt(dict map[string]interface{}) {

	// load config
	configFile := dict["-F"].(string)
	environment := dict["-e"].(string)

	err := engineroom.LoadAzureConfig(configFile, environment)
	if err != nil {
		fmt.Printf("Failed to load config: %s\n", err)
		return
	}

	if dict["count"].(bool) {
		queueName := dict["<queueName>"].(string)
		if queueName == "" {
			os.Exit(1)
		}
		var queueNames []string
		queueNames = append(queueNames, queueName)
		engineroom.Fetch(queueNames, false)
	}

	if dict["scan"].(bool) {
		queuePrefix := ""
		if !dict["-a"].(bool) {
			queuePrefix = dict["<queuePrefix>"].(string)
		}
		engineroom.Scan(queuePrefix)
	}

	if dict["tp"].(bool) {
		queueName := dict["<queueName>"].(string)
		engineroom.MeasureThroughput(queueName)
	}

	if dict["profile"].(bool) {
		queueName := dict["<queueName>"].(string)
		durationArg, ok := dict["<duration>"].(string)
		if !ok {
			durationArg = "5s"
		}

		bufferArg, ok := dict["<buffer>"].(string)
		if !ok {
			bufferArg = "10"
		}

		dur, err := time.ParseDuration(durationArg)
		if err != nil {
			fmt.Printf("Error parsing duration (expected duration e.g. 5m): %s\n", durationArg)
		}
		buf, err := strconv.Atoi(bufferArg)
		if err != nil {
			fmt.Printf("Error parsing buffer (expected int): %s\n", bufferArg)
		}

		engineroom.Profile(queueName, dur, buf)
	}

	if dict["put"].(bool) {
		var message string
		msgParam, paramOk := dict["<message>"].(string)
		_, stdinOk := dict["-"].(bool)
		if paramOk {
			message = msgParam
		} else if stdinOk {
			var err error
			message, err = readStdin()
			if err != nil {
				fmt.Printf("Error reading from Stdin: %s\n", err)
				return
			}
		}

		queueName := dict["<queueName>"].(string)
		engineroom.Put(queueName, message)
	}

	if dict["pop"].(bool) {
		queueName := dict["<queueName>"].(string)
		engineroom.Pop(queueName)
	}

	if dict["peek"].(bool) {
		queueName := dict["<queueName>"].(string)
		engineroom.Peek(queueName)
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

func readStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", errors.New("Nothing on Stdin")
	}

	bytes, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		return "", err
	}

	// Strip trailing newline
	if bytes[len(bytes)-1] == '\n' {
		if bytes[len(bytes)-2] == '\r' {
			return string(bytes[:len(bytes)-2]), nil
		}
		return string(bytes[:len(bytes)-1]), nil
	}

	return string(bytes[:]), nil

}
