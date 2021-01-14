package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	jxclientv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx-pipeline/pkg/tektonlog"
	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	"github.com/sirupsen/logrus"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/unrolled/render"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineRunHandler struct {
	TektonClient tknclient.Interface
	PAInterface  jxclientv1.PipelineActivityInterface
	Namespace    string
	Store        *visualizer.Store
	Render       *render.Render
	Logger       *logrus.Logger
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
	if err != nil {
		if errors.IsNotFound(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	pa, err := tektonlog.GetPipelineActivityForPipelineRun(context.TODO(), h.PAInterface, pr)
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pa == nil {
		http.NotFound(w, r)
		return
	}

	owner := pa.Spec.GitOwner
	repo := pa.Spec.GitRepository
	branch := pa.Spec.GitBranch
	build := pa.Spec.Build
	redirectURL := fmt.Sprintf("/%s/%s/%s/%s", owner, repo, branch, build)
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}
