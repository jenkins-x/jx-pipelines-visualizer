package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	visualizer "github.com/dailymotion/jx-pipelines-visualizer"
	"github.com/dailymotion/jx-pipelines-visualizer/internal/kube"
	"github.com/dailymotion/jx-pipelines-visualizer/web/handlers"

	jxclientset "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
)

var (
	options struct {
		namespace               string
		resyncInterval          time.Duration
		archivedLogsURLTemplate string
		kubeConfigPath          string
		listenAddr              string
		logLevel                string
		printVersion            bool
	}

	// these are set at compile time by GoReleaser through LD Flags
	version = "dev"
	commit  = "unknown"
	date    = "now"
)

func init() {
	flag.StringVar(&options.namespace, "namespace", "jx", "Name of the jx namespace")
	flag.DurationVar(&options.resyncInterval, "resync-interval", 1*time.Minute, "Resync interval between full re-list operations")
	flag.StringVar(&options.archivedLogsURLTemplate, "archived-logs-url-template", "", "Go template string used to build the archived logs URL")
	flag.StringVar(&options.logLevel, "log-level", "INFO", "Log level - one of: trace, debug, info, warn(ing), error, fatal or panic")
	flag.StringVar(&options.kubeConfigPath, "kubeconfig", kube.DefaultKubeConfigPath(), "Kubernetes Config Path. Default: KUBECONFIG env var value")
	flag.StringVar(&options.listenAddr, "listen-addr", ":8080", "Address on which the server will listen for incoming connections")
	flag.BoolVar(&options.printVersion, "version", false, "Print the version")
}

func main() {
	flag.Parse()

	if options.printVersion {
		fmt.Printf("Version %s - Commit %s - Date %s", version, commit, date)
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

	logger.WithField("namespace", options.namespace).WithField("resyncInterval", options.resyncInterval).Info("Starting Informer")
	(&visualizer.Informer{
		JXClient:       jxClient,
		Namespace:      options.namespace,
		ResyncInterval: options.resyncInterval,
		Store:          store,
		Logger:         logger,
	}).Start(ctx)

	handler, err := handlers.Router{
		Store:                   store,
		KConfig:                 kClient.Config,
		PAInterface:             jxClient.JenkinsV1().PipelineActivities(options.namespace),
		Namespace:               options.namespace,
		ArchivedLogsURLTemplate: options.archivedLogsURLTemplate,
		Logger:                  logger,
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
