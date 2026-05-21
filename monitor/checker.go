package monitor

import (
	"context"
	"net/http"
	"time"

	"github.com/manuelRestrep0/watchdog/model"
	"github.com/manuelRestrep0/watchdog/store"
)

type Monitor struct {
	store    *store.SQLiteStore
	redis    *store.RedisStore
	stoppers map[int64]chan struct{}
}

func New(s *store.SQLiteStore, r *store.RedisStore) *Monitor {
	return &Monitor{
		store:    s,
		redis:    r,
		stoppers: make(map[int64]chan struct{}),
	}
}

func (m *Monitor) Start(target model.Target) {
	stop := make(chan struct{})
	m.stoppers[target.ID] = stop

	go func() {
		ticker := time.NewTicker(time.Duration(target.Interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.check(target)
			case <-stop:
				return
			}
		}
	}()
}

func (m *Monitor) Stop(targetID int64) {
	if stop, ok := m.stoppers[targetID]; ok {
		close(stop)
		delete(m.stoppers, targetID)
	}
}

func (m *Monitor) check(target model.Target) {
	start := time.Now()
	resp, err := http.Get(target.URL)
	latency := time.Since(start).Milliseconds()

	c := &model.Check{
		TargetID:  target.ID,
		Latency:   latency,
		CheckedAt: time.Now(),
	}

	if err != nil {
		c.Ok = false
	} else {
		defer resp.Body.Close()
		c.StatusCode = resp.StatusCode
		c.Ok = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	m.store.SaveCheck(c)
	m.redis.SetLastCheck(context.Background(), c)
}
