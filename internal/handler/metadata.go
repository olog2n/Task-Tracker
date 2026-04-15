package handler

import (
	"time"
)

type MetadataService struct {
	version   string
	buildTime string
	startTime time.Time
}

func NewMetadataService(version, buildTime string) *MetadataService {
	return &MetadataService{
		version:   version,
		buildTime: buildTime,
		startTime: time.Now(),
	}
}

func (s *MetadataService) Version() string {
	return s.version
}

func (s *MetadataService) Uptime() string {
	return time.Since(s.startTime).String()
}

func (s *MetadataService) StartTime() time.Time {
	return s.startTime
}
