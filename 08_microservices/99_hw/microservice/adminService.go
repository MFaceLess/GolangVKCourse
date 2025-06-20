package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type ClientManager struct {
	clients map[int]chan *Event
	mu      sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[int]chan *Event),
	}
}

func (cm *ClientManager) AddClient(ch chan *Event) int {
	id := int(uuid.New().ID())
	cm.mu.Lock()
	cm.clients[id] = ch
	cm.mu.Unlock()
	return id
}

func (cm *ClientManager) RemoveClient(id int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, ok := cm.clients[id]; ok {
		close(ch)
		delete(cm.clients, id)
	}
}

func (cm *ClientManager) Broadcast(event *Event) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, ch := range cm.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

type AdminService struct {
	UnimplementedAdminServer

	loggingManager *ClientManager
	statsManager   *ClientManager
}

func NewAdminService() *AdminService {
	return &AdminService{
		loggingManager: NewClientManager(),
		statsManager:   NewClientManager(),
	}
}

func (a *AdminService) Logging(_ *Nothing, stream Admin_LoggingServer) error {
	buffChSize := 1000
	ch := make(chan *Event, buffChSize)

	id := a.loggingManager.AddClient(ch)
	defer a.loggingManager.RemoveClient(id)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev := <-ch:
			if err := stream.Send(ev); err != nil {
				return err
			}
		}
	}
}

func (a *AdminService) Statistics(statInterval *StatInterval, stream Admin_StatisticsServer) error {
	interval := time.Duration(statInterval.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ch := make(chan *Event)

	id := a.statsManager.AddClient(ch)
	defer a.statsManager.RemoveClient(id)

	stats := &Stat{
		ByMethod:   map[string]uint64{},
		ByConsumer: map[string]uint64{},
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil

		case ev := <-ch:
			stats.ByMethod[ev.Method]++
			stats.ByConsumer[ev.Consumer]++

		case <-ticker.C:
			if err := stream.Send(stats); err != nil {
				return err
			}
			stats = &Stat{
				ByMethod:   map[string]uint64{},
				ByConsumer: map[string]uint64{},
			}
		}
	}
}

func (a *AdminService) broadcast(event *Event) {
	a.loggingManager.Broadcast(event)
	a.statsManager.Broadcast(event)
}
