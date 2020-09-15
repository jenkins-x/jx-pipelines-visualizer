package handlers

import (
	"net/http"

	visualizer "github.com/dailymotion/jx-pipelines-visualizer"

	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type HomeHandler struct {
	Store  *visualizer.Store
	Render *render.Render
	Logger *logrus.Logger
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pipelines, err := h.Store.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Render.HTML(w, http.StatusOK, "home", struct {
		Pipelines *visualizer.Pipelines
	}{
		pipelines,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
