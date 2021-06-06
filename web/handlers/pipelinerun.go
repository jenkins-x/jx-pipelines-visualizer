package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/jenkins-x-plugins/jx-pipeline/pkg/cloud/buckets"
	"github.com/jenkins-x-plugins/jx-pipeline/pkg/tektonlog"
	jxclientv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/activities"
	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	"github.com/sirupsen/logrus"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/unrolled/render"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineRunHandler struct {
	TektonClient                  tknclient.Interface
	PAInterface                   jxclientv1.PipelineActivityInterface
	StoredPipelineRunsURLTemplate *template.Template
	Namespace                     string
	Store                         *visualizer.Store
	Render                        *render.Render
	Logger                        *logrus.Logger
}

func (h *PipelineRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineRunName := vars["pipelineRun"]
	ns := vars["namespace"]
	if ns == "" {
		ns = h.Namespace
	}

	ctx := context.Background()
	pr, err := h.TektonClient.TektonV1beta1().PipelineRuns(ns).Get(ctx, pipelineRunName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if errors.IsNotFound(err) {
		pr = nil
	}

	if pr == nil {
		pr, err = h.loadPipelineRunFromStorage(ctx, ns, pipelineRunName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if pr == nil {
		http.NotFound(w, r)
		return
	}

	var (
		owner  = activities.GetLabel(pr.Labels, activities.OwnerLabels)
		repo   = activities.GetLabel(pr.Labels, activities.RepoLabels)
		branch = activities.GetLabel(pr.Labels, activities.BranchLabels)
		build  = pr.Labels["build"]
	)
	if owner == "" || repo == "" || branch == "" || build == "" {
		pa, err := tektonlog.GetPipelineActivityForPipelineRun(context.TODO(), h.PAInterface, pr)
		if err != nil && !errors.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if pa == nil {
			http.NotFound(w, r)
			return
		}
		owner = pa.Spec.GitOwner
		repo = pa.Spec.GitRepository
		branch = pa.Spec.GitBranch
		build = pa.Spec.Build
	}

	redirectURL := fmt.Sprintf("/%s/%s/%s/%s", owner, repo, branch, build)
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

func (h *PipelineRunHandler) loadPipelineRunFromStorage(ctx context.Context, namespace, name string) (*v1beta1.PipelineRun, error) {
	storedPipelineRunURL, err := h.storedPipelineRunURL(namespace, name)
	if err != nil {
		return nil, err
	}

	if storedPipelineRunURL == "" {
		return nil, nil
	}

	httpFn := func(urlString string) (string, func(*http.Request), error) {
		return urlString, func(*http.Request) {}, nil
	}
	reader, err := buckets.ReadURL(ctx, storedPipelineRunURL, 30*time.Second, httpFn)
	if err != nil {
		if strings.Contains(err.Error(), "object doesn't exist") {
			return nil, nil
		}
		return nil, err
	}
	defer reader.Close()

	var pr v1beta1.PipelineRun
	err = yaml.NewDecoder(reader).Decode(&pr)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (h *PipelineRunHandler) storedPipelineRunURL(namespace, name string) (string, error) {
	if h.StoredPipelineRunsURLTemplate == nil {
		return "", nil
	}

	sb := new(strings.Builder)
	err := h.StoredPipelineRunsURLTemplate.Execute(sb, map[string]string{
		"Namespace": namespace,
		"Name":      name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate stored pipeline run URL: %w", err)
	}

	return sb.String(), nil
}
