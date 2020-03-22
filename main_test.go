package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

var msgURL azqueue.MessagesURL
var messageToSend = "My Test Message"

func TestMain(m *testing.M) {
	// Create Azure Resource Group
	groupName := fmt.Sprintf("gh-action-push-result-to-azure-queue-%s", getRandomString(5))
	_, err := createGroup(context.Background(), groupName)
	if err != nil {
		log.Fatalf("Failed to create group. Error: %s", err)
	}

	// Create Azure Storage Account
	storageAccountName := fmt.Sprintf("ghactionpushresult%s", getRandomString(6))
	_, err = createStorageAccount(context.Background(), storageAccountName, groupName)
	if err != nil {
		log.Fatalf("Failed to create storage account. Error: %s", err)
	}

	// Create Queue inside the Storage Account
	queueName := fmt.Sprintf("ghactionpushresult%s", getRandomString(6))
	connectionString := *getStorageAccountConnectionString(context.Background(), storageAccountName, groupName)
	queueURL, err := createQueueURL(storageAccountName, connectionString, queueName)
	err = createQueue(context.Background(), queueURL)
	if err != nil {
		log.Fatalf("Failed to create queue. Error: %s", err)
	}

	// Create MessageURL for tests
	msgURL = queueURL.NewMessagesURL()

	// Run tests
	exitVal := m.Run()

	// Delete  Azure Resource Group
	_, err = deleteGroup(context.Background(), groupName)
	if err != nil {
		log.Fatalf("Failed to delete group. Error: %s", err)
	}

	os.Exit(exitVal)
}

func TestSend(t *testing.T) {
	sendMessage(context.Background(), msgURL, messageToSend)
}

func TestReceive(t *testing.T) {
	var recvMsg string
	dequeueResp, err := msgURL.Dequeue(context.Background(), 1, 10*time.Second)

	if err != nil {
		log.Fatalf("Error dequeueing message: %s", err)
	}

	for i := int32(0); i < dequeueResp.NumMessages(); i++ {
		msg := dequeueResp.Message(i)
		recvMsg = msg.Text
		log.Printf("Deleting %v: {%v}", i, msg.Text)

		msgIDURL := msgURL.NewMessageIDURL(msg.ID)

		_, err = msgIDURL.Delete(context.Background(), msg.PopReceipt)
		if err != nil {
			log.Fatalf("Error deleting message: %s", err)
		}
	}
	if recvMsg == messageToSend {
		t.Logf("Received message = %s; want %s", recvMsg, messageToSend)
		return
	}
	t.Errorf("Received message = %s; want %s", recvMsg, messageToSend)
}

func getGroupsClient() resources.GroupsClient {
	groupsClient := resources.NewGroupsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err == nil {
		groupsClient.Authorizer = authorizer
	}
	return groupsClient
}

func createGroup(ctx context.Context, groupName string) (resources.Group, error) {
	groupsClient := getGroupsClient()
	log.Println(fmt.Sprintf("creating resource group '%s' on location: %v", groupName, os.Getenv("AZURE_LOCATION")))
	return groupsClient.CreateOrUpdate(
		ctx,
		groupName,
		resources.Group{
			Location: to.StringPtr(os.Getenv("AZURE_LOCATION")),
		})
}

func deleteGroup(ctx context.Context, groupName string) (result resources.GroupsDeleteFuture, err error) {
	log.Printf("Deleting  resource group '%s'", groupName)
	groupsClient := getGroupsClient()
	return groupsClient.Delete(ctx, groupName)
}

func getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err == nil {
		storageAccountsClient.Authorizer = authorizer
	}

	return storageAccountsClient
}

func createStorageAccount(ctx context.Context, storageAccountName, accountGroupName string) (storage.Account, error) {
	var s storage.Account
	storageAccountsClient := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(storageAccountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		return s, fmt.Errorf("storage account check-name-availability failed: %+v", err)
	}

	if !*result.NameAvailable {
		return s, fmt.Errorf(
			"storage account name [%s] not available: %v\nserver message: %v",
			storageAccountName, err, *result.Message)
	}
	log.Printf("Creating storage account '%s'", storageAccountName)
	future, err := storageAccountsClient.Create(
		ctx,
		accountGroupName,
		storageAccountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Kind:                              storage.Storage,
			Location:                          to.StringPtr(os.Getenv("AZURE_LOCATION")),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})

	if err != nil {
		return s, fmt.Errorf("failed to start creating storage account: %+v", err)
	}

	err = future.WaitForCompletionRef(ctx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("failed to finish creating storage account: %+v", err)
	}

	return future.Result(storageAccountsClient)
}

func getStorageAccountConnectionString(ctx context.Context, accountName, accountGroupName string) *string {
	storageAccountsClient := getStorageAccountsClient()
	accountListKeysResult, err := storageAccountsClient.ListKeys(ctx, accountGroupName, accountName, storage.Kerb)
	if err != nil {
		fmt.Printf("failed to list keys for storage account: %s\n", err)
	}
	return (*accountListKeysResult.Keys)[0].Value

}

func createQueue(ctx context.Context, queueURL azqueue.QueueURL) error {
	_, err := queueURL.Create(ctx, azqueue.Metadata{})
	return err
}

func getRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
