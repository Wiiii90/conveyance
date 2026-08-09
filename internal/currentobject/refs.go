package currentobject

import (
	"encoding/hex"
	"fmt"
)

const uuidTextLength = 36

type uuidRef struct {
	bytes [16]byte
}

func parseUUIDRef(value string) (uuidRef, error) {
	var ref uuidRef
	if len(value) != uuidTextLength || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return ref, ErrInvalidReference
	}
	compact := make([]byte, 0, 32)
	for i := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		c := value[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return ref, ErrInvalidReference
		}
		compact = append(compact, c)
	}
	if _, err := hex.Decode(ref.bytes[:], compact); err != nil {
		return uuidRef{}, ErrInvalidReference
	}
	return ref, nil
}

func (ref uuidRef) String() string {
	encoded := hex.EncodeToString(ref.bytes[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:])
}

type TrustDomainRef struct{ value uuidRef }

func ParseTrustDomainRef(value string) (TrustDomainRef, error) {
	ref, err := parseUUIDRef(value)
	return TrustDomainRef{value: ref}, err
}

func (ref TrustDomainRef) String() string { return ref.value.String() }

type ChannelRef struct{ value uuidRef }

func ParseChannelRef(value string) (ChannelRef, error) {
	ref, err := parseUUIDRef(value)
	return ChannelRef{value: ref}, err
}

func (ref ChannelRef) String() string { return ref.value.String() }

type EnvelopeRef struct{ value uuidRef }

func ParseEnvelopeRef(value string) (EnvelopeRef, error) {
	ref, err := parseUUIDRef(value)
	return EnvelopeRef{value: ref}, err
}

func (ref EnvelopeRef) String() string { return ref.value.String() }
