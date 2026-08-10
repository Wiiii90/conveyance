package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hpke"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	trustDomainRef = "00000000-0000-0000-0000-000000000101"
	channelRef     = "00000000-0000-0000-0000-000000000102"
	installRef     = "00000000-0000-0000-0000-000000000103"
)

type grant struct {
	GrantFormatVersion       int    `json:"grant_format_version"`
	KEMID                    int    `json:"kem_id"`
	KDFID                    int    `json:"kdf_id"`
	AEADID                   int    `json:"aead_id"`
	TrustDomainRef           string `json:"trust_domain_ref"`
	ChannelRef               string `json:"channel_ref"`
	ChannelEpoch             uint64 `json:"channel_epoch"`
	RecipientInstallationRef string `json:"recipient_installation_ref"`
	Enc                      string `json:"enc"`
	Ciphertext               string `json:"ciphertext"`
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
func b64(b []byte) string            { return base64.StdEncoding.EncodeToString(b) }
func unb64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }

func grantAAD(g grant) []byte {
	return []byte(fmt.Sprintf("conveyance-channel-key-grant-v1\ntrust_domain_ref=%s\nchannel_ref=%s\nchannel_epoch=%d\nrecipient_installation_ref=%s", g.TrustDomainRef, g.ChannelRef, g.ChannelEpoch, g.RecipientInstallationRef))
}

type envelopeFields struct {
	TrustDomainRef, ChannelRef            string
	FormatVersion, ChannelEpoch, Revision uint64
	EnvelopeRef, PreviousEnvelopeRef      string
}

func canonicalEnvelopeAAD(f envelopeFields) []byte {
	return []byte(fmt.Sprintf("conveyance-current-object-envelope-v1\ntrust_domain_ref=%s\nchannel_ref=%s\nenvelope_format_version=%d\nchannel_epoch=%d\nrevision=%d\nenvelope_ref=%s\nprevious_envelope_ref=%s", f.TrustDomainRef, f.ChannelRef, f.FormatVersion, f.ChannelEpoch, f.Revision, f.EnvelopeRef, f.PreviousEnvelopeRef))
}

func fixtureFields() envelopeFields {
	return envelopeFields{trustDomainRef, channelRef, 1, 7, 11, "00000000-0000-0000-0000-000000000105", "00000000-0000-0000-0000-000000000104"}
}

func aesFixture() error {
	key := mustHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	nonce := mustHex("a0a1a2a3a4a5a6a7a8a9aaab")
	plain := []byte(`{"fixture":"conveyance-security-interop-v1"}`)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	out := make([]byte, len(plain)+gcm.Overhead())
	tag := out[len(plain):]
	gcm.Seal(out[:0], nonce, plain, canonicalEnvelopeAAD(fixtureFields()))
	// Seal writes ciphertext and tag together; split the documented framing.
	ciphertext := out[:len(plain)]
	result := map[string]any{"nonce": b64(nonce), "ciphertext": b64(ciphertext), "tag": b64(tag), "aad": string(canonicalEnvelopeAAD(fixtureFields())), "plaintext": string(plain)}
	return json.NewEncoder(os.Stdout).Encode(result)
}

type fixtureRecord struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	Tag        string `json:"tag"`
	AAD        string `json:"aad"`
}

func aesTamper() error {
	key := mustHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	nonce := mustHex("a0a1a2a3a4a5a6a7a8a9aaab")
	plain := []byte(`{"fixture":"conveyance-security-interop-v1"}`)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	base := fixtureFields()
	sealed := gcm.Seal(nil, nonce, plain, canonicalEnvelopeAAD(base))
	ct, tag := sealed[:len(plain)], sealed[len(plain):]
	checks := []struct {
		name   string
		mutate func(envelopeFields, []byte, []byte, []byte) ([]byte, []byte, []byte, []byte)
	}{
		{"trust_domain_ref", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.TrustDomainRef = "00000000-0000-0000-0000-000000000999"
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"channel_ref", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.ChannelRef = "00000000-0000-0000-0000-000000000999"
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"envelope_format_version", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.FormatVersion = 2
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"channel_epoch", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.ChannelEpoch = 8
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"revision", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.Revision = 12
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"envelope_ref", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.EnvelopeRef = "00000000-0000-0000-0000-000000000999"
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"previous_envelope_ref", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			f.PreviousEnvelopeRef = "00000000-0000-0000-0000-000000000999"
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"nonce", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			n = append([]byte(nil), n...)
			n[0] ^= 1
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"ciphertext", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			c = append([]byte(nil), c...)
			c[0] ^= 1
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
		{"tag", func(f envelopeFields, n, c, t []byte) ([]byte, []byte, []byte, []byte) {
			t = append([]byte(nil), t...)
			t[0] ^= 1
			return n, c, t, canonicalEnvelopeAAD(f)
		}},
	}
	passed := 0
	for _, check := range checks {
		n, c, t, aad := check.mutate(base, nonce, ct, tag)
		opened, openErr := gcm.Open(nil, n, append(append([]byte(nil), c...), t...), aad)
		if openErr == nil || opened != nil {
			return fmt.Errorf("tamper accepted: %s", check.name)
		}
		passed++
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "PASS", "passed": passed, "total": len(checks), "canonical_aad_builder": "PASS"})
}

