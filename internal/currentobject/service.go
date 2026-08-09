package currentobject

import (
	"context"
	"errors"
)

const DefaultMaxPayloadSize = 8 * 1024 * 1024

type OperationContext struct {
	CanRead    bool
	CanPublish bool
}

type CurrentState struct {
	ChannelEpoch uint64
	Revision     uint64
	EnvelopeRef  EnvelopeRef
}

type CurrentObjectRepository interface {
	GetCurrent(context.Context, TrustDomainRef, ChannelRef) (Envelope, error)
	CompareAndSwapCurrent(context.Context, TrustDomainRef, ChannelRef, *CurrentState, Envelope) error
}

type PutResult struct {
	Envelope Envelope
	Created  bool
}

type Service struct {
	repository     CurrentObjectRepository
	maxPayloadSize int
}

func NewService(repository CurrentObjectRepository, maxPayloadSize int) *Service {
	if maxPayloadSize <= 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	return &Service{repository: repository, maxPayloadSize: maxPayloadSize}
}

func (service *Service) GetCurrent(ctx context.Context, operation OperationContext, trustDomainRef TrustDomainRef, channelRef ChannelRef) (Envelope, error) {
	if !operation.CanRead {
		return Envelope{}, ErrForbidden
	}
	envelope, err := service.repository.GetCurrent(ctx, trustDomainRef, channelRef)
	if err != nil {
		return Envelope{}, err
	}
	return envelope.Clone(), nil
}

func (service *Service) PutCurrent(ctx context.Context, operation OperationContext, trustDomainRef TrustDomainRef, channelRef ChannelRef, envelope Envelope) (PutResult, error) {
	if !operation.CanPublish {
		return PutResult{}, ErrForbidden
	}
	if err := envelope.validateStructure(); err != nil {
		return PutResult{}, err
	}
	if len(envelope.ProtectedPayload) > service.maxPayloadSize {
		return PutResult{}, ErrPayloadTooLarge
	}

	current, err := service.repository.GetCurrent(ctx, trustDomainRef, channelRef)
	var expected *CurrentState
	created := errors.Is(err, ErrCurrentObjectNotFound)
	if err != nil && !created {
		return PutResult{}, err
	}
	channel := NewChannel(trustDomainRef, channelRef)
	if !created {
		channel.current = &current
		expected = &CurrentState{
			ChannelEpoch: current.ChannelEpoch,
			Revision:     current.Revision,
			EnvelopeRef:  current.EnvelopeRef,
		}
	}
	if err := channel.Publish(envelope); err != nil {
		return PutResult{}, err
	}
	accepted, _ := channel.Current()
	if err := service.repository.CompareAndSwapCurrent(ctx, trustDomainRef, channelRef, expected, accepted); err != nil {
		return PutResult{}, err
	}
	return PutResult{Envelope: accepted, Created: created}, nil
}
