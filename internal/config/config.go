package config

import "time"

const (
	DefaultServerPort = ":6969"
	DefaultDBPath     = "file:tracker.db?_foreign_keys=on"
	ShutdownTimeout   = 5 * time.Second
)
