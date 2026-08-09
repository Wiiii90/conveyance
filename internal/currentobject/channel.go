package currentobject

type Channel struct {
	trustDomainRef TrustDomainRef
	channelRef     ChannelRef
	current        *Envelope
}

func NewChannel(trustDomainRef TrustDomainRef, channelRef ChannelRef) *Channel {
	return &Channel{trustDomainRef: trustDomainRef, channelRef: channelRef}
}

func (channel *Channel) Identity() (TrustDomainRef, ChannelRef) {
	return channel.trustDomainRef, channel.channelRef
}

func (channel *Channel) Current() (Envelope, bool) {
	if channel.current == nil {
		return Envelope{}, false
	}
	return channel.current.Clone(), true
}

func (channel *Channel) Publish(envelope Envelope) error {
	if err := envelope.validateStructure(); err != nil {
		return err
	}
	if channel.current == nil {
		if envelope.ChannelEpoch != 1 || envelope.Revision != 1 || envelope.PreviousEnvelopeRef != nil {
			return ErrConflict
		}
	} else {
		current := channel.current
		if envelope.PreviousEnvelopeRef == nil || *envelope.PreviousEnvelopeRef != current.EnvelopeRef {
			return ErrConflict
		}
		if envelope.ChannelEpoch == current.ChannelEpoch {
			if envelope.Revision != current.Revision+1 {
				return ErrConflict
			}
		} else if envelope.ChannelEpoch == current.ChannelEpoch+1 {
			if envelope.Revision != 1 {
				return ErrConflict
			}
		} else {
			return ErrConflict
		}
	}

	accepted := envelope.Clone()
	channel.current = &accepted
	return nil
}
