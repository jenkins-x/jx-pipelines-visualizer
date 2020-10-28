package handlers

import (
	"net/http"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type RepositoryHandler struct {
	Store  *visualizer.Store
	Render *render.Render
	Logger *logrus.Logger
}

func (h *RepositoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repository := vars["repo"]

	pipelines, err := h.Store.Query(visualizer.Query{
		Owner:      owner,
		Repository: repository,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Render.HTML(w, http.StatusOK, "home", struct {
		Owner      string
		Repository string
		Pipelines  *visualizer.Pipelines
	}{
		owner,
		repository,
		pipelines,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
