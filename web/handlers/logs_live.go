package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jenkins-x-plugins/jx-pipeline/pkg/tektonlog"
	jxclient "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	sse "github.com/subchord/go-sse"
	tknv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	pipelineruns, labelSelector, err := h.getPipelineRuns(ctx, owner, repo, branch, build)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pipelineruns) == 0 {
		http.Error(w, fmt.Sprintf("no PipelineRun found using labelSelector %s", labelSelector), http.StatusTooEarly)
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
	for logLine := range logger.GetRunningBuildLogs(ctx, pa, pipelineruns, name) {
		h.send(r.Context(), clientConnection, "log", logLine.Line)
	}

	if err := logger.Err(); err == nil && len(pipelineruns) == 1 && pipelineruns[0].Labels["jenkins.io/pipelineType"] == "meta" {
		// if we started with only the meta-pipeline, let's now retry with the "real" build pipeline
		pipelineruns, _, _ = h.getPipelineRuns(ctx, owner, repo, branch, build, "jenkins.io/pipelineType=build")
		if len(pipelineruns) > 0 {
			for logLine := range logger.GetRunningBuildLogs(ctx, pa, pipelineruns, name) {
				h.send(r.Context(), clientConnection, "log", logLine.Line)
			}
		}
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

func (h *LiveLogsHandler) getPipelineRuns(ctx context.Context, owner, repo, branch, build string, extraSelectors ...string) ([]*tknv1beta1.PipelineRun, string, error) {
	var extraLabelSet labels.Set
	for _, extraSelector := range extraSelectors {
		labelSet, err := labels.ConvertSelectorToLabelsMap(extraSelector)
		if err != nil {
			return nil, "", err
		}
		extraLabelSet = labels.Merge(extraLabelSet, labelSet)
	}

	labelSet := labels.Set(map[string]string{
		"lighthouse.jenkins-x.io/refs.org":  owner,
		"lighthouse.jenkins-x.io/refs.repo": repo,
		"lighthouse.jenkins-x.io/branch":    branch,
		"lighthouse.jenkins-x.io/buildNum":  build,
	})
	labelSelector := labels.FormatLabels(labels.Merge(extraLabelSet, labelSet))
	prList, err := h.TektonClient.TektonV1beta1().PipelineRuns(h.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, labelSelector, err
	}

	if len(prList.Items) == 0 {
		// let's also try with the "old" labels used in jx v2
		labelSet := labels.Set(map[string]string{
			"owner":      owner,
			"repository": repo,
			"branch":     branch,
			"build":      build,
		})
		labelSelector := labels.FormatLabels(labels.Merge(extraLabelSet, labelSet))
		prList, err = h.TektonClient.TektonV1beta1().PipelineRuns(h.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, labelSelector, err
		}
	}

	prs := make([]*tknv1beta1.PipelineRun, 0, len(prList.Items))
	for i := range prList.Items {
		prs = append(prs, &prList.Items[i])
	}

	return prs, labelSelector, nil
}
