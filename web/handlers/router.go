package handlers

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"text/template"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"
	"github.com/jenkins-x/jx-pipelines-visualizer/internal/version"
	"github.com/jenkins-x/jx-pipelines-visualizer/web/handlers/functions"

	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/mux"
	jxclient "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned/typed/jenkins.io/v1"
	"github.com/sirupsen/logrus"
	sse "github.com/subchord/go-sse"
	tknclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/unrolled/render"
	"github.com/urfave/negroni/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Router struct {
	Store                        *visualizer.Store
	KConfig                      *rest.Config
	PAInterface                  jenkinsv1.PipelineActivityInterface
	Namespace                    string
	ArchivedLogsURLTemplate      string
	ArchivedPipelinesURLTemplate string
	Logger                       *logrus.Logger
	render                       *render.Render
}

func (r Router) Handler() (http.Handler, error) {
	r.render = render.New(render.Options{
		Directory:     "web/templates",
		Layout:        "layout",
		IsDevelopment: version.Version == "dev",
		Funcs: []htmltemplate.FuncMap{
			sprig.HtmlFuncMap(),
			htmltemplate.FuncMap{
				"pipelinePullRequestURL": functions.PipelinePullRequestURL,
				"repositoryURL":          functions.RepositoryURL,
				"branchURL":              functions.BranchURL,
				"authorURL":              functions.AuthorURL,
				"vdate":                  functions.VDate,
				"sortPipelineCounts":     functions.SortPipelineCounts,
				"isAvailable":            functions.IsAvailable,
				"appVersion":             functions.AppVersion,
			},
		},
	})

	router := mux.NewRouter()
	router.StrictSlash(true)

	kClient, err := kubernetes.NewForConfig(r.KConfig)
	if err != nil {
		return nil, err
	}
	jxClient, err := jxclient.NewForConfig(r.KConfig)
	if err != nil {
		return nil, err
	}
	tknClient, err := tknclient.NewForConfig(r.KConfig)
	if err != nil {
		return nil, err
	}

	var archivedLogsURLTemplate *template.Template
	if len(r.ArchivedLogsURLTemplate) > 0 {
		archivedLogsURLTemplate, err = template.New("archivedLogsURL").Funcs(sprig.TxtFuncMap()).Parse(r.ArchivedLogsURLTemplate)
		if err != nil {
			return nil, err
		}
	}

	var archivedPipelinesURLTemplate *template.Template
	if len(r.ArchivedPipelinesURLTemplate) > 0 {
		archivedPipelinesURLTemplate, err = template.New("archivedPipelinesURL").Funcs(sprig.TxtFuncMap()).Parse(r.ArchivedPipelinesURLTemplate)
		if err != nil {
			return nil, err
		}
	}

	router.Handle("/", &HomeHandler{
		Store:  r.Store,
		Render: r.render,
		Logger: r.Logger,
	})

	router.Handle("/healthz", healthzHandler())

	router.Handle("/{owner}", &OwnerHandler{
		Store:  r.Store,
		Render: r.render,
		Logger: r.Logger,
	})

	router.Handle("/{owner}/{repo}", &RepositoryHandler{
		Store:  r.Store,
		Render: r.render,
		Logger: r.Logger,
	})

	router.Handle("/{owner}/{repo}/{branch}", &BranchHandler{
		Store:  r.Store,
		Render: r.render,
		Logger: r.Logger,
	})

	router.Handle("/{owner}/{repo}/{branch}/{build:[0-9]+}", &PipelineHandler{
		PAInterface:                r.PAInterface,
		StoredPipelinesURLTemplate: archivedPipelinesURLTemplate,
		BuildLogsURLTemplate:       archivedLogsURLTemplate,
		Render:                     r.render,
		Logger:                     r.Logger,
	})

	router.Handle("/{owner}/{repo}/{branch}/{build:[0-9]+}/logs", &LogsHandler{
		PAInterface:          r.PAInterface,
		BuildLogsURLTemplate: archivedLogsURLTemplate,
		Logger:               r.Logger,
	})

	router.Handle("/{owner}/{repo}/{branch}/{build:[0-9]+}/logs/live", &LiveLogsHandler{
		KubeClient:   kClient,
		JXClient:     jxClient,
		TektonClient: tknClient,
		Namespace:    r.Namespace,
		Broker:       sse.NewBroker(nil),
		Logger:       r.Logger,
	})

	router.Handle("/namespaces/{namespace}/pipelineruns/{pipelineRun}", &PipelineRunHandler{
		TektonClient: tknClient,
		PAInterface:  r.PAInterface,
		Store:        r.Store,
		Render:       r.render,
		Logger:       r.Logger,
	})

	router.Handle("/teams/{team}/projects/{owner}/{repo}/{branch}/{build:[0-9]+}", jxuiCompatibilityHandler(r.Namespace))

	handler := negroni.New(
		negroni.NewRecovery(),
		&negroni.Static{
			Dir:       http.Dir("web/static"),
			Prefix:    "/static",
			IndexFile: "index.html",
		},
		negroni.Wrap(router),
	)

	return handler, nil
}

func jxuiCompatibilityHandler(namespace string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		team := vars["team"]
		owner := vars["owner"]
		repo := vars["repo"]
		branch := vars["branch"]
		build := vars["build"]

		if team != namespace {
			http.NotFound(w, r)
			return
		}

		redirectURL := fmt.Sprintf("/%s/%s/%s/%s", owner, repo, branch, build)
		http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
	})
}

func healthzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
