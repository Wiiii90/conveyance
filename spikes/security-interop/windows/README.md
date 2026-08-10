# Windows spike client

The client uses only .NET 10 platform APIs. `create` persists an ECDSA P-256
CNG key in the CurrentUser key store and adds its client-authentication
certificate to CurrentUser/My. `request` is a separate-process reload and
TLS 1.3 request using that identity. `aes-fixture` independently runs the
frozen AES-GCM fixture. HPKE is deliberately absent because the candidate
audit did not identify an acceptable managed RFC 9180 implementation.
