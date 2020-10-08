package visualizer

import (
	"time"

	jenkinsv1 "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io/v1"
)

type Pipeline struct {
	Name            string
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

func PipelineFromPipelineActivity(pa *jenkinsv1.PipelineActivity) Pipeline {
	p := Pipeline{
		Name:            pa.Name,
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
	}
	if pa.Spec.CompletedTimestamp != nil {
		p.End = pa.Spec.CompletedTimestamp.Time
	}
	if !p.Start.IsZero() && !p.End.IsZero() {
		p.Duration = p.End.Sub(p.Start)
	}
	return p
}
