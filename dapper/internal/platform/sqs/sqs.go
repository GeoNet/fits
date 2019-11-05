// Package sqs is for messaging with AWS SQS.
package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"os"
)

type Raw struct {
	Body          string
	ReceiptHandle string
}

type SQS struct {
	client *sqs.SQS
}

// New returns an SQS using the default AWS credentials chain.
// This consults (in order) environment vars, config files, ec2 and ecs roles.
// It is an error if the AWS_REGION environment variable is not set.
// Requests with recoverable errors will be retried with the default retrier.
func New() (SQS, error) {
	if os.Getenv("AWS_REGION") == "" {
		return SQS{}, errors.New("AWS_REGION is not set")
	}

	s, err := session.NewSession()
	if err != nil {
		return SQS{}, errors.WithStack(err)
	}

	return SQS{client: sqs.New(s)}, nil
}

func NewWithCreds(creds *credentials.Credentials) (SQS, error) {
	if os.Getenv("AWS_REGION") == "" {
		return SQS{}, errors.New("AWS_REGION is not set")
	}

	s, err := session.NewSession()
	if err != nil {
		return SQS{}, errors.WithStack(err)
	}

	return SQS{client: sqs.New(s, &aws.Config{Credentials: creds})}, nil
}

// Receive a raw message or error from the queue.
// After a successful receive the message will be in flight
// until it is either deleted or the visibility timeout expires
// (at which point it is available for redelivery).
//
// Applications should be able to handle duplicate or out of order messages.
// and should back off on Receive error.
func (s *SQS) Receive(queueURL string, visibilityTimeout int64) (Raw, error) {
	input := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: aws.Int64(int64(1)),
		VisibilityTimeout:   aws.Int64(visibilityTimeout),
		WaitTimeSeconds:     aws.Int64(int64(20)),
	}

	for {
		r, err := s.client.ReceiveMessage(&input)
		if err != nil {
			return Raw{}, errors.WithStack(err)
		}

		switch {
		case r == nil || len(r.Messages) == 0:
			// no message received
			continue
		case len(r.Messages) == 1:
			raw := r.Messages[0]

			if raw == nil {
				return Raw{}, errors.New("got nil message pointer")
			}

			m := Raw{
				Body:          aws.StringValue(raw.Body),
				ReceiptHandle: aws.StringValue(raw.ReceiptHandle),
			}
			return m, nil
		case len(r.Messages) > 1:
			return Raw{}, fmt.Errorf("received more than 1 message: %d", len(r.Messages))
		}
	}
}

// Delete deletes the message referred to by receiptHandle from the queue.
func (s *SQS) Delete(queueURL, receiptHandle string) error {
	params := sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	_, err := s.client.DeleteMessage(&params)

	return err
}

// Send sends the message body to the SQS queue referred to by queueURL.
func (s *SQS) Send(queueURL string, body string) error {
	params := sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(body),
	}

	_, err := s.client.SendMessage(&params)

	return err
}
