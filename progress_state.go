package main

type ProgressState interface {
	Set(entityType string, entityID string, remoteID uint64)
	Get(entityType string, entityID string) (uint64, bool, error)
	Close() error
}
