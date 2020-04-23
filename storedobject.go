package logarchiver

import "time"

type StoredObject struct {
	Key          string    `json:"key"`
	LastModified time.Time `json:"last_modified"`
}
