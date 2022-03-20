package manager

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"sync"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/polylab/pollypass-cli/pkg/proto/gen"
)

const (
	TxKindAdd uint8 = 0x0
	TxKindDel uint8 = 0x2
)

type HashFunc func() hash.Hash

var defaultHasher = SHA1()

// SHA1 returns a new hash.Hash computing the SHA1 checksum
func SHA1() func() hash.Hash {
	return func() hash.Hash {
		return sha1.New()
	}
}

// MD5 returns a new hash.Hash computing the MD5 checksum
func MD5() func() hash.Hash {
	return func() hash.Hash {
		return md5.New()
	}
}

type Tx struct {
	Hash    []byte    `json:"hash"`
	Kind    uint8     `json:"kind"`
	Ts      time.Time `json:"ts"`
	Payload Entry     `json:"payload"`
}

func (t Tx) Sha1() string {
	return hex.EncodeToString(t.Hash)
}

type Option func(*Options)

type Options struct {
	hashFunc HashFunc
}

// WithHashFunc set hash func for generate tx hash
func WithHashFunc(f HashFunc) Option {
	return func(options *Options) {
		options.hashFunc = f
	}
}

func NewTxManager(opts ...Option) *TxManager {
	tx := &TxManager{txList: make([]Tx, 0), opts: Options{hashFunc: defaultHasher}}
	for _, o := range opts {
		o(&tx.opts)
	}

	return tx
}

type TxManager struct {
	opts Options

	mtx    sync.RWMutex
	txList []Tx
}

func (t *TxManager) View(hash string) (Tx, bool) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.view(hash)
}

func (t *TxManager) List() []Tx {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	return t.txList
}

func (t *TxManager) Each(f func(t Tx)) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	t.each(f)
}

func (t *TxManager) AddTx(e Entry) error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	return t.addTx(e)
}

func (t *TxManager) DelTx(e Entry) error {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	return t.delTx(e)
}

func (t *TxManager) Deserialize(b []byte) {
	list := gen.GetRootAsTxList(b, 0)
	length := list.ListLength()
	txs := make([]Tx, length)

	for i := 0; i < length; i++ {
		var tx gen.Tx
		if list.List(&tx, i) {
			hashLen := tx.HashLength()
			hashBytes := make([]byte, hashLen)

			for j := 0; j < hashLen; j++ {
				hashBytes[j] = tx.Hash(j)
			}

			var o gen.Entry
			tx.Payload(&o)

			txs[i] = Tx{
				Hash: hashBytes,
				Kind: tx.Kind(),
				Ts:   time.Unix(0, tx.Ts()),
				Payload: Entry{
					ID:        string(o.Id()),
					Title:     string(o.Title()),
					Password:  string(o.Password()),
					CreatedAt: time.Unix(0, o.CreatedAt()),
					UpdatedAt: time.Unix(0, o.UpdatedAt()),
				},
			}
		}
	}

	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.txList = append(t.txList, txs...)
}

func (t *TxManager) Serialize() []byte {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	builder := flatbuffers.NewBuilder(0)
	flatTxs := make([]flatbuffers.UOffsetT, len(t.txList))
	for idx, tx := range t.txList {
		idOffset := builder.CreateString(tx.Payload.ID)
		titleOffset := builder.CreateString(tx.Payload.Title)
		passwordOffset := builder.CreateString(tx.Payload.Password)

		gen.EntryStart(builder)
		gen.EntryAddId(builder, idOffset)
		gen.EntryAddTitle(builder, titleOffset)
		gen.EntryAddPassword(builder, passwordOffset)
		gen.EntryAddCreatedAt(builder, tx.Payload.CreatedAt.UnixNano())
		gen.EntryAddUpdatedAt(builder, tx.Payload.UpdatedAt.UnixNano())

		entry := gen.EntryEnd(builder)

		gen.TxStartHashVector(builder, len(tx.Hash))
		for i := len(tx.Hash) - 1; i >= 0; i-- {
			builder.PrependByte(tx.Hash[i])
		}
		hashOffset := builder.EndVector(len(tx.Hash))

		gen.TxStart(builder)
		gen.TxAddHash(builder, hashOffset)
		gen.TxAddTs(builder, tx.Ts.UnixNano())
		gen.TxAddKind(builder, tx.Kind)
		gen.TxAddPayload(builder, entry)

		flatTxs[idx] = gen.TxEnd(builder)
	}

	gen.TxListStartListVector(builder, len(flatTxs))
	for i := len(flatTxs) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(flatTxs[i])
	}
	endVec := builder.EndVector(len(flatTxs))

	gen.TxListStart(builder)
	gen.TxListAddList(builder, endVec)
	endList := gen.TxListEnd(builder)

	builder.Finish(endList)

	return builder.FinishedBytes()
}

func (t *TxManager) makeTx(kind uint8, e Entry) (Tx, error) {
	ts := time.Now().UTC()

	hashBytes, err := t.generateHash(kind, ts, e)
	if err != nil {
		return Tx{}, fmt.Errorf("generate hash: %w", err)
	}

	return Tx{
		Hash:    hashBytes,
		Kind:    kind,
		Ts:      ts,
		Payload: e,
	}, nil
}

func (t *TxManager) generateHash(kind byte, ts time.Time, e Entry) ([]byte, error) {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)

	buf.WriteByte(kind)

	tsBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuf, uint64(ts.UnixNano()))
	if _, err := buf.Write(tsBuf); err != nil {
		return nil, fmt.Errorf("buf write: %w", err)
	}

	if _, err := buf.Write([]byte(e.Title)); err != nil {
		return nil, fmt.Errorf("buf write: %w", err)
	}

	if _, err := buf.Write([]byte(e.Password)); err != nil {
		return nil, fmt.Errorf("buf write: %w", err)
	}

	binary.LittleEndian.PutUint64(tsBuf, uint64(e.CreatedAt.UnixNano()))
	if _, err := buf.Write(tsBuf); err != nil {
		return nil, fmt.Errorf("buf write: %w", err)
	}

	binary.LittleEndian.PutUint64(tsBuf, uint64(e.UpdatedAt.UnixNano()))
	if _, err := buf.Write(tsBuf); err != nil {
		return nil, fmt.Errorf("buf write: %w", err)
	}

	hasher := t.opts.hashFunc()
	hasher.Write(buf.Bytes())

	return hasher.Sum(nil), nil
}

func (t *TxManager) each(f func(t Tx)) {
	for _, item := range t.txList {
		f(item)
	}
}

func (t *TxManager) view(hash string) (Tx, bool) {
	for _, tx := range t.txList {
		if tx.Sha1() == hash {
			return tx, true
		}
	}

	return Tx{}, false
}

func (t *TxManager) addTx(e Entry) error {
	tx, err := t.makeTx(TxKindAdd, e)
	if err != nil {
		return fmt.Errorf("make tx: %w", err)
	}

	t.txList = append(t.txList, tx)

	return nil
}

func (t *TxManager) delTx(e Entry) error {
	tx, err := t.makeTx(TxKindDel, e)
	if err != nil {
		return fmt.Errorf("make tx: %w", err)
	}

	t.txList = append(t.txList, tx)

	return nil
}
