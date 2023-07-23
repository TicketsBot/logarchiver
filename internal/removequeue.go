package internal

import (
	"errors"
	"github.com/TicketsBot/common/collections"
	"go.uber.org/zap"
	"sync"
	"time"
)

type (
	RemoveQueue struct {
		logger *zap.Logger
		queue  map[uint64]RemoveOperation
		mu     sync.RWMutex
	}

	RemoveOperation struct {
		Status      Status                   `json:"status"`
		Removed     *collections.Set[string] `json:"removed"` // Object names
		Failed      *collections.Set[string] `json:"failed"`  // Object names
		Errors      map[string]error         `json:"errors"`
		LastUpdated time.Time                `json:"-"`
	}

	Status string
)

const (
	StatusInProgress Status = "in_progress"
	StatusComplete   Status = "complete"
	StatusFailed     Status = "failed"
)

var (
	ErrOperationNotFound   = errors.New("remove state for guild not found")
	ErrOperationInProgress = errors.New("operation already in progress")
)

func NewRemoveQueue(logger *zap.Logger) RemoveQueue {
	return RemoveQueue{
		logger: logger,
		queue:  make(map[uint64]RemoveOperation),
	}
}

func (q *RemoveQueue) StartOperation(guildId uint64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.queue[guildId]; ok {
		return ErrOperationInProgress
	}

	q.queue[guildId] = RemoveOperation{
		Status:      StatusInProgress,
		Removed:     collections.NewSet[string](),
		Failed:      collections.NewSet[string](),
		Errors:      make(map[string]error),
		LastUpdated: time.Now(),
	}

	return nil
}

func (q *RemoveQueue) GetOperation(guildId uint64) (RemoveOperation, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if operation, ok := q.queue[guildId]; ok {
		return operation, nil
	} else {
		return RemoveOperation{}, ErrOperationNotFound
	}
}

func (q *RemoveQueue) Status(guildId uint64) (Status, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if operation, ok := q.queue[guildId]; ok {
		return operation.Status, nil
	} else {
		return "", ErrOperationNotFound
	}
}

func (q *RemoveQueue) SetStatus(guildId uint64, status Status) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if operation, ok := q.queue[guildId]; ok {
		operation.Status = status
		operation.LastUpdated = time.Now()
		q.queue[guildId] = operation
	}
}

func (q *RemoveQueue) AddRemovedObject(guildId uint64, objectName string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if operation, ok := q.queue[guildId]; ok {
		operation.Removed.Add(objectName)
		operation.LastUpdated = time.Now()
		q.queue[guildId] = operation
	}
}

func (q *RemoveQueue) AddError(guildId uint64, objectName string, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if operation, ok := q.queue[guildId]; ok {
		operation.Errors[objectName] = err
		operation.Failed.Add(objectName)
		operation.Removed.Remove(objectName)
		operation.LastUpdated = time.Now()
		q.queue[guildId] = operation
	}
}

func (q *RemoveQueue) StartReaper() {
	ticker := time.NewTicker(time.Minute)

	for {
		<-ticker.C

		q.logger.Debug("Running reaper")

		q.mu.Lock()
		for guildId, operation := range q.queue {
			if operation.LastUpdated.Before(time.Now().Add(-10 * time.Minute)) {
				if operation.Status == StatusInProgress {
					q.logger.Warn(
						"Removing in-progress operation from queue due to lack of updates",
						zap.Uint64("guild", guildId),
						zap.String("status", string(operation.Status)),
					)
				} else {
					q.logger.Debug(
						"Removing completed operation from queue",
						zap.Uint64("guild", guildId),
						zap.String("status", string(operation.Status)),
					)
				}

				delete(q.queue, guildId)
			}
		}
		q.mu.Unlock()
	}
}
