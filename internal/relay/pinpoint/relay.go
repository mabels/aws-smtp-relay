package pinpoint

import (
	"context"
	"net"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/pinpointemail"
	pinpointemailtypes "github.com/aws/aws-sdk-go-v2/service/pinpointemail/types"
	"github.com/blueimp/aws-smtp-relay/internal"
	"github.com/blueimp/aws-smtp-relay/internal/relay/filter"
)

type PinpointEmailClient interface {
	SendEmail(context.Context, *pinpointemail.SendEmailInput, ...func(*pinpointemail.Options)) (*pinpointemail.SendEmailOutput, error)
}

// Client implements the Relay interface.
type Client struct {
	pinpointClient  PinpointEmailClient
	setName         *string
	allowFromRegExp *regexp.Regexp
	denyToRegExp    *regexp.Regexp
}

// Send uses the given Pinpoint API to send email data
func (c Client) Send(
	origin net.Addr,
	from string,
	to []string,
	data []byte,
) error {
	allowedRecipients, deniedRecipients, err := filter.FilterAddresses(
		from,
		to,
		c.allowFromRegExp,
		c.denyToRegExp,
	)
	if err != nil {
		internal.Log(origin, from, deniedRecipients, err)
	}
	if len(allowedRecipients) > 0 {
		_, err := c.pinpointClient.SendEmail(context.Background(), &pinpointemail.SendEmailInput{
			Content:                        &pinpointemailtypes.EmailContent{Raw: &pinpointemailtypes.RawMessage{Data: data}},
			Destination:                    &pinpointemailtypes.Destination{ToAddresses: allowedRecipients},
			ConfigurationSetName:           c.setName,
			EmailTags:                      []pinpointemailtypes.MessageTag{},
			FeedbackForwardingEmailAddress: new(string),
			FromEmailAddress:               &from,
			ReplyToAddresses:               to,
		})
		internal.Log(origin, from, allowedRecipients, err)
		if err != nil {
			return err
		}
	}
	return err
}

// New creates a new client with a session.
func New(
	configurationSetName *string,
	allowFromRegExp *regexp.Regexp,
	denyToRegExp *regexp.Regexp,
) Client {
	return Client{
		pinpointClient:  pinpointemail.New(pinpointemail.Options{}),
		setName:         configurationSetName,
		allowFromRegExp: allowFromRegExp,
		denyToRegExp:    denyToRegExp,
	}
}
