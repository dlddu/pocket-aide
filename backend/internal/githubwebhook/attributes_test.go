package githubwebhook

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func TestGitHubEventType_Present(t *testing.T) {
	msg := types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			attrGitHubEvent: {DataType: aws.String("String"), StringValue: aws.String("workflow_run")},
		},
	}
	if got := githubEventType(msg); got != "workflow_run" {
		t.Errorf("got %q want %q", got, "workflow_run")
	}
}

func TestGitHubEventType_AbsentAttribute(t *testing.T) {
	// No message attributes at all — e.g. a direct SendMessage from outside the
	// API Gateway integration. Not an error; the caller decides to drop it.
	if got := githubEventType(types.Message{}); got != "" {
		t.Errorf("got %q want empty", got)
	}
}

func TestGitHubEventType_NilStringValue(t *testing.T) {
	// Attribute present but with no StringValue (e.g. a Binary-typed value).
	msg := types.Message{
		MessageAttributes: map[string]types.MessageAttributeValue{
			attrGitHubEvent: {DataType: aws.String("Binary")},
		},
	}
	if got := githubEventType(msg); got != "" {
		t.Errorf("got %q want empty", got)
	}
}
