package visualizer

import (
	"context"
	"time"

	jenkinsv1 "github.com/jenkins-x/jx-api/pkg/apis/jenkins.io/v1"
	jxclientset "github.com/jenkins-x/jx-api/pkg/client/clientset/versioned"
	informers "github.com/jenkins-x/jx-api/pkg/client/informers/externalversions"
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
			pa := obj.(*jenkinsv1.PipelineActivity)

			// don't index (static) Jenkins PipelineActivity
			if _, found := GetContext(pa); !found {
				if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
					i.Logger.WithField("PipelineActivity", pa.Name).Debug("Ignoring PipelineActivity without context")
				}
				return
			}

			if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
				i.Logger.WithField("PipelineActivity", pa.Name).Debug("Indexing new PipelineActivity")
			}
			p := PipelineFromPipelineActivity(pa)
			err := i.Store.Add(p)
			if err != nil && i.Logger != nil {
				i.Logger.WithError(err).WithField("PipelineActivity", pa.Name).Error("failed to index new PipelineActivity")
			}
		},
		UpdateFunc: func(old, new interface{}) {
			pa := new.(*jenkinsv1.PipelineActivity)

			// don't index (static) Jenkins PipelineActivity
			if _, found := GetContext(pa); !found {
				if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
					i.Logger.WithField("PipelineActivity", pa.Name).Debug("Ignoring PipelineActivity without context")
				}
				return
			}

			if i.Logger != nil && i.Logger.IsLevelEnabled(logrus.DebugLevel) {
				i.Logger.WithField("PipelineActivity", pa.Name).Debug("Re-indexing PipelineActivity")
			}
			p := PipelineFromPipelineActivity(pa)
			err := i.Store.Add(p)
			if err != nil && i.Logger != nil {
				i.Logger.WithError(err).WithField("PipelineActivity", pa.Name).Error("failed to re-index PipelineActivity")
			}
		},
		DeleteFunc: func(obj interface{}) {
			pa := obj.(*jenkinsv1.PipelineActivity)
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

// GetContext returns the pipeline context using the v2 or v3 labels
func GetContext(pa *jenkinsv1.PipelineActivity) (string, bool) {
	if pa.Spec.Context != "" {
		return pa.Spec.Context, true
	}
	for _, label := range []string{"context", "lighthouse.jenkins-x.io/context"} {
		answer := pa.Labels[label]
		if answer != "" {
			return answer, true
		}
	}
	return "", false
}
