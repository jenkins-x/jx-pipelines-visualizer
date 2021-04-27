package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	"github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

// ShieldsIOBadge is documented at https://shields.io/endpoint
type ShieldsIOBadge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

type ShieldsIOHandler struct {
	Store  *visualizer.Store
	Render *render.Render
	Logger *logrus.Logger
}

func (h *ShieldsIOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repository := vars["repo"]
	branch := vars["branch"]

	pipelines, err := h.Store.Query(visualizer.Query{
		Owner:      owner,
		Repository: repository,
		Branch:     branch,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pipelines.Pipelines) == 0 {
		http.NotFound(w, r)
		return
	}
	lastPipeline := pipelines.Pipelines[0]

	shieldsIOBadge := ShieldsIOBadge{
		SchemaVersion: 1,
		Label:         "Jenkins X",
		Message:       fmt.Sprintf("%s #%v %s", lastPipeline.Branch, lastPipeline.Build, lastPipeline.Status),
		Color:         h.pipelineStatusToColor(lastPipeline.Status),
	}

	err = h.Render.JSON(w, http.StatusOK, shieldsIOBadge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ShieldsIOHandler) pipelineStatusToColor(status string) string {
	switch status {
	case "Succeeded":
		return "green"
	case "Failed":
		return "red"
	case "Running":
		return "blue"
	default:
		return "grey"
	}
}
