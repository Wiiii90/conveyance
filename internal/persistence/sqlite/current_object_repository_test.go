package sqlite

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Wiiii90/conveyance/internal/currentobject"
)

func TestRepositoryPersistsCurrentObjectAndIsIdempotent(t *testing.T) {
	path := t.TempDir() + "\\current-object.db"
	repository := openTestRepository(t, path)
	trust, channel := testRoute(t)
	envelope := testEnvelope(t, 1, 1, "00000000-0000-0000-0000-000000000003", nil, []byte{0, 1, 255})
	if _, err := repository.GetCurrent(context.Background(), trust, channel); !errors.Is(err, currentobject.ErrCurrentObjectNotFound) {
		t.Fatalf("absent GetCurrent error = %v", err)
	}
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, nil, envelope); err != nil {
		t.Fatalf("first CAS error = %v", err)
	}
	assertEnvelopeEqual(t, envelope, getTestCurrent(t, repository, trust, channel))
	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}
	repository = openTestRepository(t, path)
	assertEnvelopeEqual(t, envelope, getTestCurrent(t, repository, trust, channel))
}

func TestRepositoryReplacesCurrentObjectByEpochAndRevision(t *testing.T) {
	repository := openTestRepository(t, t.TempDir()+"\\current-object.db")
	trust, channel := testRoute(t)
	first := testEnvelope(t, 1, 1, "00000000-0000-0000-0000-000000000003", nil, []byte("first"))
	second := testEnvelope(t, 1, 2, "00000000-0000-0000-0000-000000000004", &first.EnvelopeRef, []byte("second"))
	third := testEnvelope(t, 2, 1, "00000000-0000-0000-0000-000000000005", &second.EnvelopeRef, []byte("third"))
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, nil, first); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, &currentobject.CurrentState{ChannelEpoch: 1, Revision: 1, EnvelopeRef: first.EnvelopeRef}, second); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, &currentobject.CurrentState{ChannelEpoch: 1, Revision: 2, EnvelopeRef: second.EnvelopeRef}, third); err != nil {
		t.Fatal(err)
	}
	assertEnvelopeEqual(t, third, getTestCurrent(t, repository, trust, channel))
}

func TestRepositoryCASConflictsOnExistingOrStaleExpectedState(t *testing.T) {
	repository := openTestRepository(t, t.TempDir()+"\\current-object.db")
	trust, channel := testRoute(t)
	first := testEnvelope(t, 1, 1, "00000000-0000-0000-0000-000000000003", nil, []byte("first"))
	second := testEnvelope(t, 1, 2, "00000000-0000-0000-0000-000000000004", &first.EnvelopeRef, []byte("second"))
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, nil, first); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, nil, second); !errors.Is(err, currentobject.ErrConflict) {
		t.Fatalf("expected first-publish conflict, got %v", err)
	}
	stale := []currentobject.CurrentState{
		{ChannelEpoch: 0, Revision: 1, EnvelopeRef: first.EnvelopeRef},
		{ChannelEpoch: 1, Revision: 0, EnvelopeRef: first.EnvelopeRef},
		{ChannelEpoch: 1, Revision: 1, EnvelopeRef: testRef(t, "00000000-0000-0000-0000-000000000099")},
	}
	for _, expected := range stale {
		if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, &expected, second); !errors.Is(err, currentobject.ErrConflict) {
			t.Fatalf("CAS error = %v", err)
		}
		assertEnvelopeEqual(t, first, getTestCurrent(t, repository, trust, channel))
	}
}

func TestRepositoryConcurrentCASHasOneWinnerAndNoHistory(t *testing.T) {
	repository := openTestRepository(t, t.TempDir()+"\\current-object.db")
	trust, channel := testRoute(t)
	first := testEnvelope(t, 1, 1, "00000000-0000-0000-0000-000000000003", nil, []byte("first"))
	left := testEnvelope(t, 1, 2, "00000000-0000-0000-0000-000000000004", &first.EnvelopeRef, []byte("left"))
	right := testEnvelope(t, 1, 2, "00000000-0000-0000-0000-000000000005", &first.EnvelopeRef, []byte("right"))
	if err := repository.CompareAndSwapCurrent(context.Background(), trust, channel, nil, first); err != nil {
		t.Fatal(err)
	}
	expected := &currentobject.CurrentState{ChannelEpoch: 1, Revision: 1, EnvelopeRef: first.EnvelopeRef}
	results := make(chan error, 2)
	var group sync.WaitGroup
	for _, next := range []currentobject.Envelope{left, right} {
		group.Add(1)
		go func(next currentobject.Envelope) {
			defer group.Done()
			results <- repository.CompareAndSwapCurrent(context.Background(), trust, channel, expected, next)
		}(next)
	}
	group.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, currentobject.ErrConflict) {
			t.Fatalf("CAS error = %v", err)
		}
	}
	if winners != 1 {
		t.Fatalf("winners = %d, want 1", winners)
	}
	var count int
	if err := repository.db.QueryRow("SELECT COUNT(*) FROM current_objects").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("current object rows = %d, want 1", count)
	}
}

func openTestRepository(t *testing.T, path string) *Repository {
	t.Helper()
	repository, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	return repository
}
func getTestCurrent(t *testing.T, repository *Repository, trust currentobject.TrustDomainRef, channel currentobject.ChannelRef) currentobject.Envelope {
	t.Helper()
	got, err := repository.GetCurrent(context.Background(), trust, channel)
	if err != nil {
		t.Fatal(err)
	}
	return got
}
func testRoute(t *testing.T) (currentobject.TrustDomainRef, currentobject.ChannelRef) {
	t.Helper()
	return testTrust(t, "00000000-0000-0000-0000-000000000001"), testChannel(t, "00000000-0000-0000-0000-000000000002")
}
func testTrust(t *testing.T, value string) currentobject.TrustDomainRef {
	t.Helper()
	ref, err := currentobject.ParseTrustDomainRef(value)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
func testChannel(t *testing.T, value string) currentobject.ChannelRef {
	t.Helper()
	ref, err := currentobject.ParseChannelRef(value)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
func testRef(t *testing.T, value string) currentobject.EnvelopeRef {
	t.Helper()
	ref, err := currentobject.ParseEnvelopeRef(value)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
func testEnvelope(t *testing.T, epoch, revision uint64, ref string, previous *currentobject.EnvelopeRef, payload []byte) currentobject.Envelope {
	t.Helper()
	return currentobject.NewEnvelope(1, epoch, revision, testRef(t, ref), previous, payload)
}
func assertEnvelopeEqual(t *testing.T, want, got currentobject.Envelope) {
	t.Helper()
	if want.EnvelopeFormatVersion != got.EnvelopeFormatVersion || want.ChannelEpoch != got.ChannelEpoch || want.Revision != got.Revision || want.EnvelopeRef != got.EnvelopeRef || !samePrevious(want.PreviousEnvelopeRef, got.PreviousEnvelopeRef) || string(want.ProtectedPayload) != string(got.ProtectedPayload) {
		t.Fatalf("envelope mismatch: want %#v, got %#v", want, got)
	}
}
func samePrevious(want, got *currentobject.EnvelopeRef) bool {
	if want == nil || got == nil {
		return want == got
	}
	return *want == *got
}
