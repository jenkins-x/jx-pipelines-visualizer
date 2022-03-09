package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/jenkins-x-plugins/jx-pipeline/pkg/cloud/buckets"
	jxclientv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogsHandler struct {
	PAInterfaceFactory   func(namespace string) jxclientv1.PipelineActivityInterface
	DefaultJXNamespace   string
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
	namespace := vars["namespace"]
	if namespace == "" {
		namespace = h.DefaultJXNamespace
	}

	name := naming.ToValidName(fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, build))

	ctx := context.Background()
	var buildLogsURL string
	pa, err := h.PAInterfaceFactory(namespace).Get(ctx, name, metav1.GetOptions{})
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
	reader, err := buckets.ReadURL(ctx, buildLogsURL, 30*time.Second, httpFn)
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
