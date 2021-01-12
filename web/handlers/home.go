package handlers

import (
	"net/http"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type HomeHandler struct {
	Store  *visualizer.Store
	Render *render.Render
	Logger *logrus.Logger
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		query     = r.URL.Query().Get("q")
		pipelines *visualizer.Pipelines
		err       error
	)
	if query != "" {
		pipelines, err = h.Store.Query(visualizer.Query{
			Query: query,
		})
	} else {
		pipelines, err = h.Store.All()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Render.HTML(w, http.StatusOK, "home", struct {
		Pipelines *visualizer.Pipelines
		Query     string
	}{
		pipelines,
		query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
