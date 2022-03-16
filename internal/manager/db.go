package manager

import (
	"errors"
	"github.com/google/uuid"
	"sync"
)

var ErrRecordNotFound = errors.New("record not found")

type CryptFS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

type db struct {
	fs CryptFS

	mtx    sync.RWMutex
	data   []Entry
	txList txList
}

func (s *db) load() error {
	b, err := s.fs.Open()
	if err != nil {
		return err
	}
	_ = b
	return nil
}

func (s *db) rebuild() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.data = s.data[:0]
	s.txList.each(func(t1 Tx) {
		if t1.Kind == txKindAdd {
			s.data = append(s.data, t1.Payload)

			return
		}

		if t1.Kind == txKindDel {
			for idx, entry := range s.data {
				if entry.ID == t1.Payload.ID {
					s.data = append(s.data[:idx], s.data[idx+1:]...)
				}
			}

			return
		}
	})
}

func (s *db) Add(e Entry) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.add(e)
}

func (s *db) add(e Entry) {
	s.txList.add(makeTx(txKindAdd, e))
	s.rebuild()
}

func (s *db) Fetch(id uuid.UUID) (Entry, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.fetch(id)
}

func (s *db) fetch(id uuid.UUID) (Entry, error) {
	for _, entry := range s.data {
		if entry.ID == id {
			return entry, nil
		}
	}

	return Entry{}, ErrRecordNotFound
}

func (s *db) findByPosition(pos int) (Entry, error) {
	if pos > len(s.data)-1 {
		return Entry{}, ErrRecordNotFound
	}

	return s.data[pos], nil
}

func (s *db) Change(pos int, password string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, err := s.findByPosition(pos)
	if err != nil {
		return err
	}

	return s.change(entry.ID, password)
}

func (s *db) change(id uuid.UUID, password string) error {
	var entry *Entry

	for _, e := range s.data {
		if e.ID == id {
			entry = &e
			break
		}
	}

	if entry == nil {
		return ErrRecordNotFound
	}

	(*entry).Password = password

	s.txList.add(makeTx(txKindDel, *entry))
	s.txList.add(makeTx(txKindAdd, *entry))

	s.rebuild()

	return nil
}
