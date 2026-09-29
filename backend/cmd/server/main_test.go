package main

import (
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/githubwebhook"
)

func TestFormatPushText(t *testing.T) {
	cases := []struct {
		name      string
		evt       githubwebhook.WorkflowRunEvent
		wantTitle string
		wantBody  string
	}{
		{
			name: "start event with PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "feature/x",
				Conclusion: "queued", PRNumber: 42, PRTitle: "feat(x): hello",
			},
			wantTitle: "CI 시작 — dlddu/pocket-aide #42",
			wantBody:  "feat(x): hello",
		},
		{
			name: "start event in_progress without PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "main",
				Conclusion: "in_progress",
			},
			wantTitle: "dlddu/pocket-aide — in_progress",
			wantBody:  "CI on main",
		},
		{
			name: "completed success with PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "feature/x",
				Conclusion: "success", PRNumber: 7, PRTitle: "fix: thing",
			},
			wantTitle: "CI 통과 — dlddu/pocket-aide #7",
			wantBody:  "fix: thing",
		},
		{
			name: "completed failure without PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "main",
				Conclusion: "failure",
			},
			wantTitle: "dlddu/pocket-aide — failure",
			wantBody:  "CI on main",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			title, body := formatPushText(tc.evt)
			if title != tc.wantTitle {
				t.Errorf("title: got %q want %q", title, tc.wantTitle)
			}
			if body != tc.wantBody {
				t.Errorf("body: got %q want %q", body, tc.wantBody)
			}
		})
	}
}

func TestShouldPush(t *testing.T) {
	cases := []struct {
		name string
		evt  githubwebhook.WorkflowRunEvent
		want bool
	}{
		{"start queued is history only", githubwebhook.WorkflowRunEvent{Conclusion: "queued"}, false},
		{"start in_progress is history only", githubwebhook.WorkflowRunEvent{Conclusion: "in_progress"}, false},
		{"completed success", githubwebhook.WorkflowRunEvent{Conclusion: "success", Completed: true}, true},
		{"completed failure", githubwebhook.WorkflowRunEvent{Conclusion: "failure", Completed: true}, true},
		{"completed cancelled", githubwebhook.WorkflowRunEvent{Conclusion: "cancelled", Completed: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldPush(tc.evt); got != tc.want {
				t.Errorf("shouldPush: got %v want %v", got, tc.want)
			}
		})
	}
}
