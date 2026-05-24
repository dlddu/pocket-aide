package githubwebhook

import "github.com/aws/aws-sdk-go-v2/service/sqs/types"

// The API Gateway → SQS proxy integration (see the PR-monitor runbook §3)
// forwards GitHub's webhook HTTP headers as SQS message attributes.
// attrGitHubEvent carries the X-GitHub-Event header — the event type
// (workflow_run, push, ping, …) that used to come from the EventBridge
// envelope's detail-type field.
const attrGitHubEvent = "x-github-event"

// githubEventType returns the GitHub event type from the SQS message
// attributes, or "" when the attribute is absent (e.g. a message published by
// something other than the API Gateway integration). Whether a given type is
// one we act on is the caller's decision, so a missing attribute is not an
// error here.
func githubEventType(msg types.Message) string {
	if attr, ok := msg.MessageAttributes[attrGitHubEvent]; ok && attr.StringValue != nil {
		return *attr.StringValue
	}
	return ""
}
