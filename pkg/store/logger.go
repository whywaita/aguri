package store

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// SyncLoggerMap is logger of sync.Map
type SyncLoggerMap struct {
	s sync.Map
}

// NewSyncLoggerMap create SyncLoggerMap
func NewSyncLoggerMap() *SyncLoggerMap {
	return &SyncLoggerMap{}
}

// Store logger
func (s *SyncLoggerMap) Store(workspace string, logger *logrus.Logger) {
	s.s.Store(workspace, logger)
}

// Load load logger
func (s *SyncLoggerMap) Load(workspace string) (*logrus.Logger, error) {
	v, ok := s.s.Load(workspace)
	if !ok {
		return nil, fmt.Errorf("logger not found")
	}

	return v.(*logrus.Logger), nil
}
