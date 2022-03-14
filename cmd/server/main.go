package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	v1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"net/http"
	"time"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	"github.com/jenkins-x/jx-pipelines-visualizer/internal/kube"
	"github.com/jenkins-x/jx-pipelines-visualizer/internal/version"
	"github.com/jenkins-x/jx-pipelines-visualizer/web/handlers"

	jxclientset "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
)

var (
	options struct {
		namespace                       string
		defaultJXNamespace              string
		resyncInterval                  time.Duration
		archivedLogsURLTemplate         string
		archivedPipelinesURLTemplate    string
		archivedPipelineRunsURLTemplate string
		pipelineTraceURLTemplate        string
		kubeConfigPath                  string
		listenAddr                      string
		logLevel                        string
		printVersion                    bool
	}
)

func init() {
	flag.StringVar(&options.namespace, "namespace", "", "Name of the jx namespace")
	flag.StringVar(&options.defaultJXNamespace, "default-jx-namespace", "jx", "Default Jenkins X installation namespace")
	flag.DurationVar(&options.resyncInterval, "resync-interval", 1*time.Minute, "Resync interval between full re-list operations")
	flag.StringVar(&options.archivedLogsURLTemplate, "archived-logs-url-template", "", "Go template string used to build the archived logs URL")
	flag.StringVar(&options.archivedPipelinesURLTemplate, "archived-pipelines-url-template", "", "Go template string used to build the archived pipelines URL")
	flag.StringVar(&options.archivedPipelineRunsURLTemplate, "archived-pipelineruns-url-template", "", "Go template string used to build the archived pipelineruns URL")
	flag.StringVar(&options.pipelineTraceURLTemplate, "pipeline-trace-url-template", "", "Go template string used to build the pipeline trace URL")
	flag.StringVar(&options.logLevel, "log-level", "INFO", "Log level - one of: trace, debug, info, warn(ing), error, fatal or panic")
	flag.StringVar(&options.kubeConfigPath, "kubeconfig", kube.DefaultKubeConfigPath(), "Kubernetes Config Path. Default: KUBECONFIG env var value")
	flag.StringVar(&options.listenAddr, "listen-addr", ":8080", "Address on which the server will listen for incoming connections")
	flag.BoolVar(&options.printVersion, "version", false, "Print the version")
}

func main() {
	flag.Parse()

	if options.printVersion {
		fmt.Printf("Version %s - Revision %s - Date %s", version.Version, version.Revision, version.Date)
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(options.logLevel)
	if err != nil {
		logger.WithField("logLevel", options.logLevel).Error("Failed to set log level")
	} else {
		logger.SetLevel(logLevel)
	}
	logger.WithField("logLevel", logLevel).Info("Starting")

	kClient, err := kube.NewClient(options.kubeConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to create a Kubernetes client")
	}
	jxClient, err := jxclientset.NewForConfig(kClient.Config)
	if err != nil {
		logger.WithError(err).Fatal("failed to create a Jenkins X client")
	}

	store, err := visualizer.NewStore()
	if err != nil {
		logger.WithError(err).Fatal("failed to create a new store")
	}

	runningPipelines := new(visualizer.RunningPipelines)

	logger.WithField("namespace", options.namespace).WithField("resyncInterval", options.resyncInterval).Info("Starting Informer")
	(&visualizer.Informer{
		JXClient:         jxClient,
		Namespace:        options.namespace,
		ResyncInterval:   options.resyncInterval,
		Store:            store,
		RunningPipelines: runningPipelines,
		Logger:           logger,
	}).Start(ctx)

	handler, err := handlers.Router{
		Store:            store,
		RunningPipelines: runningPipelines,
		KConfig:          kClient.Config,
		PAInterfaceFactory: func(namespace string) v1.PipelineActivityInterface {
			return jxClient.JenkinsV1().PipelineActivities(namespace)
		},
		Namespace:                       options.namespace,
		DefaultJXNamespace:              options.defaultJXNamespace,
		ArchivedLogsURLTemplate:         options.archivedLogsURLTemplate,
		ArchivedPipelinesURLTemplate:    options.archivedPipelinesURLTemplate,
		ArchivedPipelineRunsURLTemplate: options.archivedPipelineRunsURLTemplate,
		PipelineTraceURLTemplate:        options.pipelineTraceURLTemplate,
		Logger:                          logger,
	}.Handler()
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize the HTTP handler")
	}
	http.Handle("/", handler)

	logger.WithField("listenAddr", options.listenAddr).Info("Starting HTTP Server")
	err = http.ListenAndServe(options.listenAddr, nil)
	if !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatal("failed to start HTTP server")
	}
}
