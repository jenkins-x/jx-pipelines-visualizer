package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	jxclientv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx-pipeline/pkg/cloud/buckets"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineHandler struct {
	PAInterface                jxclientv1.PipelineActivityInterface
	BuildLogsURLTemplate       *template.Template
	StoredPipelinesURLTemplate *template.Template
	Render                     *render.Render
	Logger                     *logrus.Logger
}

func (h *PipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	branch := vars["branch"]
	if strings.HasPrefix(branch, "pr-") {
		branch = strings.ToUpper(branch)
	}
	build := vars["build"]

	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, build))

	ctx := context.Background()
	pa, err := h.PAInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if errors.IsNotFound(err) {
		pa = nil
	}

	if pa == nil {
		pa, err = h.loadPipelineFromStorage(ctx, owner, repo, branch, build)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if pa == nil {
		err := h.Render.HTML(w, http.StatusOK, "archived_logs", map[string]string{
			"Owner":      owner,
			"Repository": repo,
			"Branch":     branch,
			"Build":      build,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.Render.HTML(w, http.StatusOK, "pipeline", struct {
		Pipeline *jenkinsv1.PipelineActivity
	}{
		pa,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PipelineHandler) loadPipelineFromStorage(ctx context.Context, owner, repo, branch, build string) (*jenkinsv1.PipelineActivity, error) {
	storedPipelineURL, err := h.storedPipelineURL(owner, repo, branch, build)
	if err != nil {
		return nil, err
	}

	if storedPipelineURL == "" {
		return nil, nil
	}

	httpFn := func(urlString string) (string, func(*http.Request), error) {
		return urlString, func(*http.Request) {}, nil
	}
	reader, err := buckets.ReadURL(ctx, storedPipelineURL, 30*time.Second, httpFn)
	if err != nil {
		if strings.Contains(err.Error(), "object doesn't exist") {
			return nil, nil
		}
		return nil, err
	}
	defer reader.Close()

	var pa jenkinsv1.PipelineActivity
	err = yaml.NewDecoder(reader).Decode(&pa)
	if err != nil {
		return nil, err
	}

	if pa.Spec.BuildLogsURL == "" {
		pa.Spec.BuildLogsURL, err = h.buildLogsURL(owner, repo, branch, build)
		if err != nil {
			return nil, err
		}
	}

	return &pa, nil
}

func (h *PipelineHandler) storedPipelineURL(owner, repo, branch, build string) (string, error) {
	if h.StoredPipelinesURLTemplate == nil {
		return "", nil
	}

	sb := new(strings.Builder)
	err := h.StoredPipelinesURLTemplate.Execute(sb, map[string]string{
		"Owner":      owner,
		"Repository": repo,
		"Branch":     branch,
		"Build":      build,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate stored pipeline URL: %w", err)
	}

	return sb.String(), nil
}

func (h *PipelineHandler) buildLogsURL(owner, repo, branch, build string) (string, error) {
	if h.BuildLogsURLTemplate == nil {
		return "", nil
	}

	sb := new(strings.Builder)
	err := h.BuildLogsURLTemplate.Execute(sb, map[string]string{
		"Owner":      owner,
		"Repository": repo,
		"Branch":     branch,
		"Build":      build,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate build logs URL: %w", err)
	}

	return sb.String(), nil
}
