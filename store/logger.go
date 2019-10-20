package store

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SyncLoggerMap struct {
	s sync.Map
}

func NewSyncLoggerMap() *SyncLoggerMap {
	return &SyncLoggerMap{}
}

func (s *SyncLoggerMap) Store(workspace string, logger *logrus.Logger) {
	s.s.Store(workspace, logger)
}

func (s *SyncLoggerMap) Load(workspace string) (*logrus.Logger, error) {
	v, ok := s.s.Load(workspace)
	if !ok {
		return nil, errors.New("logger not found")
	}

	return v.(*logrus.Logger), nil
}
