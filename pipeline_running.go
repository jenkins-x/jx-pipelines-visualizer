package visualizer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	"github.com/sirupsen/logrus"
)

type RunningPipeline struct {
	Pipeline
	Stage          string
	StageStartTime time.Time
	Step           string
	StepStartTime  time.Time
}

func (running RunningPipeline) String() string {
	return fmt.Sprintf("%s/%s/%s", running.Name, running.Stage, running.Step)
}

func (running RunningPipeline) JSON() string {
	data, err := json.Marshal(running)
	if err != nil {
		return ""
	}
	return string(data)
}

type Watcher struct {
	Name    string
	Added   chan RunningPipeline
	Deleted chan RunningPipeline
}

type RunningPipelines struct {
	Logger   *logrus.Logger
	running  sync.Map
	watchers sync.Map
}

func (pipelines *RunningPipelines) Add(pa *jenkinsv1.PipelineActivity) {
	if pa == nil {
		return
	}

	runnings := pipelines.getForActivity(pa)
	if len(runnings) == 0 {
		if pa.Spec.Status.IsTerminated() {
			return
		} else {
			runnings = RunningPipelinesFromPipelineActivity(pa)
			for _, running := range runnings {
				pipelines.running.Store(running.String(), running)
				pipelines.onRunningPipelineAdded(running)
			}
			return
		}
	}

	if pa.Spec.Status.IsTerminated() {
		for _, running := range runnings {
			pipelines.running.Delete(running.String())
			pipelines.onRunningPipelineDeleted(running)
		}
		return
	}

	// delete runnings which are finished
	for _, running := range runnings {
		for _, stage := range pa.Spec.Steps {
			if stage.Stage != nil && stage.Stage.Name == running.Stage {
				for _, step := range stage.Stage.Steps {
					if step.Name == running.Step {
						if step.Status.IsTerminated() {
							pipelines.running.Delete(running.String())
							pipelines.onRunningPipelineDeleted(running)
						}
					}
				}
			}
		}
	}

	currentlyRunnings := RunningPipelinesFromPipelineActivity(pa)
	for _, currentlyRunning := range currentlyRunnings {
		var alreadyRunning bool
		for _, running := range runnings {
			if running.String() == currentlyRunning.String() {
				alreadyRunning = true
				break
			}
		}
		if !alreadyRunning {
			pipelines.running.Store(currentlyRunning.String(), currentlyRunning)
			pipelines.onRunningPipelineAdded(currentlyRunning)
		}
	}
}

func (pipelines *RunningPipelines) Get() []RunningPipeline {
	var runnings []RunningPipeline
	pipelines.running.Range(func(key, value interface{}) bool {
		if running, ok := value.(RunningPipeline); ok {
			runnings = append(runnings, running)
		}
		return true
	})
	return runnings
}

func (pipelines *RunningPipelines) getForActivity(pa *jenkinsv1.PipelineActivity) []RunningPipeline {
	var runnings []RunningPipeline
	pipelines.running.Range(func(key, value interface{}) bool {
		if running, ok := value.(RunningPipeline); ok {
			if running.Name == pa.Name {
				runnings = append(runnings, running)
			}
		}
		return true
	})
	return runnings
}

func (pipelines *RunningPipelines) onRunningPipelineAdded(running RunningPipeline) {
	pipelines.watchers.Range(func(key, value interface{}) bool {
		if watcher, ok := value.(Watcher); ok {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						pipelines.Logger.WithField("msg", r).Error("Panic when writing to channel")
					}
				}()
				watcher.Added <- running
			}()
		}
		return true
	})
}

func (pipelines *RunningPipelines) onRunningPipelineDeleted(running RunningPipeline) {
	pipelines.watchers.Range(func(key, value interface{}) bool {
		if watcher, ok := value.(Watcher); ok {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						pipelines.Logger.WithField("msg", r).Error("Panic when writing to channel")
					}
				}()
				watcher.Deleted <- running
			}()
		}
		return true
	})
}

func (pipelines *RunningPipelines) Register(watcher Watcher) {
	pipelines.watchers.Store(watcher.Name, watcher)
}

func (pipelines *RunningPipelines) UnRegister(watcher Watcher) {
	pipelines.watchers.Delete(watcher.Name)
}

func RunningPipelinesFromPipelineActivity(pa *jenkinsv1.PipelineActivity) []RunningPipeline {
	var runnings []RunningPipeline
	for _, stage := range pa.Spec.Steps {
		if stage.Stage != nil {
			for _, step := range stage.Stage.Steps {
				if step.Status == jenkinsv1.ActivityStatusTypeRunning {
					running := RunningPipeline{
						Pipeline:       PipelineFromPipelineActivity(pa),
						Stage:          stage.Stage.Name,
						StageStartTime: stage.Stage.StartedTimestamp.Time,
						Step:           step.Name,
						StepStartTime:  step.StartedTimestamp.Time,
					}
					runnings = append(runnings, running)
				}
			}
		}
	}
	return runnings
}
