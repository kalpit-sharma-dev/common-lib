package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/smithy-go"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const (
	DefaultRegion  string         = "us-east-1"
	DefaultCharset string         = "UTF-8"
	FULL_NAME      PlaceholderKey = "{fullName}"
	EMAIL_BODY     PlaceholderKey = "{emailBody}"
	ButtonLabel    PlaceholderKey = "{buttonLabel}"
	LINK           PlaceholderKey = "{link}"
)

func (ec *EmailContent) AddRecipientEmail(recipient string) *EmailContent {
	if recipient != "" {
		ec.ToAddresses = append(ec.ToAddresses, recipient)
	}
	return ec
}

func (ec *EmailContent) AddCCEmail(ccaddr string) *EmailContent {
	if ccaddr != "" {
		ec.CCAddresses = append(ec.CCAddresses, ccaddr)
	}
	return ec
}
func (ec *EmailContent) AddTemplateKey(key PlaceholderKey, value string) *EmailContent {
	if ec.ContentKeyValue == nil {
		ec.ContentKeyValue = make(map[PlaceholderKey]string)
	}
	if key != "" {
		ec.ContentKeyValue[key] = value
	}

	return ec
}
func (ec *EmailContent) AddBodyKey(key PlaceholderKey, value string) *EmailContent {
	if ec.BodyKeyValue == nil {
		ec.BodyKeyValue = make(map[PlaceholderKey]string)
	}
	if key != "" {
		ec.BodyKeyValue[key] = value
	}
	return ec
}

func (ec *EmailContent) replaceTemplate() {
	if len(ec.HTMLTemplate) > 0 {
		for k, v := range ec.ContentKeyValue {
			ec.HTMLTemplate = strings.ReplaceAll(ec.HTMLTemplate, string(k), v)
		}
	}
}

func (ec *EmailContent) replaceBody() {
	if len(ec.HTMLBody) > 0 {
		for k, v := range ec.BodyKeyValue {
			ec.HTMLBody = strings.ReplaceAll(ec.HTMLBody, string(k), v)
		}
		ec.HTMLTemplate = strings.ReplaceAll(ec.HTMLTemplate, string(EMAIL_BODY), ec.HTMLBody)
	}
}

func (ec *EmailContent) validate() error {
	if len(ec.ToAddresses) <= 0 {
		return errors.New("please provide valid recepient(s)")
	}
	if len(ec.Sender) == 0 {
		return errors.New("please provide valid source")
	}
	return nil
}

// Composition of SES Client
type AWSEmailClient struct {
	Region    string
	SESClient Sesiface
	Log       logger.Log
	context   context.Context
}

// Configure AWS SES Client. This method will create a new context and override any others that have been passed in.
func (ases *AWSEmailClient) Configure() error {
	return ases.ConfigureWithContext(context.TODO())
}

// Configure AWS SES Client. This method is the same as Configure, it just allows the context to be passed in.
func (ases *AWSEmailClient) ConfigureWithContext(ctx context.Context) error {
	ases.context = ctx

	if ases.Region == "" {
		ases.Region = DefaultRegion
	}

	con, err := config.LoadDefaultConfig(ases.context, config.WithRegion(ases.Region))
	if err != nil {
		return fmt.Errorf("could not create aws config object;%w", err)
	}

	ases.SESClient = ses.NewFromConfig(con)
	return nil
}

func (ases *AWSEmailClient) buildMessage(ec *EmailContent) *ses.SendEmailInput {
	ec.replaceBody()
	ec.replaceTemplate()
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			CcAddresses: ec.CCAddresses,
			ToAddresses: ec.ToAddresses,
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String(ec.CharSet),
					Data:    aws.String(ec.HTMLTemplate),
				},
				Text: &types.Content{
					Charset: aws.String(ec.CharSet),
					Data:    aws.String(ec.TextBody),
				},
			},
			Subject: &types.Content{
				Charset: aws.String(ec.CharSet),
				Data:    aws.String(ec.Subject),
			},
		},
		Source: aws.String(ec.Sender),
	}
	return input
}

// Error Handler passed to handle AWS error codes.
func (ases *AWSEmailClient) AWSErrorHandler(transactionID string, err error) {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		ases.logError(transactionID, ae.ErrorCode(), ae.ErrorMessage())
	}
}

func (ases *AWSEmailClient) logError(transactionID string, errorCode string, errorMessage string) {
	if ases.Log != nil {
		ases.Log.Info(transactionID, errorCode, errorMessage)
	} else {
		fmt.Println(transactionID, errorCode, errorMessage)
	}
}

func (ases *AWSEmailClient) SendEmail(transactionID string, ec *EmailContent, errorHandlerCallback ErrorHandler) (*EmailOutput, error) {
	// Attempt to send the email.
	err := ec.validate()
	if err != nil {
		return nil, err
	}
	result, err := ases.SESClient.SendEmail(ases.context, ases.buildMessage(ec))
	if err != nil {
		if errorHandlerCallback != nil {
			errorHandlerCallback(transactionID, err)
		}
		return nil, err
	}

	return &EmailOutput{result.MessageId}, nil
}
