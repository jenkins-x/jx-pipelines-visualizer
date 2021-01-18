package visualizer

import (
	"context"
	"strings"
	"time"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	jxclientset "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	informers "github.com/jenkins-x/jx-api/v4/pkg/client/informers/externalversions"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
)

type Informer struct {
	JXClient       *jxclientset.Clientset
	Namespace      string
	ResyncInterval time.Duration
	Store          *Store
	Logger         *logrus.Logger
}

func (i *Informer) Start(ctx context.Context) {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		i.JXClient,
		i.ResyncInterval,
		informers.WithNamespace(i.Namespace),
	)

	informerFactory.Jenkins().V1().PipelineActivities().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pa, ok := obj.(*jenkinsv1.PipelineActivity)
			if !ok {
				return
			}

			i.indexPipelineActivity(pa, "index")
		},
		UpdateFunc: func(old, new interface{}) {
			pa, ok := new.(*jenkinsv1.PipelineActivity)
			if !ok {
				return
			}

			i.indexPipelineActivity(pa, "re-index")
		},
		DeleteFunc: func(obj interface{}) {
			pa, ok := obj.(*jenkinsv1.PipelineActivity)
			if !ok {
				return
			}

			if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
				i.Logger.WithField("PipelineActivity", pa.Name).Debug("Deleting PipelineActivity")
			}
			err := i.Store.Delete(pa.Name)
			if err != nil && i.Logger != nil {
				i.Logger.WithError(err).WithField("PipelineActivity", pa.Name).Error("failed to delete PipelineActivity")
			}
		},
	})

	informerFactory.Start(ctx.Done())
}

func (i *Informer) indexPipelineActivity(pa *jenkinsv1.PipelineActivity, operation string) {
	if isJenkinsPipelineActivity(pa) {
		if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
			i.Logger.WithField("PipelineActivity", pa.Name).Debug("Ignoring PipelineActivity created by Jenkins")
		}
		return
	}

	if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
		i.Logger.WithField("PipelineActivity", pa.Name).Debugf("%sing new PipelineActivity", strings.Title(operation))
	}
	p := PipelineFromPipelineActivity(pa)
	err := i.Store.Add(p)
	if err != nil && i.Logger != nil {
		i.Logger.WithError(err).WithField("PipelineActivity", pa.Name).Errorf("failed to %s new PipelineActivity", operation)
	}
}

// isJenkinsPipelineActivity returns true if the given PipelineActivity has been created by Jenkins
// see https://github.com/jenkinsci/jx-resources-plugin/blob/master/src/main/java/org/jenkinsci/plugins/jx/resources/BuildSyncRunListener.java#L106
func isJenkinsPipelineActivity(pa *jenkinsv1.PipelineActivity) bool {
	if strings.Contains(pa.Spec.BuildURL, "/blue/organizations/jenkins/") {
		return true
	}
	if strings.Contains(pa.Spec.BuildLogsURL, "/job/") && strings.HasSuffix(pa.Spec.BuildLogsURL, "/console") {
		return true
	}
	return false
}
