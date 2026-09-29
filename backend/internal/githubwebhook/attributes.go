package githubwebhook

import "github.com/aws/aws-sdk-go-v2/service/sqs/types"

// Set by the API Gateway → SQS integration's MessageAttributes mapping
// (runbook §3); renaming it here also requires changing that mapping.
const attrGitHubEvent = "x-github-event"

func githubEventType(msg types.Message) string {
	if attr, ok := msg.MessageAttributes[attrGitHubEvent]; ok && attr.StringValue != nil {
		return *attr.StringValue
	}
	return ""
}
