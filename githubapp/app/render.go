package app

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/google/go-github/v64/github"
	"github.com/samber/lo"
)

//go:embed check_run.tmpl
var checkRunTplContent string
var tpl *template.Template

func init() {
	var err error
	tpl, err = template.New("check_run").Parse(checkRunTplContent)
	if err != nil {
		panic(err)
	}
}

type RenderCheckRunOptions struct {
	RunnerJob *api.JobExecution
	RenderJob *RenderJob
}

type RenderJob struct {
	UID   string
	Name  string
	Steps []*RenderStep
}

type RenderStep struct {
	Name string
}

func GenerateCreateCheckRunOptions(baseURL string, sha string, option *RenderCheckRunOptions) (github.CreateCheckRunOptions, error) {
	runnerJob := option.RunnerJob
	checkRunContent, err := generateCheckRunSummary(baseURL, option)
	if err != nil {
		return github.CreateCheckRunOptions{}, errors.WithMessage(err, "failed to render check run")
	}
	name := genCheckRunName(option)
	var title string
	switch {
	case runnerJob.Status.IsPreDispatch():
		title = "Queued"
	case runnerJob.Status.IsRunning():
		title = "Running"
	case runnerJob.Status == api.StatusSucceeded:
		title = "Succeeded"
	case runnerJob.Status == api.StatusFailed:
	}
	status, conclusion := getCheckRunStatus(runnerJob.Status)

	return github.CreateCheckRunOptions{
		Name:       name,
		HeadSHA:    sha,
		Status:     status,
		Conclusion: conclusion,
		Output: &github.CheckRunOutput{
			Title:   lo.ToPtr(fmt.Sprintf(title)),
			Summary: lo.ToPtr(string(checkRunContent)),
		},
	}, nil
}

func GenerateUpdateCheckRunOptions(baseURL string, option *RenderCheckRunOptions) (github.UpdateCheckRunOptions, error) {
	runnerJob := option.RunnerJob
	checkRunContent, err := generateCheckRunSummary(baseURL, option)
	if err != nil {
		return github.UpdateCheckRunOptions{}, errors.WithMessage(err, "failed to render check run")
	}
	name := genCheckRunName(option)
	var title string
	switch {
	case runnerJob.Status.IsPreDispatch():
		title = "Queued"
	case runnerJob.Status.IsRunning():
		title = "Running"
	case runnerJob.Status == api.StatusSucceeded:
		title = "Succeeded"
	case runnerJob.Status == api.StatusFailed:
	}
	status, conclusion := getCheckRunStatus(runnerJob.Status)
	return github.UpdateCheckRunOptions{
		Name:       name,
		Status:     status,
		Conclusion: conclusion,
		Output: &github.CheckRunOutput{
			Title:   lo.ToPtr(fmt.Sprintf(title)),
			Summary: lo.ToPtr(string(checkRunContent)),
		},
	}, nil
}

func genCheckRunName(option *RenderCheckRunOptions) string {
	name := fmt.Sprintf("[XX-CI] %s", option.RenderJob.UID)
	return name
}

func getCheckRunStatus(runnerStatus api.Status) (status *string, conclusion *string) {
	switch {
	case runnerStatus.IsPreDispatch():
		return lo.ToPtr("queued"), nil
	case runnerStatus.IsRunning():
		return lo.ToPtr("in_progress"), nil
	case runnerStatus == api.StatusSucceeded:
		return lo.ToPtr("completed"), lo.ToPtr("success")
	case runnerStatus == api.StatusFailed:
		return lo.ToPtr("completed"), lo.ToPtr("failure")
	default:
		return lo.ToPtr("queued"), nil
	}
}

func generateCheckRunSummary(baseURL string, option *RenderCheckRunOptions) ([]byte, error) {
	output := &bytes.Buffer{}
	runnerJob := option.RunnerJob
	renderData := &renderCheckRunVO{
		JobName:     option.RenderJob.Name,
		StatusEmoji: convertStatus(runnerJob.Status),
	}
	for idx, step := range option.RenderJob.Steps {
		runnerStep := runnerJob.Steps[idx]
		renderData.Steps = append(renderData.Steps, &renderStepVO{
			Name:        step.Name,
			StatusEmoji: convertStatus(runnerStep.Status),
			LogURL:      fmt.Sprintf("%s/log?log_id=%d&log_name=%d", baseURL, runnerJob.ID, idx),
		})
	}
	err := tpl.Execute(output, renderData)
	return output.Bytes(), err
}

type renderCheckRunVO struct {
	JobName     string
	StatusEmoji string
	Steps       []*renderStepVO
}

type renderStepVO struct {
	Name        string
	StatusEmoji string
	LogURL      string
}

func convertStatus(status api.Status) string {
	switch {
	case status.IsPreDispatch():
		return "⏸"
	case status.IsRunning():
		return "▶️"
	case status == api.StatusSucceeded:
		return "✅"
	case status == api.StatusFailed:
		return "❌"
	default:
		return "⏸"
	}
}
