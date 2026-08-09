package currentobject

import (
	"context"
	"errors"
	"testing"
)

const (
	trustDomainText = "00000000-0000-0000-0000-000000000001"
	channelText     = "00000000-0000-0000-0000-000000000002"
	envelope1Text   = "00000000-0000-0000-0000-000000000003"
	envelope2Text   = "00000000-0000-0000-0000-000000000004"
	envelope3Text   = "00000000-0000-0000-0000-000000000005"
)

func mustRefs(t *testing.T) (TrustDomainRef, ChannelRef, EnvelopeRef, EnvelopeRef, EnvelopeRef) {
	t.Helper()
	trust, err := ParseTrustDomainRef(trustDomainText)
	if err != nil {
		t.Fatal(err)
	}
	channel, err := ParseChannelRef(channelText)
	if err != nil {
		t.Fatal(err)
	}
	first, err := ParseEnvelopeRef(envelope1Text)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseEnvelopeRef(envelope2Text)
	if err != nil {
		t.Fatal(err)
	}
	third, err := ParseEnvelopeRef(envelope3Text)
	if err != nil {
		t.Fatal(err)
	}
	return trust, channel, first, second, third
}

func TestReferencesRequireCanonicalUUIDText(t *testing.T) {
	ref, err := ParseEnvelopeRef(envelope1Text)
	if err != nil || ref.String() != envelope1Text {
		t.Fatalf("canonical reference round trip failed: %q, %v", ref.String(), err)
	}
	for _, value := range []string{
		"00000000-0000-0000-0000-00000000000A",
		"000000000000-0000-0000-0000-000000000003",
		"00000000-0000-0000-0000-00000000003",
		"00000000-0000-0000-0000-000000000003 ",
	} {
		if _, err := ParseEnvelopeRef(value); !errors.Is(err, ErrInvalidReference) {
			t.Errorf("ParseEnvelopeRef(%q) error = %v", value, err)
		}
	}
}

func TestEnvelopeOwnsPayload(t *testing.T) {
	trust, channelRef, first, _, _ := mustRefs(t)
	payload := []byte{1, 2, 3}
	envelope := NewEnvelope(1, 1, 1, first, nil, payload)
	payload[0] = 9
	if envelope.ProtectedPayload[0] != 1 {
		t.Fatal("constructor retained caller payload")
	}
	channel := NewChannel(trust, channelRef)
	if err := channel.Publish(envelope); err != nil {
		t.Fatal(err)
	}
	envelope.ProtectedPayload[1] = 8
	current, _ := channel.Current()
	if current.ProtectedPayload[1] != 2 {
		t.Fatal("channel retained caller payload")
	}
	current.ProtectedPayload[2] = 7
	again, _ := channel.Current()
	if again.ProtectedPayload[2] != 3 {
		t.Fatal("current retrieval exposed owned payload")
	}
}

func TestChannelTransitions(t *testing.T) {
	trust, channelRef, firstRef, secondRef, thirdRef := mustRefs(t)
	channel := NewChannel(trust, channelRef)
	first := NewEnvelope(1, 1, 1, firstRef, nil, []byte("one"))
	if err := channel.Publish(first); err != nil {
		t.Fatalf("first publish: %v", err)
	}
	second := NewEnvelope(1, 1, 2, secondRef, &firstRef, []byte("two"))
	if err := channel.Publish(second); err != nil {
		t.Fatalf("same epoch: %v", err)
	}
	third := NewEnvelope(1, 2, 1, thirdRef, &secondRef, []byte("three"))
	if err := channel.Publish(third); err != nil {
		t.Fatalf("epoch advance: %v", err)
	}
	current, ok := channel.Current()
	if !ok || current.EnvelopeRef != thirdRef || string(current.ProtectedPayload) != "three" {
		t.Fatalf("unexpected current: %#v, %v", current, ok)
	}
}

