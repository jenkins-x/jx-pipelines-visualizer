package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	jxclientv1 "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx/v2/pkg/cloud/buckets"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogsHandler struct {
	PAInterface          jxclientv1.PipelineActivityInterface
	BuildLogsURLTemplate *template.Template
	Logger               *logrus.Logger
}

func (h *LogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	branch := vars["branch"]
	if strings.HasPrefix(branch, "pr-") {
		branch = strings.ToUpper(branch)
	}
	build := vars["build"]

	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, build))

	var buildLogsURL string
	pa, err := h.PAInterface.Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pa != nil {
		buildLogsURL = pa.Spec.BuildLogsURL
	}
	if len(buildLogsURL) == 0 {
		buildLogsURL, err = h.buildLogsURL(owner, repo, branch, build)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if len(buildLogsURL) == 0 {
		http.NotFound(w, r)
		return
	}

	httpFn := func(urlString string) (string, func(*http.Request), error) {
		return urlString, func(*http.Request) {}, nil
	}
	reader, err := buckets.ReadURL(buildLogsURL, 30*time.Second, httpFn)
	if err != nil {
		if strings.Contains(err.Error(), "object doesn't exist") {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *LogsHandler) buildLogsURL(owner, repo, branch, build string) (string, error) {
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
