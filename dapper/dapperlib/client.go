package dapperlib

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

const RECORDS_PER_BATCH = 500 // AWS Firehose limit

type SendClient struct {
	firehose *firehose.Firehose
	fhStream string
}

func NewSendClient(fhStream string) (*SendClient, error) {
	sc := &SendClient{}

	sc.fhStream = fhStream

	sess, err := session.NewSession()
	if err != nil {
		return sc, fmt.Errorf("failed to create AWS session: %v", err)
	}
	sc.firehose = firehose.New(sess)

	return sc, nil
}

func NewSendClientWithSession(fhStream string, sess *session.Session) (*SendClient, error) {
	sc := &SendClient{}

	sc.fhStream = fhStream

	sc.firehose = firehose.New(sess)

	return sc, nil
}

func (sc SendClient) Send(data []Record) error {
	recordsBatchInput := &firehose.PutRecordBatchInput{}
	recordsBatchInput = recordsBatchInput.SetDeliveryStreamName(sc.fhStream)
	records := []*firehose.Record{}

	for _, result := range data {
		d := RecordToCSV(result)
		records = append(records, &firehose.Record{Data: []byte(d)})
	}

	// PutRecordBatch has a maximum number per batch so we'll split them
	pos := 0
	for {
		if pos >= len(records) {
			break
		}

		if pos+RECORDS_PER_BATCH >= len(records) {
			recordsBatchInput.SetRecords(records[pos:])
		} else {
			recordsBatchInput.SetRecords(records[pos:(pos + RECORDS_PER_BATCH)])
		}

		_, err := sc.firehose.PutRecordBatch(recordsBatchInput)
		if err != nil {
			return fmt.Errorf("failed to put records: %v", err)
		}
		pos += RECORDS_PER_BATCH
	}

	return nil
}
