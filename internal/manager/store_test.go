package manager

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestStore_Add(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		entry Entry
	}{
		{
			name: "test_add_0",
			entry: Entry{
				ID:        uuid.New().String(),
				Title:     "title",
				Password:  "title",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := testProvideMockDeps(t)
			deps.fs.
				EXPECT().
				Open().
				Return(nil, nil).
				AnyTimes()
			deps.fs.
				EXPECT().
				Write(gomock.Any()).
				Return(nil).AnyTimes()

			store := NewStore(deps.fs, NewTxManager())
			if err := store.Add(tc.entry); err != nil {
				t.Fatalf("store add: %v", err)
			}

			entry, ok := store.FindByID(tc.entry.ID)
			if diff := cmp.Diff(true, ok); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(tc.entry, entry); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}
		})
	}
}

func TestStore_DeleteByID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		entry         Entry
		expectedTxLen int
	}{
		{
			name: "test_delete_0",
			entry: Entry{
				ID:        uuid.New().String(),
				Title:     "title",
				Password:  "title",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedTxLen: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := testProvideMockDeps(t)
			deps.fs.
				EXPECT().
				Open().
				Return(nil, nil).
				AnyTimes()
			deps.fs.
				EXPECT().
				Write(gomock.Any()).
				Return(nil).AnyTimes()

			store := NewStore(deps.fs, NewTxManager())
			if err := store.Add(tc.entry); err != nil {
				t.Fatalf("store add: %v", err)
			}

			entry, ok := store.FindByID(tc.entry.ID)
			if diff := cmp.Diff(true, ok); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(tc.entry, entry); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if err := store.DeleteByID(tc.entry.ID); err != nil {
				t.Fatalf("store delete: %v", err)
			}

			_, ok = store.FindByID(tc.entry.ID)
			if diff := cmp.Diff(false, ok); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(tc.expectedTxLen, len(store.txManager.List())); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}
		})
	}
}

func TestStore_Change(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		entry Entry

		changedPassword string
		expectedTxLen   int
	}{
		{
			name: "test_delete_0",
			entry: Entry{
				ID:        uuid.New().String(),
				Title:     "title",
				Password:  "title",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},

			changedPassword: "title 1",
			expectedTxLen:   2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := testProvideMockDeps(t)
			deps.fs.
				EXPECT().
				Open().
				Return(nil, nil).
				AnyTimes()
			deps.fs.
				EXPECT().
				Write(gomock.Any()).
				Return(nil).AnyTimes()

			store := NewStore(deps.fs, NewTxManager())
			if err := store.Add(tc.entry); err != nil {
				t.Fatalf("store add: %v", err)
			}

			entry, ok := store.FindByID(tc.entry.ID)
			if diff := cmp.Diff(true, ok); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(tc.entry, entry); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if err := store.ChangeByID(
				tc.entry.ID, ChangeEntry{
					Password: &tc.changedPassword,
				},
			); err != nil {
				t.Fatalf("store change by id: %v", err)
			}

			entry, ok = store.FindByID(tc.entry.ID)
			if diff := cmp.Diff(true, ok); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}

			if diff := cmp.Diff(
				Entry{
					ID:        tc.entry.ID,
					Title:     tc.entry.Title,
					Password:  tc.changedPassword,
					CreatedAt: tc.entry.CreatedAt,
					UpdatedAt: entry.UpdatedAt,
				}, entry); diff != "" {
				t.Errorf("diff (+got, -want): %s", diff)
			}
		})
	}
}

type mockDeps struct {
	ctrl *gomock.Controller
	fs   *MockCryptFS
}

func testProvideMockDeps(t *testing.T) mockDeps {
	var deps mockDeps

	deps.ctrl = gomock.NewController(t)
	deps.fs = NewMockCryptFS(deps.ctrl)

	return deps
}
