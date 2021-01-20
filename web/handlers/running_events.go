package handlers

import (
	"context"
	"net/http"

	visualizer "github.com/jenkins-x/jx-pipelines-visualizer"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	sse "github.com/subchord/go-sse"
)

type RunningEventsHandler struct {
	RunningPipelines *visualizer.RunningPipelines
	Broker           *sse.Broker
	Logger           *logrus.Logger
}

func (h *RunningEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := xid.New().String()
	clientConnection, err := h.Broker.Connect(clientID, w, r)
	if err != nil {
		// streaming unsupported. http.Error() already used in broker.Connect()
		return
	}

	watcher := visualizer.Watcher{
		Name:    clientID,
		Added:   make(chan visualizer.RunningPipeline),
		Deleted: make(chan visualizer.RunningPipeline),
	}
	h.RunningPipelines.Register(watcher)

	for {
		select {
		case running := <-watcher.Added:
			h.send(r.Context(), clientConnection, "added", running.JSON())
		case running := <-watcher.Deleted:
			h.send(r.Context(), clientConnection, "deleted", running.JSON())
		case <-clientConnection.Done():
			h.RunningPipelines.UnRegister(watcher)
			return
		case <-r.Context().Done():
			h.RunningPipelines.UnRegister(watcher)
			return
		}
	}
}

func (h *RunningEventsHandler) send(ctx context.Context, clientConnection *sse.ClientConnection, eventType, eventData string) {
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
