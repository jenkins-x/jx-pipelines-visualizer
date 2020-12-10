package handlers

import (
	"net/http"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	jxclientv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/jenkins-x/jx-pipeline/pkg/tektonlog"
	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"

	"context"

	"k8s.io/apimachinery/pkg/api/errors"
)

type PipelineRunHandler struct {
	TektonClient tknclient.Interface
	PAInterface  jxclientv1.PipelineActivityInterface
	Store        *visualizer.Store
	Render       *render.Render
	Logger       *logrus.Logger
}

func (h *PipelineRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineRunName := vars["pipelineRun"]
	ns := vars["namespace"]
	if ns == "" {
		ns = "jx"
	}
	h.Logger.Info("rendering PipelineRun", pipelineRunName)

	ctx := context.Background()
	pr, err := h.TektonClient.TektonV1beta1().PipelineRuns(ns).Get(ctx, pipelineRunName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			h.RenderNotFound(w)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	pa, err := tektonlog.GetPipelineActivityForPipelineRun(context.TODO(), h.PAInterface, pr)
	if err != nil {
		if errors.IsNotFound(err) {
			h.RenderNotFound(w)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if pa == nil {
		h.RenderNotFound(w)
		return
	}

	owner := pa.Spec.GitOwner
	repo := pa.Spec.GitRepository
	branch := pa.Spec.GitBranch
	build := pa.Spec.Build

	if errors.IsNotFound(err) {
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

func (h *PipelineRunHandler) RenderNotFound(w http.ResponseWriter) {
	err := h.Render.HTML(w, http.StatusOK, "archived_logs", map[string]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
