package email

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/notifications/email/mocks"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	sender = "nitin.kothari@gmail.com"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	recipient = "Recipient@gmail.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	subject = "Important Notification from google"

	htmlTemplate = "<p>Welcome to ITSupport247. Your account has been successfully created. {emailBody} </p>"
	// The HTML body for the email.
	htmlBody = "<p>Login to the ITSupport247 Portal using the following information:</p>" +
		"<p>URL: {link}</p><p>User Name: {fullName}</p>" +
		"<p>Password: Click <a href=&quot;{link}&quot;>here</a> to create a new password for your account.</p>"

	//The email body for recipients with non-HTML email clients.
	textBody = "Welcome to ITS Portal. To log in to your ITSupport247 portal account, you will need to create a new password. Click on the link below to open a secure browser window and reset your password now."
)

func TestAWSEmailClient(t *testing.T) {
	content := &EmailContent{}
	content.AddRecipientEmail(recipient).AddRecipientEmail("nitin.kothari@gmail.com")
	content.CharSet = DefaultCharset
	content.TextBody = textBody
	content.Sender = sender
	content.Subject = subject
	content.HTMLBody = htmlBody
	content.HTMLTemplate = htmlTemplate
	content.AddCCEmail("dummy@dummy.com")
	content.AddTemplateKey(EMAIL_BODY, htmlBody)
	content.AddBodyKey(FULL_NAME, "nkothari")

	ctrl := gomock.NewController(t)
	mockSESClient := mocks.NewMockSesiface(ctrl)
	awsClient := &AWSEmailClient{"us-east-1", mockSESClient, nil, context.Background()}

	dummyId := "DummyID-1111"
	mockSESClient.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(&ses.SendEmailOutput{MessageId: &dummyId}, nil).AnyTimes()

	emailOutput, _ := awsClient.SendEmail("", content, nil)
	assert.Equal(t, &dummyId, emailOutput.MessageId)
}

func TestSES_ErrCodeConfigurationSet(t *testing.T) {
	content := &EmailContent{}
	content.AddRecipientEmail(recipient).AddRecipientEmail("nitin.kothari@gmail.com")
	content.Sender = sender

	ctrl := gomock.NewController(t)
	mockSESClient := mocks.NewMockSesiface(ctrl)
	awsClient := &AWSEmailClient{"us-east-1", mockSESClient, nil, context.Background()}

	mockSESClient.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(nil, errors.New("ErrConfigurationSetNotSet")).AnyTimes()
	_, err := awsClient.SendEmail("", content, awsClient.AWSErrorHandler)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ErrConfigurationSetNotSet")
}

func TestSES_InvalidReceipients(t *testing.T) {
	content := &EmailContent{}
	content.CharSet = DefaultCharset
	content.TextBody = textBody
	content.Sender = sender
	content.Subject = subject
	content.HTMLBody = htmlBody

	ctrl := gomock.NewController(t)
	mockSESClient := mocks.NewMockSesiface(ctrl)
	awsClient := &AWSEmailClient{"us-east-1", mockSESClient, nil, context.Background()}

	dummyId := "DummyID-1111"
	mockSESClient.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(&ses.SendEmailOutput{MessageId: &dummyId}, nil).AnyTimes()
	_, err := awsClient.SendEmail("", content, awsClient.AWSErrorHandler)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "please provide valid recepient(s)")
}

func TestSES_Configure(t *testing.T) {
	awsClient := &AWSEmailClient{Region: "us-west-3"}

	err := awsClient.Configure() //under test

	assert.Nil(t, err)
	assert.NotNil(t, awsClient.context)
	assert.Equal(t, "us-west-3", awsClient.Region)
	assert.NotNil(t, awsClient.SESClient)
}

func TestSES_ConfigureWithContext(t *testing.T) {
	awsClient := &AWSEmailClient{Region: "us-west-3"}

	ctx := context.Background()

	err := awsClient.ConfigureWithContext(ctx) //under test

	assert.Nil(t, err)
	assert.NotNil(t, ctx)
	assert.Equal(t, ctx, awsClient.context)
	assert.Equal(t, "us-west-3", awsClient.Region)
	assert.NotNil(t, awsClient.SESClient)
}
