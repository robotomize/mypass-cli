package manager

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrNotFound = errors.New("record not found")

//go:generate mockgen -source=store.go -destination=mocks.go -package=manager

type FS interface {
	Open() ([]byte, error)
	Write(b []byte) error
}

type CipherFS interface {
	FS
	VerifyCipher() error
}

type Entry struct {
	ID        string
	Title     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChangeEntry struct {
	Title    *string
	Password *string
}

func NewStore(fs CipherFS, txManager *TxManager) (*Store, error) {
	s := &Store{fs: fs, txManager: txManager, data: make([]Entry, 0)}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("load tx: %w", err)
	}

	return s, nil
}

type Store struct {
	fs CipherFS

	mtx       sync.RWMutex
	data      []Entry
	txManager *TxManager
}

func (s *Store) Add(e Entry) error {
	if err := s.txManager.AddTx(e); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *Store) DeleteByID(id string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByID(id)
	if !ok {
		return ErrNotFound
	}

	if err := s.txManager.DelTx(entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *Store) DeleteByNumber(pos int) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByNumber(pos)
	if !ok {
		return ErrNotFound
	}

	if err := s.txManager.DelTx(entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *Store) FindByID(id string) (Entry, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.findByID(id)
}

func (s *Store) List() []Entry {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	list := make([]Entry, len(s.data))
	copy(list, s.data)

	return list
}

func (s *Store) FindByNumber(pos int) (Entry, bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	entry, ok := s.findByNumber(pos)
	if !ok {
		return Entry{}, false
	}

	return entry, true
}

func (s *Store) ChangeByPos(pos int, changed ChangeEntry) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByNumber(pos)
	if !ok {
		return ErrNotFound
	}

	if err := s.change(entry.ID, changed); err != nil {
		return fmt.Errorf("change: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *Store) ChangeByID(id string, changed ChangeEntry) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry, ok := s.findByID(id)
	if !ok {
		return ErrNotFound
	}

	if err := s.change(entry.ID, changed); err != nil {
		return fmt.Errorf("change: %w", err)
	}

	s.rebuild()

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

func (s *Store) load() error {
	b, err := s.fs.Open()
	if err != nil {
		return fmt.Errorf("fs load: %w", err)
	}

	s.txManager.Deserialize(b)
	s.rebuild()

	return nil
}

func (s *Store) rebuild() {
	s.data = s.data[:0]
	s.txManager.Each(func(t1 Tx) {
		if t1.Kind == TxKindAdd {
			s.data = append(s.data, t1.Payload)

			return
		}

		if t1.Kind == TxKindDel {
			for idx, entry := range s.data {
				if entry.ID == t1.Payload.ID {
					s.data = append(s.data[:idx], s.data[idx+1:]...)
				}
			}

			return
		}
	})
}

func (s *Store) change(id string, changed ChangeEntry) error {
	var entry *Entry

	for _, e := range s.data {
		if e.ID == id {
			entry = &e
			break
		}
	}

	if entry == nil {
		return ErrNotFound
	}

	if changed.Title != nil {
		entry.Title = *changed.Title
	}

	if changed.Password != nil {
		entry.Password = *changed.Password
	}

	entry.UpdatedAt = time.Now().UTC()

	if err := s.txManager.DelTx(*entry); err != nil {
		return fmt.Errorf("del tx: %w", err)
	}

	if err := s.txManager.AddTx(*entry); err != nil {
		return fmt.Errorf("add tx: %w", err)
	}

	return nil
}

func (s *Store) findByID(id string) (Entry, bool) {
	for _, entry := range s.data {
		if entry.ID == id {
			return entry, true
		}
	}

	return Entry{}, false
}

func (s *Store) findByNumber(pos int) (Entry, bool) {
	if pos > len(s.data)-1 {
		return Entry{}, false
	}

	return s.data[pos], true
}

func (s *Store) sync() error {
	bytes := s.txManager.Serialize()
	if err := s.fs.Write(bytes); err != nil {
		return fmt.Errorf("fs write: %w", err)
	}

	return nil
}
