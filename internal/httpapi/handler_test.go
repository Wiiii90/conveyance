package httpapi

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wiiii90/conveyance/internal/currentobject"
	"github.com/Wiiii90/conveyance/internal/persistence/sqlite"
)

const testPath = "/v1/trust-domains/00000000-0000-0000-0000-000000000001/channels/00000000-0000-0000-0000-000000000002/current"

func TestHandlerCurrentObjectLifecycle(t *testing.T) {
	repository := openRepository(t)
	service := currentobject.NewService(repository, currentobject.DefaultMaxPayloadSize)
	handler := NewHandler(service, currentobject.OperationContext{CanRead: true, CanPublish: true})

	response := request(handler, http.MethodGet, testPath, "")
	assertError(t, response, http.StatusNotFound, "current_object_not_found")

	first := `{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AAH/"}`
	response = request(handler, http.MethodPut, testPath, first)
	assertStatus(t, response, http.StatusCreated)
	assertJSON(t, response, first)

	response = request(handler, http.MethodGet, testPath, "")
	assertStatus(t, response, http.StatusOK)
	assertJSON(t, response, first)

	second := `{"envelope_format_version":1,"channel_epoch":1,"revision":2,"envelope_ref":"00000000-0000-0000-0000-000000000004","previous_envelope_ref":"00000000-0000-0000-0000-000000000003","protected_payload":"c2Vjb25k"}`
	response = request(handler, http.MethodPut, testPath, second)
	assertStatus(t, response, http.StatusOK)
	assertJSON(t, response, second)
	response = request(handler, http.MethodGet, testPath, "")
	assertJSON(t, response, second)

	third := `{"envelope_format_version":1,"channel_epoch":2,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000005","previous_envelope_ref":"00000000-0000-0000-0000-000000000004","protected_payload":"dGhpcmQ="}`
	response = request(handler, http.MethodPut, testPath, third)
	assertStatus(t, response, http.StatusOK)

	stale := `{"envelope_format_version":1,"channel_epoch":2,"revision":2,"envelope_ref":"00000000-0000-0000-0000-000000000006","previous_envelope_ref":"00000000-0000-0000-0000-000000000004","protected_payload":"c3RhbGU="}`
	response = request(handler, http.MethodPut, testPath, stale)
	assertError(t, response, http.StatusConflict, "conflict")
	response = request(handler, http.MethodGet, testPath, "")
	assertJSON(t, response, third)
}

func TestHandlerRejectsInvalidRoutesAndEnvelopeStructure(t *testing.T) {
	repository := openRepository(t)
	service := currentobject.NewService(repository, currentobject.DefaultMaxPayloadSize)
	handler := NewHandler(service, currentobject.OperationContext{CanRead: true, CanPublish: true})
	badPath := "/v1/trust-domains/not-a-uuid/channels/00000000-0000-0000-0000-000000000002/current"
	assertError(t, request(handler, http.MethodGet, badPath, ""), http.StatusBadRequest, "invalid_reference")

	cases := []string{
		`{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"bad","previous_envelope_ref":null,"protected_payload":"AA=="}`,
		`{"envelope_format_version":1,"channel_epoch":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA=="}`,
		`{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA==","extra":true}`,
		`{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA=="} {}`,
		`{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA"}`,
	}
	for _, body := range cases {
		assertError(t, request(handler, http.MethodPut, testPath, body), http.StatusBadRequest, "invalid_envelope")
	}
}

func TestHandlerMapsInvalidPositiveNumberFieldsToInvalidEnvelope(t *testing.T) {
	repository := openRepository(t)
	service := currentobject.NewService(repository, currentobject.DefaultMaxPayloadSize)
	handler := NewHandler(service, currentobject.OperationContext{CanPublish: true})
	base := `{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA=="}`
	for _, field := range []string{"envelope_format_version", "channel_epoch", "revision"} {
		body := strings.Replace(base, `"`+field+`":1`, `"`+field+`":0`, 1)
		assertError(t, request(handler, http.MethodPut, testPath, body), http.StatusBadRequest, "invalid_envelope")
	}
}

func TestHandlerRejectsInvalidUint64JSONNumbers(t *testing.T) {
	repository := openRepository(t)
	service := currentobject.NewService(repository, currentobject.DefaultMaxPayloadSize)
	handler := NewHandler(service, currentobject.OperationContext{CanPublish: true})
	base := `{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA=="}`
	cases := []string{
		strings.Replace(base, `"channel_epoch":1`, `"channel_epoch":-1`, 1),
		strings.Replace(base, `"channel_epoch":1`, `"channel_epoch":1.5`, 1),
		strings.Replace(base, `"channel_epoch":1`, `"channel_epoch":18446744073709551616`, 1),
	}
	for _, body := range cases {
		assertError(t, request(handler, http.MethodPut, testPath, body), http.StatusBadRequest, "invalid_envelope")
	}
}

func TestHandlerMapsPermissionsLimitsAndApplicationErrors(t *testing.T) {
	repository := openRepository(t)
	service := currentobject.NewService(repository, 1)
	path := testPath
	body := `{"envelope_format_version":1,"channel_epoch":1,"revision":1,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"QUI="}`
	assertError(t, request(NewHandler(service, currentobject.OperationContext{CanRead: true}), http.MethodPut, path, body), http.StatusForbidden, "forbidden")
	assertError(t, request(NewHandler(service, currentobject.OperationContext{CanPublish: true}), http.MethodGet, path, ""), http.StatusForbidden, "forbidden")
	assertError(t, request(NewHandler(service, currentobject.OperationContext{CanPublish: true}), http.MethodPut, path, body), http.StatusRequestEntityTooLarge, "payload_too_large")

	unsupported := strings.Replace(body, `"envelope_format_version":1`, `"envelope_format_version":2`, 1)
	assertError(t, request(NewHandler(service, currentobject.OperationContext{CanPublish: true}), http.MethodPut, path, unsupported), http.StatusUnprocessableEntity, "unsupported_envelope_format")
	_ = repository.Close()
	assertError(t, request(NewHandler(service, currentobject.OperationContext{CanRead: true}), http.MethodGet, path, ""), http.StatusServiceUnavailable, "unavailable")
}

func TestDecodeEnvelopePreservesUint64(t *testing.T) {
	body := `{"envelope_format_version":1,"channel_epoch":18446744073709551615,"revision":18446744073709551615,"envelope_ref":"00000000-0000-0000-0000-000000000003","previous_envelope_ref":null,"protected_payload":"AA=="}`
	envelope, err := decodeEnvelope(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if envelope.ChannelEpoch != math.MaxUint64 || envelope.Revision != math.MaxUint64 {
		t.Fatalf("uint64 values changed: %#v", envelope)
	}
}

func openRepository(t *testing.T) *sqlite.Repository {
	t.Helper()
	repository, err := sqlite.Open(filepath.Join(t.TempDir(), "conveyance.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	return repository
}

func request(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, want, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q", got)
	}
}

func assertJSON(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var gotValue, wantValue any
	if err := json.Unmarshal(response.Body.Bytes(), &gotValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatal(err)
	}
	gotBytes, _ := json.Marshal(gotValue)
	wantBytes, _ := json.Marshal(wantValue)
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Fatalf("JSON = %s, want %s", gotBytes, wantBytes)
	}
}

func assertError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	assertStatus(t, response, status)
	want := `{"error":{"code":"` + code + `"}}` + "\n"
	if response.Body.String() != want {
		t.Fatalf("error body = %q, want %q", response.Body.String(), want)
	}
}