func TestChannelRejectsInvalidTransitions(t *testing.T) {
	trust, channelRef, firstRef, secondRef, thirdRef := mustRefs(t)
	newChannel := func() *Channel {
		channel := NewChannel(trust, channelRef)
		if err := channel.Publish(NewEnvelope(1, 1, 1, firstRef, nil, nil)); err != nil {
			t.Fatal(err)
		}
		return channel
	}
	advancedChannel := func() *Channel {
		channel := newChannel()
		if err := channel.Publish(NewEnvelope(1, 2, 1, secondRef, &firstRef, nil)); err != nil {
			t.Fatal(err)
		}
		return channel
	}
	tests := []struct {
		name     string
		envelope Envelope
	}{
		{"skipped epoch", NewEnvelope(1, 3, 1, secondRef, &firstRef, nil)},
		{"non consecutive revision", NewEnvelope(1, 1, 3, secondRef, &firstRef, nil)},
		{"revision on epoch advance", NewEnvelope(1, 2, 2, secondRef, &firstRef, nil)},
		{"missing previous", NewEnvelope(1, 1, 2, secondRef, nil, nil)},
		{"stale previous", NewEnvelope(1, 1, 2, secondRef, &thirdRef, nil)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := newChannel().Publish(test.envelope); !errors.Is(err, ErrConflict) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	if err := advancedChannel().Publish(NewEnvelope(1, 1, 1, thirdRef, &secondRef, nil)); !errors.Is(err, ErrConflict) {
		t.Fatalf("decreasing epoch error = %v", err)
	}
	empty := NewChannel(trust, channelRef)
	if err := empty.Publish(NewEnvelope(1, 2, 1, firstRef, nil, nil)); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid first epoch error = %v", err)
	}
	if err := empty.Publish(NewEnvelope(1, 1, 2, firstRef, nil, nil)); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid first revision error = %v", err)
	}
	if err := empty.Publish(NewEnvelope(1, 1, 1, firstRef, &secondRef, nil)); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid first previous error = %v", err)
	}
}

type testRepository struct {
	current *Envelope
	getErr  error
	casErr  error
}

func (repo *testRepository) GetCurrent(context.Context, TrustDomainRef, ChannelRef) (Envelope, error) {
	if repo.getErr != nil {
		return Envelope{}, repo.getErr
	}
	if repo.current == nil {
		return Envelope{}, ErrCurrentObjectNotFound
	}
	return repo.current.Clone(), nil
}

func (repo *testRepository) CompareAndSwapCurrent(_ context.Context, _ TrustDomainRef, _ ChannelRef, expected *EnvelopeRef, next Envelope) error {
	if repo.casErr != nil {
		return repo.casErr
	}
	if repo.current == nil {
		if expected != nil {
			return ErrConflict
		}
	} else if expected == nil || *expected != repo.current.EnvelopeRef {
		return ErrConflict
	}
	copy := next.Clone()
	repo.current = &copy
	return nil
}

func testService(t *testing.T, limit int) (*Service, *testRepository, TrustDomainRef, ChannelRef, EnvelopeRef, EnvelopeRef) {
	t.Helper()
	trust, channel, first, second, _ := mustRefs(t)
	repo := &testRepository{}
	return NewService(repo, limit), repo, trust, channel, first, second
}

func TestServiceAuthorizationAndAbsentRead(t *testing.T) {
	service, _, trust, channel, _, _ := testService(t, 100)
	if _, err := service.GetCurrent(context.Background(), OperationContext{}, trust, channel); !errors.Is(err, ErrForbidden) {
		t.Fatalf("read permission error = %v", err)
	}
	if _, err := service.GetCurrent(context.Background(), OperationContext{CanRead: true}, trust, channel); !errors.Is(err, ErrCurrentObjectNotFound) {
		t.Fatalf("absent read error = %v", err)
	}
	if _, err := service.PutCurrent(context.Background(), OperationContext{}, trust, channel, Envelope{}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("publish permission error = %v", err)
	}
}

func TestServicePublishLimitFormatAndConflict(t *testing.T) {
	service, repo, trust, channel, first, second := testService(t, 3)
	tooLarge := NewEnvelope(1, 1, 1, first, nil, []byte("1234"))
	if _, err := service.PutCurrent(context.Background(), OperationContext{CanPublish: true}, trust, channel, tooLarge); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("large payload error = %v", err)
	}
	unsupported := NewEnvelope(2, 1, 1, first, nil, []byte("123"))
	if _, err := service.PutCurrent(context.Background(), OperationContext{CanPublish: true}, trust, channel, unsupported); !errors.Is(err, ErrUnsupportedEnvelopeFormat) {
		t.Fatalf("unsupported format error = %v", err)
	}
	accepted := NewEnvelope(1, 1, 1, first, nil, []byte("123"))
	result, err := service.PutCurrent(context.Background(), OperationContext{CanPublish: true}, trust, channel, accepted)
	if err != nil || !result.Created {
		t.Fatalf("first service publish: result=%#v error=%v", result, err)
	}
	repo.casErr = ErrConflict
	replacement := NewEnvelope(1, 1, 2, second, &first, []byte("ok"))
	if _, err := service.PutCurrent(context.Background(), OperationContext{CanPublish: true}, trust, channel, replacement); !errors.Is(err, ErrConflict) {
		t.Fatalf("repository conflict error = %v", err)
	}
}

func TestConfiguredZeroLimitUsesSafeDefault(t *testing.T) {
	service, _, trust, channel, first, _ := testService(t, 0)
	payload := make([]byte, DefaultMaxPayloadSize+1)
	if _, err := service.PutCurrent(context.Background(), OperationContext{CanPublish: true}, trust, channel, NewEnvelope(1, 1, 1, first, nil, payload)); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("zero limit error = %v", err)
	}
}
