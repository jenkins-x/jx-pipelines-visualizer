package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	jxclient "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	"github.com/jenkins-x/jx-pipeline/pkg/tektonlog"
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

	name := naming.ToValidName(fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, build))

	ctx := context.Background()
	pa, err := h.JXClient.JenkinsV1().PipelineActivities(h.Namespace).Get(ctx, name, metav1.GetOptions{})
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

	logger := &tektonlog.TektonLogger{
		KubeClient:   h.KubeClient,
		JXClient:     h.JXClient,
		TektonClient: h.TektonClient,
		Namespace:    h.Namespace,
	}
	logContext := vars["logContext"]
	if logContext == "" {
		logContext = pa.Spec.Context
	}
	buildFilter := &tektonlog.BuildPodInfoFilter{
		Owner:      owner,
		Repository: repo,
		Branch:     branch,
		Build:      build,
		Context:    logContext,
	}
	_, _, prMap, err := logger.GetTektonPipelinesWithActivePipelineActivity(context.TODO(), buildFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := strings.ToLower(fmt.Sprintf("%s/%s/%s #%s %s", owner, repo, branch, build, logContext))
	prList := prMap[key]
	for logLine := range logger.GetRunningBuildLogs(context.TODO(), pa, prList, name) {
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
