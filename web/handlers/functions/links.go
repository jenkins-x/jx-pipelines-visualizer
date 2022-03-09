package functions

import (
	"fmt"
	"strings"
	"text/template"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

func TraceURLFunc(pipelineTraceURLTemplate *template.Template) func(string) string {
	return func(traceID string) string {
		return traceIDToTraceURL(traceID, pipelineTraceURLTemplate)
	}
}

func RepositoryURL(pipeline interface{}) string {
	switch p := pipeline.(type) {
	case visualizer.Pipeline:
		return repositoryURLForPipeline(p)
	case visualizer.RunningPipeline:
		return repositoryURLForPipeline(p.Pipeline)
	case *jenkinsv1.PipelineActivity:
		return repositoryURLForPipelineActivity(p)
	default:
		return ""
	}
}

func PullRequestURL(pipeline interface{}) string {
	switch p := pipeline.(type) {
	case visualizer.Pipeline:
		return pullRequestURLForPipeline(p)
	case visualizer.RunningPipeline:
		return pullRequestURLForPipeline(p.Pipeline)
	case *jenkinsv1.PipelineActivity:
		return pullRequestURLForPipelineActivity(p)
	default:
		return ""
	}
}

func BranchURL(pipeline interface{}) string {
	switch p := pipeline.(type) {
	case visualizer.Pipeline:
		return branchURLForPipeline(p)
	case visualizer.RunningPipeline:
		return branchURLForPipeline(p.Pipeline)
	case *jenkinsv1.PipelineActivity:
		return branchURLForPipelineActivity(p)
	default:
		return ""
	}
}

func CommitURL(pipeline interface{}) string {
	switch p := pipeline.(type) {
	case *jenkinsv1.PipelineActivity:
		return commitURLForPipelineActivity(p)
	default:
		return ""
	}
}

func AuthorURL(pipeline interface{}) string {
	switch p := pipeline.(type) {
	case visualizer.Pipeline:
		return authorURLForPipeline(p)
	case visualizer.RunningPipeline:
		return authorURLForPipeline(p.Pipeline)
	case *jenkinsv1.PipelineActivity:
		return authorURLForPipelineActivity(p)
	default:
		return ""
	}
}

func authorURLForPipeline(pipeline visualizer.Pipeline) string {
	switch pipeline.Provider {
	case "github":
		return fmt.Sprintf("https://github.com/%s", pipeline.Author)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s", pipeline.Author)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s", pipeline.Author)
	default:
		return ""
	}
}

func authorURLForPipelineActivity(pa *jenkinsv1.PipelineActivity) string {
	switch pipelineActivityProvider(pa) {
	case "github":
		return fmt.Sprintf("https://github.com/%s", pa.Spec.Author)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s", pa.Spec.Author)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s", pa.Spec.Author)
	default:
		return ""
	}
}

func repositoryURLForPipeline(pipeline visualizer.Pipeline) string {
	if pipeline.GitUrl != "" {
		return pipeline.GitUrl
	}

	switch pipeline.Provider {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s", pipeline.Owner, pipeline.Repository)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s", pipeline.Owner, pipeline.Repository)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s", pipeline.Owner, pipeline.Repository)
	default:
		return ""
	}
}

func repositoryURLForPipelineActivity(pa *jenkinsv1.PipelineActivity) string {
	switch pipelineActivityProvider(pa) {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s", pa.Spec.GitOwner, pa.Spec.GitRepository)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s", pa.Spec.GitOwner, pa.Spec.GitRepository)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s", pa.Spec.GitOwner, pa.Spec.GitRepository)
	default:
		return ""
	}
}

