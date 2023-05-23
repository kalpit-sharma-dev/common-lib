package main

import (
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/notifications/email"
)

func main() {
	var content = &email.EmailContent{Sender: "firstName", TextBody: "This is sample text for email body", Subject: "Example email"}
	content.AddRecipientEmail("firstName.lastName@kksharmadev.com")
	content.AddCCEmail("ccFirstName.ccLastName@kksharmadev.com")
	content.AddTemplateKey("template_key1", "template_value1")
	content.AddBodyKey("body_key1", "body_value1")
	var awsClient = &email.AWSEmailClient{}
	_ = awsClient.Configure()
	_, _ = awsClient.SendEmail("transactionId", content, awsClient.AWSErrorHandler)
}
