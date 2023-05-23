<p align="center">
<img height=70px src="../../../docs/images/ContinuumNewLogo.png">
<img height=100px src="../../../docs/images/Go-Logo_Blue.png">
</p>

# Email Service

This is a thin wrapper over AWS SES Managed Service. It primarily should be initialized and invoked from within the microservice to send email notifications to any users using AWS SES emails/domains.

## Prerequisites

- The email addresses to be used in from/to should be configured in AWS SES. For more information, refer [How to verify email/domains addresses](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/verify-email-addresses.html)
- Bounce back email configurations should also be used, but it could be optional based on the type of notifications.
- This wrapper loads AWS configurations from local instance config files in development environment
- However, In Production, AWS recommends configuring AWS EC2 roles ( instead of AWS IAM credentials hardcoded ) configured in EC2 instances of microservices.
- The EC2 roles should give full access to SES managed service actions. For more information, refer
[AWS GoLang SDK](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/examples-send-using-sdk.html) and [AWS Shared Credentals File](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/create-shared-credentials-file.html).

## Licensing

- It uses [AWS GoLang SDK](https://github.com/aws/aws-sdk-go-v2) under [Apache 2.0 License](https://github.com/aws/aws-sdk-go-v2/blob/main/LICENSE.txt)

## Development Env

- Install Go 1.12 or above. It only supports Go Modules for dependency management.
- Setup AWS CLI and configure AWS IAM creds or IAM role for SES

## Enhancements

- It can be further enhanced to suit various needs. For eg, it exposes interfaces in following file which can be implemented for own needs and contribution here. Eg, `AWSEmailClient` implements `EmailClient` interface.

```go
interfaces.go 
```

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/notifications/email"
```

## Example

- Populate Email Content. Methods can be *chained* like below:

```go
    content := &EmailContent{}
    content.AddRecipientEmail(Recipient).
            AddRecipientEmail("nitin.kothari@continuum.net")
    // One or more receipents email addresses
    content.CharSet = DefaultCharset   //Defaults to UTF-8
    content.TextBody = TextBody //Text body in case non-html content
    content.Sender = Sender     //From email address
    content.Subject = Subject   //Represents Subject in the email
    content.HTMLBody = HtmlBody //Represents html content (optional)
    
```

- **Optionally**, we can also replace placeholders inside HTML body. For eg, you might want to replace placeholder `{link}` inside your html body with actul links. In that case, you just add body key like below:

```go
    content.AddBodyKey(LINK, "https://control.itsupport247.net/")
```

where `LINK` is of type `PlaceholderKey` and the value is `{link}`. It automatically replaces the keys while sending emails by calling various replacement functions.

- Initialize AWS Email Client as below
- Region is mandatory in client. It defaults to `us-east-1` if none provided.
- `AWSEmailClient` wraps AWS SES client

```go
    awsClient := &AWSEmailClient{"us-east-1", nil, nil}
    awsClient.Configure()       //Required.Configures AWS SES client underneath
```

- **Alternatively**, it can also be initialized with logger, it accepts only `src/runtime/logger` which is our default logger at the moment.

```go
    awsClient := &AWSEmailClient{"us-east-1", nil, log}
    awsClient.Configure()
```

- Send Email

```go
    emailOutput := awsClient.SendEmail("transactionId",content, nil)
```

- Or, Send Email with (optional) `ErrorHandler` to provide custom logic to handle errors from AWS.

```go
    emailOutput := awsClient.SendEmail(content, awsClient.AWSErrorHandler)
```

- ErrorHandler is of type `ErrorHandler func()`
- Email Output is of type `EmailOutput`
- If the email is sent successfully via SES, Email Output will be:

```go
    fmt.Println("Message: " + *emailOutput.MessageId)
```

- An error is populated in output in case of error:

```go
    fmt.Println("Message: " + *emailOutput.Error)
```

### Contribution

If you want to contribute, please reach out to contributors
