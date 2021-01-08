package visualizer

import (
	"strings"
	"time"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

type Pipeline struct {
	Name            string
	Provider        string
	Owner           string
	Repository      string
	Branch          string
	Build           string
	Context         string
	Author          string
	AuthorAvatarURL string
	Commit          string
	Status          string
	Start           time.Time
	End             time.Time
	Duration        time.Duration
}

func (p Pipeline) PullRequestNumber() string {
	if strings.HasPrefix(p.Branch, "PR-") {
		return strings.TrimPrefix(p.Branch, "PR-")
	}
	return ""
}

func PipelineFromPipelineActivity(pa *jenkinsv1.PipelineActivity) Pipeline {
	p := Pipeline{
		Name:            pa.Name,
		Provider:        pa.Labels["provider"],
		Owner:           pa.Spec.GitOwner,
		Repository:      pa.Spec.GitRepository,
		Branch:          pa.Spec.GitBranch,
		Build:           pa.Spec.Build,
		Context:         pa.Spec.Context,
		Author:          pa.Spec.Author,
		AuthorAvatarURL: pa.Spec.AuthorAvatarURL,
		Commit:          pa.Spec.LastCommitSHA,
		Status:          string(pa.Spec.Status),
	}
	if pa.Spec.StartedTimestamp != nil {
		p.Start = pa.Spec.StartedTimestamp.Time
	} else {
		p.Start = pa.CreationTimestamp.Time
	}
	if pa.Spec.CompletedTimestamp != nil {
		p.End = pa.Spec.CompletedTimestamp.Time
	}
	if !p.Start.IsZero() && !p.End.IsZero() {
		p.Duration = p.End.Sub(p.Start)
	}
	return p
}
