package currentobject

const SupportedEnvelopeFormatVersion uint64 = 1

type Envelope struct {
	EnvelopeFormatVersion uint64
	ChannelEpoch          uint64
	Revision              uint64
	EnvelopeRef           EnvelopeRef
	PreviousEnvelopeRef   *EnvelopeRef
	ProtectedPayload      []byte
}

func NewEnvelope(formatVersion, epoch, revision uint64, ref EnvelopeRef, previous *EnvelopeRef, payload []byte) Envelope {
	return Envelope{
		EnvelopeFormatVersion: formatVersion,
		ChannelEpoch:          epoch,
		Revision:              revision,
		EnvelopeRef:           ref,
		PreviousEnvelopeRef:   copyEnvelopeRef(previous),
		ProtectedPayload:      append([]byte(nil), payload...),
	}
}

func (envelope Envelope) Clone() Envelope {
	return NewEnvelope(envelope.EnvelopeFormatVersion, envelope.ChannelEpoch, envelope.Revision,
		envelope.EnvelopeRef, envelope.PreviousEnvelopeRef, envelope.ProtectedPayload)
}

func copyEnvelopeRef(ref *EnvelopeRef) *EnvelopeRef {
	if ref == nil {
		return nil
	}
	copy := *ref
	return &copy
}

func (envelope Envelope) validateStructure() error {
	if envelope.EnvelopeFormatVersion == 0 || envelope.ChannelEpoch == 0 || envelope.Revision == 0 {
		return ErrInvalidEnvelope
	}
	if envelope.EnvelopeFormatVersion != SupportedEnvelopeFormatVersion {
		return ErrUnsupportedEnvelopeFormat
	}
	return nil
}
