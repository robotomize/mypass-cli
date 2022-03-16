package manager

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"github.com/google/uuid"
	"sync"
	"time"
)

type EntryList struct {
	Items []Entry
}

const (
	txKindAdd uint8 = 0x0
	txKindDel uint8 = 0x2
)

func makeTx(kind uint8, e Entry) Tx {
	ts := time.Now().UTC()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(ts.UnixNano()))

	return Tx{
		Hash:    sha1.New().Sum(buf),
		Kind:    kind,
		Ts:      ts,
		Payload: e,
	}
}

type Tx struct {
	Hash    []byte
	Kind    uint8
	Ts      time.Time
	Payload Entry
}

func (t Tx) sha1() string {
	return hex.EncodeToString(t.Hash)
}

type txList struct {
	mtx   sync.RWMutex
	items []Tx
}

func (t *txList) each(f func(t1 Tx)) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	for _, item := range t.items {
		f(item)
	}
}

func (t *txList) add(t1 Tx) {
	t.items = append(t.items, t1)
}

type Entry struct {
	ID        uuid.UUID
	Title     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChangeEntry struct {
	Title    string
	Password string
}

type Manager struct {
	db db
}
