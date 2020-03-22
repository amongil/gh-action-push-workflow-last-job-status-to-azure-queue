# Push workflow last job status to Azure Queue action

This action sends last job's in the github workflow status, along with some metadata, to an Azure Queue. 

## Inputs

### `STORAGE_ACCOUNT_NAME`

**Required** Storage account name where the Azure queue is located.

### `STORAGE_ACCOUNT_KEY`

**Required** Storage account key in order to authenticate to.

### `QUEUE_NAME`

**Required** Name of the queue where the messages are going to be sent to.

### `JOB_STATUS`

**Required** Exit status of last job.

## Outputs

### `data-sent`

The json data sent to the Azure Queue.

## Example usage

```
uses: amongil/gh-action-push-result-to-azure-queue@v0.1.0
with:
  STORAGE_ACCOUNT_NAME: mystorageaccount
  STORAGE_ACCOUNT_KEY: ${{ secrets.AZURE_ACTION_STORAGE_ACCOUNT_KEY}}
  QUEUE_NAME: myqueue
  JOB_STATUS: ${{ job.status }}
```

## Example data sent
```
{
   "Workflow":"myworkflow",
   "RunID":"60849633",
   "RunNumber":"14",
   "Actor":"amongil",
   "Repository":"amongil/github-action-poc",
   "EventName":"push",
   "EventPath":"/github/workflow/event.json",
   "Sha":"c2f56fb797480f470d4a3e1a564de4c6583f0481",
   "Ref":"refs/heads/master",
   "HeadRef":"",
   "BaseRef":"",
   "JobStatus":"Success"
}
```

## Testing

Testing is done manually authenticating through [local auth file](https://docs.microsoft.com/en-us/azure/python/python-sdk-azure-authenticate#mgmt-auth-file). Be careful as you might incur in costs by testing.

For testing, simply run the following command:

```
go test
```

Tests will, in the given subscription:
- Create a storage account
- Retrieve the storage account key
- Create a queue
- Send a predefined message to the queue
- Pop the queue and check that the received message equals to the one sent before

Example output of the `go test` command:

```
2020/03/21 15:54:26 creating resource group 'gh-action-push-result-to-azure-queue-aenog' on location: northeurope
2020/03/21 15:54:29 Creating storage account 'ghactionpushresultnrwwn1'
2020/03/21 15:54:49 Creating queue 'ghactionpushresulto56z4s'
2020/03/21 15:54:50 Added message "My Test Message" to the queue
2020/03/21 15:54:50 Deleting 0: {My Test Message}
PASS
2020/03/21 15:54:50 Deleting  resource group 'gh-action-push-result-to-azure-queue-aenog'
ok  	amongil/gh-action-push-result-to-azure-queue	26.210s
```
