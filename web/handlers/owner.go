package handlers

import (
	"net/http"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type OwnerHandler struct {
	Store  *visualizer.Store
	Render *render.Render
	Logger *logrus.Logger
}

func (h *OwnerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]

	pipelines, err := h.Store.Query(visualizer.Query{
		Owner: owner,
		Query: r.URL.Query().Get("q"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Render.HTML(w, http.StatusOK, "home", struct {
		Owner     string
		Query     string
		Pipelines *visualizer.Pipelines
	}{
		owner,
		r.URL.Query().Get("q"),
		pipelines,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
