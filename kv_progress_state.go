package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"modernc.org/kv"
)

type KVProgressState struct {
	db *kv.DB
}

func (p KVProgressState) generateKey(entityType string, entityID string) []byte {
	keyName := fmt.Sprintf("%s-%s", entityType, entityID)
	return []byte(keyName)
}

func (p KVProgressState) Set(entityType string, entityID string, remoteID uint64) {
	keyName := p.generateKey(entityType, entityID)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, remoteID)
	p.db.Set(keyName, buf)

}
func (p KVProgressState) Get(entityType string, entityID string) (remoteID uint64, found bool, err error) {
	buf := make([]byte, 8)
	res, err := p.db.Get(buf, p.generateKey(entityType, entityID))

	if err != nil {
		return 0, false, err
	}

	if res == nil {
		return 0, false, nil
	}
	remoteID = binary.LittleEndian.Uint64(res)

	return remoteID, true, nil
}

func (p KVProgressState) Close() error {
	if p.db != nil {
		return p.db.Close()
	}

	return nil
}

func CreateKVProgressState(fileName string) (*KVProgressState, error) {
	var db *kv.DB
	var dberror error
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		db, dberror = kv.Create(fileName, &kv.Options{})
	} else {
		db, dberror = kv.Open(fileName, &kv.Options{})
	}

	if dberror != nil {
		return nil, dberror
	}

	return &KVProgressState{
		db,
	}, nil
}
