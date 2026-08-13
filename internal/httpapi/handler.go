package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/wgt-system/conveyance/internal/currentobject"
)

type Handler struct {
	service   *currentobject.Service
	operation currentobject.OperationContext
}

func NewHandler(service *currentobject.Service, operation currentobject.OperationContext) http.Handler {
	handler := &Handler{service: service, operation: operation}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current", handler.getCurrent)
	mux.HandleFunc("PUT /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current", handler.putCurrent)
	return mux
}

type envelopeJSON struct {
	EnvelopeFormatVersion uint64  `json:"envelope_format_version"`
	ChannelEpoch          uint64  `json:"channel_epoch"`
	Revision              uint64  `json:"revision"`
	EnvelopeRef           string  `json:"envelope_ref"`
	PreviousEnvelopeRef   *string `json:"previous_envelope_ref"`
	ProtectedPayload      string  `json:"protected_payload"`
}

type errorJSON struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func (handler *Handler) getCurrent(response http.ResponseWriter, request *http.Request) {
	trust, channel, ok := parseRoute(response, request)
	if !ok {
		return
	}
	envelope, err := handler.service.GetCurrent(request.Context(), handler.operation, trust, channel)
	if err != nil {
		writeApplicationError(response, err)
		return
	}
	writeEnvelope(response, http.StatusOK, envelope)
}

func (handler *Handler) putCurrent(response http.ResponseWriter, request *http.Request) {
	trust, channel, ok := parseRoute(response, request)
	if !ok {
		return
	}
	envelope, err := decodeEnvelope(request.Body)
	if err != nil {
		writeError(response, http.StatusBadRequest, "invalid_envelope")
		return
	}
	result, err := handler.service.PutCurrent(request.Context(), handler.operation, trust, channel, envelope)
	if err != nil {
		writeApplicationError(response, err)
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	writeEnvelope(response, status, result.Envelope)
}

func parseRoute(response http.ResponseWriter, request *http.Request) (currentobject.TrustDomainRef, currentobject.ChannelRef, bool) {
	trust, trustErr := currentobject.ParseTrustDomainRef(request.PathValue("trust_domain_ref"))
	channel, channelErr := currentobject.ParseChannelRef(request.PathValue("channel_ref"))
	if trustErr != nil || channelErr != nil {
		writeError(response, http.StatusBadRequest, "invalid_reference")
		return currentobject.TrustDomainRef{}, currentobject.ChannelRef{}, false
	}
	return trust, channel, true
}

func decodeEnvelope(body io.Reader) (currentobject.Envelope, error) {
	fields, err := decodeObjectFields(body)
	if err != nil {
		return currentobject.Envelope{}, err
	}
	allowed := map[string]bool{
		"envelope_format_version": true, "channel_epoch": true, "revision": true,
		"envelope_ref": true, "previous_envelope_ref": true, "protected_payload": true,
	}
	if len(fields) != len(allowed) {
		return currentobject.Envelope{}, errors.New("invalid field set")
	}
	for name := range fields {
		if !allowed[name] {
			return currentobject.Envelope{}, errors.New("unknown field")
		}
	}
	var formatVersion, epoch, revision uint64
	if err := decodeSingle(fields["envelope_format_version"], &formatVersion); err != nil {
		return currentobject.Envelope{}, err
	}
	if err := decodeSingle(fields["channel_epoch"], &epoch); err != nil {
		return currentobject.Envelope{}, err
	}
	if err := decodeSingle(fields["revision"], &revision); err != nil {
		return currentobject.Envelope{}, err
	}
	var envelopeRefText string
	if err := decodeSingle(fields["envelope_ref"], &envelopeRefText); err != nil {
		return currentobject.Envelope{}, err
	}
	envelopeRef, err := currentobject.ParseEnvelopeRef(envelopeRefText)
	if err != nil {
		return currentobject.Envelope{}, err
	}
	var previous *currentobject.EnvelopeRef
	previousRaw := fields["previous_envelope_ref"]
	if string(bytes.TrimSpace(previousRaw)) != "null" {
		var previousText string
		if err := decodeSingle(previousRaw, &previousText); err != nil {
			return currentobject.Envelope{}, err
		}
		parsed, err := currentobject.ParseEnvelopeRef(previousText)
		if err != nil {
			return currentobject.Envelope{}, err
		}
		previous = &parsed
	}
	var payloadText string
	if err := decodeSingle(fields["protected_payload"], &payloadText); err != nil {
		return currentobject.Envelope{}, err
	}
	payload, err := base64.StdEncoding.DecodeString(payloadText)
	if err != nil || base64.StdEncoding.EncodeToString(payload) != payloadText {
		return currentobject.Envelope{}, errors.New("invalid base64")
	}
	return currentobject.NewEnvelope(formatVersion, epoch, revision, envelopeRef, previous, payload), nil
}

func decodeObjectFields(body io.Reader) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(body)
	start, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := start.(json.Delim)
	if !ok || delim != '{' {
		return nil, errors.New("envelope must be an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok {
			return nil, errors.New("invalid object key")
		}
		if _, exists := fields[name]; exists {
			return nil, errors.New("duplicate field")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if err := requireEnd(decoder); err != nil {
		return nil, err
	}
	return fields, nil
}

func decodeSingle(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return requireEnd(decoder)
}

func requireEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON")
		}
		return err
	}
	return nil
}

func writeEnvelope(response http.ResponseWriter, status int, envelope currentobject.Envelope) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(envelopeJSONFrom(envelope))
}

func envelopeJSONFrom(envelope currentobject.Envelope) envelopeJSON {
	var previous *string
	if envelope.PreviousEnvelopeRef != nil {
		value := envelope.PreviousEnvelopeRef.String()
		previous = &value
	}
	return envelopeJSON{envelope.EnvelopeFormatVersion, envelope.ChannelEpoch, envelope.Revision, envelope.EnvelopeRef.String(), previous, base64.StdEncoding.EncodeToString(envelope.ProtectedPayload)}
}

func writeApplicationError(response http.ResponseWriter, err error) {
	status, code := http.StatusServiceUnavailable, "unavailable"
	switch {
	case errors.Is(err, currentobject.ErrInvalidEnvelope):
		status, code = http.StatusBadRequest, "invalid_envelope"
	case errors.Is(err, currentobject.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, currentobject.ErrCurrentObjectNotFound):
		status, code = http.StatusNotFound, "current_object_not_found"
	case errors.Is(err, currentobject.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, currentobject.ErrPayloadTooLarge):
		status, code = http.StatusRequestEntityTooLarge, "payload_too_large"
	case errors.Is(err, currentobject.ErrUnsupportedEnvelopeFormat):
		status, code = http.StatusUnprocessableEntity, "unsupported_envelope_format"
	}
	writeError(response, status, code)
}

func writeError(response http.ResponseWriter, status int, code string) {
	response.Header().Set("Content-Type", "application/json")
	var body errorJSON
	body.Error.Code = code
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
