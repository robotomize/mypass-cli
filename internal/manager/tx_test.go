package manager

import (
	"crypto/sha1"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestTxManager_Serialize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		txs         []Tx
		expectedLen int
	}{
		{
			name:        "test_serialize_0",
			expectedLen: 4,
			txs: []Tx{
				{
					Hash: func() []byte {
						h := sha1.New()
						h.Write([]byte(`test`))

						return h.Sum(nil)
					}(),
					Kind: TxKindAdd,
					Ts:   time.Now().UTC(),
					Payload: Entry{
						ID:        uuid.New().String(),
						Title:     "title title title",
						Password:  "title title",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
				{
					Hash: func() []byte {
						h := sha1.New()
						h.Write([]byte(`test1`))

						return h.Sum(nil)
					}(),
					Kind: TxKindDel,
					Ts:   time.Now().UTC(),
					Payload: Entry{
						ID:        uuid.New().String(),
						Title:     "title1 title1 title1",
						Password:  "title1 title1",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
				{
					Hash: func() []byte {
						h := sha1.New()
						h.Write([]byte(`test2`))

						return h.Sum(nil)
					}(),
					Kind: TxKindAdd,
					Ts:   time.Now().UTC(),
					Payload: Entry{
						ID:        uuid.New().String(),
						Title:     "title2 title2 title2",
						Password:  "title2 title2",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
				{
					Hash: func() []byte {
						h := sha1.New()
						h.Write([]byte(`test3`))

						return h.Sum(nil)
					}(),
					Kind: TxKindAdd,
					Ts:   time.Now().UTC(),
					Payload: Entry{
						ID:        uuid.New().String(),
						Title:     "title3 title3 title3",
						Password:  "title3 title3",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx := NewTxManager()
			tx.txList = append(tx.txList, tc.txs...)
			bytes := tx.Serialize()
			tx.Deserialize(bytes)

			if diff := cmp.Diff(tc.expectedLen, len(tx.txList)); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(tc.txs, tx.txList); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}
		})
	}
}
