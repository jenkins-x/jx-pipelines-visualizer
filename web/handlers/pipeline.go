package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	jenkinsv1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	jxclientv1 "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineHandler struct {
	PAInterface jxclientv1.PipelineActivityInterface
	Render      *render.Render
	Logger      *logrus.Logger
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

	pa, err := h.PAInterface.Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
