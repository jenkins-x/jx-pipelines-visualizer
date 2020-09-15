package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	jxclient "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/v2/pkg/logs"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	sse "github.com/subchord/go-sse"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LiveLogsHandler struct {
	JXClient     jxclient.Interface
	TektonClient tknclient.Interface
	KubeClient   kubernetes.Interface
	Namespace    string
	Broker       *sse.Broker
	Logger       *logrus.Logger
}

func (h *LiveLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]
	branch := vars["branch"]
	if strings.HasPrefix(branch, "pr-") {
		branch = strings.ToUpper(branch)
	}
	build := vars["build"]

	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, build))

	pa, err := h.JXClient.JenkinsV1().PipelineActivities(h.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clientConnection, err := h.Broker.Connect(xid.New().String(), w, r)
	if err != nil {
		// streaming unsupported. http.Error() already used in broker.Connect()
		return
	}

	logger := &logs.TektonLogger{
		KubeClient:   h.KubeClient,
		JXClient:     h.JXClient,
		TektonClient: h.TektonClient,
		Namespace:    h.Namespace,
	}
	for logLine := range logger.GetRunningBuildLogs(pa, "", false) {
		h.send(r.Context(), clientConnection, "log", logLine.Line)
	}

	if err := logger.Err(); err != nil {
		h.send(r.Context(), clientConnection, "error", err.Error())
	}

	h.send(r.Context(), clientConnection, "EOF", "End Of Feed")

	select {
	case <-clientConnection.Done():
	case <-r.Context().Done():
	}
}

func (h *LiveLogsHandler) send(ctx context.Context, clientConnection *sse.ClientConnection, eventType, eventData string) {
	select {
	case <-clientConnection.Done():
		return
	case <-ctx.Done():
		return
	default:
		clientConnection.Send(sse.StringEvent{
			Id:    xid.New().String(),
			Event: eventType,
			Data:  eventData,
		})
	}
}
