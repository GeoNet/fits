package dapperlib

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

const RECORDS_PER_BATCH = 500 // AWS Firehose limit

type SendClient struct {
	firehose *firehose.Client
	fhStream string
}

// NewSendClient returns a SendClient struct that wraps a Firehose client using the default AWS credentials chain.
// This consults (in order) environment vars, config files, EC2 and ECS roles.
// It is an error if the AWS_REGION environment variable is not set.
// Requests with recoverable errors will be retried with the default retrier.
func NewSendClient(fhStream string) (*SendClient, error) {
	cfg, err := getConfig()
	if err != nil {
		return &SendClient{}, err
	}
	return &SendClient{firehose: firehose.NewFromConfig(cfg), fhStream: fhStream}, nil
}

// NewSendClientWithConfig returns a SendClient struct that wraps a Firehose client using
// a provided AWS Config.
func NewSendClientWithConfig(fhStream string, cfg aws.Config) *SendClient {
	return &SendClient{firehose: firehose.NewFromConfig(cfg), fhStream: fhStream}
}

// getConfig returns the default AWS Config struct.
func getConfig() (aws.Config, error) {
	if os.Getenv("AWS_REGION") == "" {
		return aws.Config{}, errors.New("AWS_REGION is not set")
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// Send sends a list of Dapper Records to the AWS Firehose.
func (sc SendClient) Send(data []Record) error {

	records := make([]types.Record, 0)

	for _, result := range data {
		d := RecordToCSV(result)
		records = append(records, types.Record{Data: []byte(d)})
	}

	// PutRecordBatch has a maximum number per batch so we'll split them
	pos := 0
	for {
		if pos >= len(records) {
			break
		}

		recordsBatchInput := firehose.PutRecordBatchInput{
			DeliveryStreamName: &sc.fhStream,
		}

		if pos+RECORDS_PER_BATCH >= len(records) {
			recordsBatchInput.Records = records[pos:]
		} else {
			recordsBatchInput.Records = records[pos:(pos + RECORDS_PER_BATCH)]
		}
		_, err := sc.firehose.PutRecordBatch(context.TODO(), &recordsBatchInput)
		if err != nil {
			return fmt.Errorf("failed to put records: %v", err)
		}
		pos += RECORDS_PER_BATCH
	}

	return nil
}
