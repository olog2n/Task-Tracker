package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
)

type VersionHandler struct {
	metadata *MetadataService
}

func NewVersionHandler(metadata *MetadataService) *VersionHandler {
	return &VersionHandler{metadata: metadata}
}

type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func (h *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Version:   h.metadata.Version(),
		BuildTime: h.metadata.buildTime,
		GoVersion: runtime.Version(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
