package handlers

import (
	"net/http"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type RunningHandler struct {
	RunningPipelines *visualizer.RunningPipelines
	Render           *render.Render
	Logger           *logrus.Logger
}

func (h *RunningHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Render.HTML(w, http.StatusOK, "running", struct {
		Pipelines []visualizer.RunningPipeline
	}{
		h.RunningPipelines.Get(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