func pullRequestURLForPipeline(pipeline visualizer.Pipeline) string {
	if pipeline.PullRequestNumber() == "" {
		return "" // not a PR
	}
	switch pipeline.Provider {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s/pull/%s", pipeline.Owner, pipeline.Repository, pipeline.PullRequestNumber())
	case "gitlab":
		if pipeline.GitUrl != "" {
			return fmt.Sprintf("%s/%s/%s/-/merge_requests/%s", pipeline.GitUrl, pipeline.Owner, pipeline.Repository, pipeline.PullRequestNumber())
		}

		return fmt.Sprintf("https://gitlab.com/%s/%s/-/merge_requests/%s", pipeline.Owner, pipeline.Repository, pipeline.PullRequestNumber())
	case "bitbucket":
		if pipeline.GitUrl != "" {
			return fmt.Sprintf("%s/%s/%s/pull-requests/%s", pipeline.GitUrl, pipeline.Owner, pipeline.Repository, pipeline.PullRequestNumber())
		}

		return fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%s", pipeline.Owner, pipeline.Repository, pipeline.PullRequestNumber())
	default:
		return ""
	}
}

func pullRequestURLForPipelineActivity(pa *jenkinsv1.PipelineActivity) string {
	if !strings.HasPrefix(pa.Spec.GitBranch, "PR-") {
		return "" // not a PR
	}
	prNumber := strings.TrimPrefix(pa.Spec.GitBranch, "PR-")
	switch pipelineActivityProvider(pa) {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s/pull/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, prNumber)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s/-/merge_requests/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, prNumber)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, prNumber)
	default:
		return ""
	}
}

func branchURLForPipeline(pipeline visualizer.Pipeline) string {
	if pipeline.PullRequestNumber() != "" {
		return pullRequestURLForPipeline(pipeline)
	}
	switch pipeline.Provider {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s/tree/%s", pipeline.Owner, pipeline.Repository, pipeline.Branch)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s/-/tree/%s", pipeline.Owner, pipeline.Repository, pipeline.Branch)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s/branch/%s", pipeline.Owner, pipeline.Repository, pipeline.Branch)
	default:
		return ""
	}
}

func branchURLForPipelineActivity(pa *jenkinsv1.PipelineActivity) string {
	if strings.HasPrefix(pa.Spec.GitBranch, "PR-") {
		return pullRequestURLForPipelineActivity(pa)
	}
	switch pipelineActivityProvider(pa) {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s/tree/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.GitBranch)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s/-/tree/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.GitBranch)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s/branch/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.GitBranch)
	default:
		return ""
	}
}

func commitURLForPipelineActivity(pa *jenkinsv1.PipelineActivity) string {
	if len(pa.Spec.LastCommitURL) > 0 {
		return pa.Spec.LastCommitURL
	}
	switch pipelineActivityProvider(pa) {
	case "github":
		return fmt.Sprintf("https://github.com/%s/%s/commit/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.LastCommitSHA)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s/-/commit/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.LastCommitSHA)
	case "bitbucket":
		return fmt.Sprintf("https://bitbucket.org/%s/%s/commits/%s", pa.Spec.GitOwner, pa.Spec.GitRepository, pa.Spec.LastCommitSHA)
	default:
		return ""
	}
}

func pipelineActivityProvider(pa *jenkinsv1.PipelineActivity) string {
	if provider := pa.Labels["provider"]; provider != "" {
		return provider
	}

	if strings.Contains(pa.Spec.GitURL, "github") {
		return "github"
	}
	if strings.Contains(pa.Spec.GitURL, "gitlab") {
		return "gitlab"
	}
	if strings.Contains(pa.Spec.GitURL, "bitbucket") {
		return "bitbucket"
	}

	return ""
}

func traceIDToTraceURL(traceID string, pipelineTraceURLTemplate *template.Template) string {
	if pipelineTraceURLTemplate == nil {
		return ""
	}
	if traceID == "" {
		return ""
	}

	sb := new(strings.Builder)
	err := pipelineTraceURLTemplate.Execute(sb, map[string]string{
		"TraceID": traceID,
	})
	if err != nil {
		return err.Error()
	}
	return sb.String()
}
