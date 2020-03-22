package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-queue-go/azqueue"
)

// RunInfo stores the workflow data that will be sent to the Azure Queue
type RunInfo struct {
	Workflow   string
	RunID      string
	RunNumber  string
	Actor      string
	Repository string
	EventName  string
	EventPath  string
	Sha        string
	Ref        string
	HeadRef    string
	BaseRef    string
	JobStatus  string
}

type connectionInfo struct {
	storageAccountName string
	storageAccountKey  string
	queueName          string
}

func main() {
	runInfo := RunInfo{
		Workflow:   os.Getenv("GITHUB_WORKFLOW"),
		RunID:      os.Getenv("GITHUB_RUN_ID"),
		RunNumber:  os.Getenv("GITHUB_RUN_NUMBER"),
		Actor:      os.Getenv("GITHUB_ACTOR"),
		Repository: os.Getenv("GITHUB_REPOSITORY"),
		EventName:  os.Getenv("GITHUB_EVENT_NAME"),
		EventPath:  os.Getenv("GITHUB_EVENT_PATH"),
		Sha:        os.Getenv("GITHUB_SHA"),
		Ref:        os.Getenv("GITHUB_REF"),
		HeadRef:    os.Getenv("GITHUB_HEAD_REF"),
		BaseRef:    os.Getenv("GITHUB_BASE_REF"),
		JobStatus:  os.Getenv("INPUT_JOB_STATUS"),
	}

	runInfoJSON, err := json.Marshal(runInfo)
	if err != nil {
		log.Fatal(err)
	}

	connectionInfo := connectionInfo{
		storageAccountName: os.Getenv("INPUT_STORAGE_ACCOUNT_NAME"),
		storageAccountKey:  os.Getenv("INPUT_STORAGE_ACCOUNT_KEY"),
		queueName:          os.Getenv("INPUT_QUEUE_NAME"),
	}

	queueURL, err := createQueueURL(
		connectionInfo.storageAccountName,
		connectionInfo.storageAccountKey,
		connectionInfo.queueName,
	)
	if err != nil {
		log.Fatalf("Error creating Queue URL: %s", err)
	}

	msgURL := queueURL.NewMessagesURL()

	sendMessage(context.Background(), msgURL, string(runInfoJSON))
	fmt.Printf("::set-output name=data-sent::%s\n", runInfoJSON)
}

func sendMessage(ctx context.Context, msgURL azqueue.MessagesURL, messageToSend string) {
	_, err := msgURL.Enqueue(context.Background(), messageToSend, 0, 0)
	if err != nil {
		log.Fatalf("Error adding message to queue: ", err)
	}

	log.Printf("Added message \"%v\" to the queue", messageToSend)
}

func createQueueURL(storageAccountName, storageAccountKey, queueName string) (azqueue.QueueURL, error) {
	log.Printf("Creating queue '%s'", queueName)
	url, err := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net/%s", storageAccountName, queueName))
	if err != nil {
		log.Fatalf("Error parsing url: %s", err)
	}

	credential, err := azqueue.NewSharedKeyCredential(storageAccountName, storageAccountKey)
	if err != nil {
		log.Fatalf("Error creating credentials: %s", err)
	}

	return azqueue.NewQueueURL(*url, azqueue.NewPipeline(credential, azqueue.PipelineOptions{})), err
}