func aesOpen(path string) error {
	var f fixtureRecord
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &f); err != nil {
		return err
	}
	nonce, err := unb64(f.Nonce)
	if err != nil {
		return err
	}
	ciphertext, err := unb64(f.Ciphertext)
	if err != nil {
		return err
	}
	tag, err := unb64(f.Tag)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(mustHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	if _, err = gcm.Open(nil, nonce, append(ciphertext, tag...), []byte(f.AAD)); err != nil {
		return err
	}
	f.Ciphertext = b64(ciphertext)
	f.Tag = b64(tag)
	return json.NewEncoder(os.Stdout).Encode(map[string]string{"open": "PASS"})
}

func randomGrant() error {
	kem, kdf, aead := hpke.DHKEM(ecdh.X25519()), hpke.HKDFSHA256(), hpke.AES256GCM()
	recipient, err := kem.GenerateKey()
	if err != nil {
		return err
	}
	g := grant{1, int(kem.ID()), int(kdf.ID()), int(aead.ID()), trustDomainRef, channelRef, 7, installRef, "", ""}
	enc, sender, err := hpke.NewSender(recipient.PublicKey(), kdf, aead, []byte("conveyance/channel-key-grant/1.0"))
	if err != nil {
		return err
	}
	channelKey := mustHex("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	ct, err := sender.Seal(grantAAD(g), channelKey)
	if err != nil {
		return err
	}
	g.Enc, g.Ciphertext = b64(enc), b64(ct)
	r, err := hpke.NewRecipient(enc, recipient, kdf, aead, []byte("conveyance/channel-key-grant/1.0"))
	if err != nil {
		return err
	}
	opened, err := r.Open(grantAAD(g), ct)
	if err != nil {
		return err
	}
	if string(opened) != string(channelKey) {
		return errors.New("HPKE round trip changed Channel Key")
	}
	tamper := g
	tamper.ChannelRef = "00000000-0000-0000-0000-000000000999"
	if _, err = r.Open(grantAAD(tamper), ct); err == nil {
		return errors.New("tampered grant AAD was accepted")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"grant": g, "round_trip": "PASS", "tamper": "PASS", "recipient_public_key": b64(recipient.PublicKey().Bytes())})
}

func vectorCheck() error {
	// The installed Go testdata is intentionally compact: the selected vector
	// contains the official key/encapsulation material and accumulated outputs.
	// Public crypto/hpke APIs do not expose the test-only deterministic sender
	// hook, so this check proves material parsing and a same-suite AAD round trip.
	kem, kdf, aead := hpke.DHKEM(ecdh.X25519()), hpke.HKDFSHA256(), hpke.AES256GCM()
	priv, err := kem.NewPrivateKey(mustHex("497b4502664cfea5d5af0b39934dac72242a74f8480451e1aee7d6a53320333d"))
	if err != nil {
		return err
	}
	if hex.EncodeToString(priv.PublicKey().Bytes()) != "430f4b9859665145a6b1ba274024487bd66f03a2dd577d7753c68d7d7d00c00c" {
		return errors.New("RFC vector recipient public key mismatch")
	}
	enc := mustHex("6c93e09869df3402d7bf231bf540fadd35cd56be14f97178f0954db94b7fc256")
	if _, err = hpke.NewRecipient(enc, priv, kdf, aead, mustHex("4f6465206f6e2061204772656369616e2055726e")); err != nil {
		return err
	}
	enc2, sender, err := hpke.NewSender(priv.PublicKey(), kdf, aead, []byte("vector-info"))
	if err != nil {
		return err
	}
	recipient, err := hpke.NewRecipient(enc2, priv, kdf, aead, []byte("vector-info"))
	if err != nil {
		return err
	}
	ct, err := sender.Seal([]byte("vector-aad"), []byte("vector-plaintext"))
	if err != nil {
		return err
	}
	pt, err := recipient.Open([]byte("vector-aad"), ct)
	if err != nil || string(pt) != "vector-plaintext" {
		return errors.New("same-suite AAD round trip failed")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"suite": "DHKEM(X25519, HKDF-SHA256)/HKDF-SHA256/AES-256-GCM", "selected_vector_material": "PASS", "official_deterministic_sender": "UNAVAILABLE_PUBLIC_API", "aad_round_trip": "PASS"})
}

func writePEM(path, typ string, der []byte) error {
	return os.WriteFile(path, pemEncode(typ, der), 0600)
}
func pemEncode(typ string, der []byte) []byte {
	return []byte(fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", typ, base64.StdEncoding.EncodeToString(der), typ))
}

func makeServerCert(certPath, keyPath string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	tmpl := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: "conveyance-security-interop-server"}, DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return err
	}
	priv, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	if err = writePEM(certPath, "CERTIFICATE", der); err != nil {
		return err
	}
	if err = writePEM(keyPath, "PRIVATE KEY", priv); err != nil {
		return err
	}
	return os.WriteFile(certPath+".der", der, 0600)
}

func serve(addr, certPath, keyPath, clientCertPath, expectedInstall string) error {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}
	clientDER, err := os.ReadFile(clientCertPath)
	if err != nil {
		return err
	}
	clientCert, err := x509.ParseCertificate(clientDER)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	pool.AddCert(clientCert)
	h := sha256.Sum256(clientCert.RawSubjectPublicKeyInfo)
	registered := hex.EncodeToString(h[:])
	config := &tls.Config{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool}
	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	defer listener.Close()
	actual := listener.Addr().String()
	fmt.Printf("READY %s %s\n", actual, registered)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/spike/mTLS" {
			http.NotFound(w, r)
			return
		}
		ref := r.URL.Query().Get("installation_ref")
		if r.TLS == nil || r.TLS.Version != tls.VersionTLS13 {
			http.Error(w, "tls_version_not_1.3", http.StatusForbidden)
			return
		}
		if len(r.TLS.PeerCertificates) != 1 {
			http.Error(w, "client_certificate_count_invalid", http.StatusForbidden)
			return
		}
		peerHash := sha256.Sum256(r.TLS.PeerCertificates[0].RawSubjectPublicKeyInfo)
		if hex.EncodeToString(peerHash[:]) != registered {
			http.Error(w, "registered_spki_mismatch", http.StatusForbidden)
			return
		}
		if ref != expectedInstall {
			http.Error(w, "installation_ref_mismatch", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tls_version":"TLS 1.3","registered_spki_sha256":"%s","result":"PASS"}`, registered)
	})}
	// ServeTLS cannot be used after tls.Listen; serve accepted TLS connections directly.
	server.ConnState = func(net.Conn, http.ConnState) {}
	return server.Serve(listener)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "commands: aes-fixture hpke vector make-server serve")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "aes-fixture":
		err = aesFixture()
	case "aes-open":
		f := flag.NewFlagSet("aes-open", flag.ExitOnError)
		p := f.String("input", "", "fixture JSON")
		f.Parse(os.Args[2:])
		err = aesOpen(*p)
	case "aes-tamper":
		err = aesTamper()
	case "hpke":
		err = randomGrant()
	case "vector":
		err = vectorCheck()
	case "make-server":
		f := flag.NewFlagSet("make-server", flag.ExitOnError)
		c := f.String("cert", "", "certificate PEM")
		key := f.String("key", "", "private key PEM")
		f.Parse(os.Args[2:])
		err = makeServerCert(*c, *key)
	case "serve":
		f := flag.NewFlagSet("serve", flag.ExitOnError)
		a := f.String("addr", "127.0.0.1:0", "listen address")
		c := f.String("cert", "", "certificate PEM")
		k := f.String("key", "", "private key PEM")
		cc := f.String("client-cert", "", "registered client certificate DER")
		i := f.String("installation-ref", installRef, "expected installation reference")
		f.Parse(os.Args[2:])
		err = serve(*a, *c, *k, *cc, *i)
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
