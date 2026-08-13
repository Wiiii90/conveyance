using System.Net.Http;
using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using System.Text.Json;

static class Program
{
    const string ProviderName = "Microsoft Software Key Storage Provider";
    static readonly JsonSerializerOptions JsonOptions = new() { WriteIndented = true };

    static int Main(string[] args)
    {
        try
        {
            if (args.Length == 0) throw new ArgumentException("commands: create request aes-fixture cleanup");
            return args[0] switch
            {
                "create" => Create(args),
                "request" => Request(args),
                "request-no-cert" => RequestNoCertificate(args),
                "aes-fixture" => AesFixture(args),
                "aes-open" => AesOpen(args),
                "aes-tamper" => AesTamper(),
                "diagnostic" => Diagnostic(args),
                "diagnostic-reopen" => DiagnosticReopen(args),
                "diagnostic-cleanup" => DiagnosticCleanup(args),
                "cleanup" => Cleanup(args),
                _ => throw new ArgumentException($"unknown command {args[0]}")
            };
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"{ex.GetType().FullName}: {ex.Message}\n{ex.StackTrace}");
            return 1;
        }
    }

    static int Create(string[] args)
    {
        Require(args, 3); string dir = args[1]; string logicalName = args[2]; Directory.CreateDirectory(dir);
        string keyName = $"Conveyance.SecurityInterop.{logicalName}.{Guid.NewGuid():N}";
        var creation = new CngKeyCreationParameters
        {
            Provider = new CngProvider(ProviderName),
            KeyCreationOptions = CngKeyCreationOptions.None,
            ExportPolicy = CngExportPolicies.None,
            KeyUsage = CngKeyUsages.Signing
        };
        using CngKey cng = CngKey.Create(CngAlgorithm.ECDsaP256, keyName, creation);
        using ECDsa ecdsa = new ECDsaCng(cng);
        var request = new CertificateRequest($"CN=Conveyance Security Interop {logicalName}", ecdsa, HashAlgorithmName.SHA256);
        request.CertificateExtensions.Add(new X509BasicConstraintsExtension(true, true, 0, false));
        request.CertificateExtensions.Add(new X509KeyUsageExtension(X509KeyUsageFlags.DigitalSignature, false));
        var eku = new OidCollection { new Oid("1.3.6.1.5.5.7.3.2", "TLS Web Client Authentication") };
        request.CertificateExtensions.Add(new X509EnhancedKeyUsageExtension(eku, false));
        using X509Certificate2 cert = request.CreateSelfSigned(DateTimeOffset.UtcNow.AddMinutes(-1), DateTimeOffset.UtcNow.AddHours(1));
        using var store = new X509Store(StoreName.My, StoreLocation.CurrentUser);
        store.Open(OpenFlags.ReadWrite); store.Add(cert);
        string exportResult;
        try { _ = cng.Export(CngKeyBlobFormat.EccPrivateBlob); exportResult = "UNEXPECTED_SUCCESS"; }
        catch (Exception ex) { exportResult = $"FAIL:{ex.GetType().Name}"; }
        var result = new CredentialRecord(keyName, cert.Thumbprint!, cert.NotBefore, cert.NotAfter, ProviderName, "CurrentUser\\My", exportResult, "public certificate only; private-key bytes were not written by the spike");
        File.WriteAllText(Path.Combine(dir, $"{logicalName}.json"), JsonSerializer.Serialize(result, JsonOptions));
        File.WriteAllBytes(Path.Combine(dir, $"{logicalName}.cer"), cert.Export(X509ContentType.Cert));
        Console.WriteLine(JsonSerializer.Serialize(result, JsonOptions));
        return 0;
    }

    static int Request(string[] args)
    {
        try
        {
            Require(args, 5); string dir = args[1]; string logicalName = args[2]; string serverCertPath = args[3]; string url = args[4]; string installationRef = args.Length > 5 ? args[5] : "00000000-0000-0000-0000-000000000103";
            var record = JsonSerializer.Deserialize<CredentialRecord>(File.ReadAllText(Path.Combine(dir, $"{logicalName}.json"))) ?? throw new InvalidOperationException("credential record missing");
            using var store = new X509Store(StoreName.My, StoreLocation.CurrentUser); store.Open(OpenFlags.ReadOnly);
            using X509Certificate2? cert = store.Certificates.Find(X509FindType.FindByThumbprint, record.Thumbprint, false).OfType<X509Certificate2>().FirstOrDefault();
            if (cert is null || !cert.HasPrivateKey) throw new InvalidOperationException("reloaded certificate has no private key");
            using var serverCert = X509CertificateLoader.LoadCertificateFromFile(serverCertPath);
            string expectedSpki = Convert.ToHexString(SHA256.HashData(ExportSpki(serverCert)));
            using var handler = new HttpClientHandler { ClientCertificateOptions = ClientCertificateOption.Manual, SslProtocols = System.Security.Authentication.SslProtocols.Tls13 };
            handler.ClientCertificates.Add(cert);
            handler.ServerCertificateCustomValidationCallback = (_, presented, _, errors) => ValidateServerCertificate(presented, errors, expectedSpki);
            using var client = new HttpClient(handler); using var response = client.GetAsync($"{url}?trust_domain_ref=00000000-0000-0000-0000-000000000101&channel_ref=00000000-0000-0000-0000-000000000102&installation_ref={installationRef}").GetAwaiter().GetResult();
            string body = response.Content.ReadAsStringAsync().GetAwaiter().GetResult(); Console.WriteLine(JsonSerializer.Serialize(new { status = (int)response.StatusCode, body, tls = "TLS1.3" }, JsonOptions));
            return response.IsSuccessStatusCode ? 0 : 2;
        }
        catch (Exception ex) { Console.Error.WriteLine($"request_failure={Describe(ex)}"); return 1; }
    }

    static int RequestNoCertificate(string[] args)
    {
        try
        {
            Require(args, 3); using var serverCert = X509CertificateLoader.LoadCertificateFromFile(args[1]); string expected = Convert.ToHexString(SHA256.HashData(ExportSpki(serverCert)));
            using var handler = new HttpClientHandler { ClientCertificateOptions = ClientCertificateOption.Manual, SslProtocols = System.Security.Authentication.SslProtocols.Tls13 }; handler.ServerCertificateCustomValidationCallback = (_, presented, _, _) => presented is not null && Convert.ToHexString(SHA256.HashData(ExportSpki(presented))) == expected;
            using var client = new HttpClient(handler); using var response = client.GetAsync($"{args[2]}?trust_domain_ref=00000000-0000-0000-0000-000000000101&channel_ref=00000000-0000-0000-0000-000000000102&installation_ref=00000000-0000-0000-0000-000000000103").GetAwaiter().GetResult();
            Console.WriteLine($"status={(int)response.StatusCode}; tls=TLS1.3"); return response.IsSuccessStatusCode ? 0 : 2;
        }
        catch (Exception ex) { Console.Error.WriteLine($"request_failure={Describe(ex)}"); return 1; }
    }

    static int AesFixture(string[] args)
    {
        Require(args, 2); Directory.CreateDirectory(args[1]);
        byte[] key = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        byte[] nonce = Convert.FromHexString("a0a1a2a3a4a5a6a7a8a9aaab");
        byte[] plaintext = Encoding.UTF8.GetBytes("{\"fixture\":\"conveyance-security-interop-v1\"}");
        string aad = CanonicalAad(FixtureFields());
        byte[] ciphertext = new byte[plaintext.Length], tag = new byte[16]; using (var aes = new AesGcm(key, 16)) aes.Encrypt(nonce, plaintext, ciphertext, tag, Encoding.UTF8.GetBytes(aad));
        var result = new { nonce = Convert.ToBase64String(nonce), ciphertext = Convert.ToBase64String(ciphertext), tag = Convert.ToBase64String(tag), aad, plaintext = Encoding.UTF8.GetString(plaintext) };
        File.WriteAllText(Path.Combine(args[1], "windows-aes.json"), JsonSerializer.Serialize(result, JsonOptions)); Console.WriteLine(JsonSerializer.Serialize(result, JsonOptions)); return 0;
    }

    static int AesTamper()
    {
        byte[] key = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"), nonce = Convert.FromHexString("a0a1a2a3a4a5a6a7a8a9aaab"), plaintext = Encoding.UTF8.GetBytes("{\"fixture\":\"conveyance-security-interop-v1\"}");
        var baseFields = FixtureFields(); byte[] aad = Encoding.UTF8.GetBytes(CanonicalAad(baseFields)), ciphertext = new byte[plaintext.Length], tag = new byte[16];
        using (var aes = new AesGcm(key, 16)) aes.Encrypt(nonce, plaintext, ciphertext, tag, aad);
        var checks = new List<(string Name, EnvelopeFields Fields, byte[] Nonce, byte[] Ciphertext, byte[] Tag)>
        {
            ("trust_domain_ref", baseFields with { TrustDomainRef = "00000000-0000-0000-0000-000000000999" }, nonce, ciphertext, tag),
            ("channel_ref", baseFields with { ChannelRef = "00000000-0000-0000-0000-000000000999" }, nonce, ciphertext, tag),
            ("envelope_format_version", baseFields with { FormatVersion = 2 }, nonce, ciphertext, tag),
            ("channel_epoch", baseFields with { ChannelEpoch = 8 }, nonce, ciphertext, tag),
            ("revision", baseFields with { Revision = 12 }, nonce, ciphertext, tag),
            ("envelope_ref", baseFields with { EnvelopeRef = "00000000-0000-0000-0000-000000000999" }, nonce, ciphertext, tag),
            ("previous_envelope_ref", baseFields with { PreviousEnvelopeRef = "00000000-0000-0000-0000-000000000999" }, nonce, ciphertext, tag),
            ("nonce", baseFields, Mutate(nonce), ciphertext, tag),
            ("ciphertext", baseFields, nonce, Mutate(ciphertext), tag),
            ("tag", baseFields, nonce, ciphertext, Mutate(tag))
        };
        foreach (var c in checks)
        {
            byte[] output = new byte[plaintext.Length]; try { using var aes = new AesGcm(key, 16); aes.Decrypt(c.Nonce, c.Ciphertext, c.Tag, output, Encoding.UTF8.GetBytes(CanonicalAad(c.Fields))); throw new InvalidOperationException($"tamper accepted: {c.Name}"); }
            catch (CryptographicException) { if (output.Any(b => b != 0)) throw new InvalidOperationException($"partial plaintext surfaced: {c.Name}"); }
        }
        Console.WriteLine(JsonSerializer.Serialize(new { result = "PASS", passed = checks.Count, total = checks.Count, canonical_aad_builder = "PASS" }, JsonOptions)); return 0;
    }

    static byte[] Mutate(byte[] source) { var copy = source.ToArray(); copy[0] ^= 1; return copy; }

    static EnvelopeFields FixtureFields() => new("00000000-0000-0000-0000-000000000101", "00000000-0000-0000-0000-000000000102", 1, 7, 11, "00000000-0000-0000-0000-000000000105", "00000000-0000-0000-0000-000000000104");
    static string CanonicalAad(EnvelopeFields f) => $"conveyance-current-object-envelope-v1\ntrust_domain_ref={f.TrustDomainRef}\nchannel_ref={f.ChannelRef}\nenvelope_format_version={f.FormatVersion}\nchannel_epoch={f.ChannelEpoch}\nrevision={f.Revision}\nenvelope_ref={f.EnvelopeRef}\nprevious_envelope_ref={f.PreviousEnvelopeRef}";

    static int AesOpen(string[] args)
    {
        Require(args, 2); var f = JsonSerializer.Deserialize<Fixture>(File.ReadAllText(args[1]), new JsonSerializerOptions { PropertyNameCaseInsensitive = true }) ?? throw new InvalidOperationException("fixture missing");
        byte[] key = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        byte[] nonce = Convert.FromBase64String(f.Nonce), ciphertext = Convert.FromBase64String(f.Ciphertext), tag = Convert.FromBase64String(f.Tag), plaintext = new byte[ciphertext.Length];
        using var aes = new AesGcm(key, 16); aes.Decrypt(nonce, ciphertext, tag, plaintext, Encoding.UTF8.GetBytes(f.Aad)); Console.WriteLine("open=PASS"); return 0;
    }

    static int Cleanup(string[] args)
    {
        Require(args, 3); string dir = args[1]; string logicalName = args[2]; var record = JsonSerializer.Deserialize<CredentialRecord>(File.ReadAllText(Path.Combine(dir, $"{logicalName}.json")))!;
        using var store = new X509Store(StoreName.My, StoreLocation.CurrentUser); store.Open(OpenFlags.ReadWrite); foreach (var cert in store.Certificates.Find(X509FindType.FindByThumbprint, record.Thumbprint, false)) store.Remove(cert);
        using var key = CngKey.Open(record.KeyName, new CngProvider(record.Provider)); key.Delete(); Console.WriteLine("cleanup=PASS"); return 0;
    }

    static int Diagnostic(string[] args)
    {
        Require(args, 2); string dir = args[1]; Directory.CreateDirectory(dir); var result = new Dictionary<string, string>();
        try { using var ephemeral = new ECDsaCng(); result["ephemeral_ecdsa_p256"] = "PASS"; _ = ephemeral.SignData(new byte[] { 1 }, HashAlgorithmName.SHA256); } catch (Exception ex) { result["ephemeral_ecdsa_p256"] = Describe(ex); }
        string keyName = $"Conveyance.SecurityInterop.Diagnostic.{Guid.NewGuid():N}"; result["provider"] = ProviderName;
        try { using var key = CngKey.Create(CngAlgorithm.ECDsaP256, keyName, new CngKeyCreationParameters { Provider = new CngProvider(ProviderName), ExportPolicy = CngExportPolicies.None, KeyUsage = CngKeyUsages.Signing }); result["persisted_minimal"] = "PASS"; using var ecdsa = new ECDsaCng(key); _ = ecdsa.SignData(new byte[] { 2 }, HashAlgorithmName.SHA256); result["persisted_signing"] = "PASS"; try { _ = key.Export(CngKeyBlobFormat.EccPrivateBlob); result["non_exportable"] = "UNEXPECTED_SUCCESS"; } catch (Exception ex) { result["non_exportable"] = $"PASS:{Describe(ex)}"; } File.WriteAllText(Path.Combine(dir, "diagnostic-key.json"), JsonSerializer.Serialize(new DiagnosticKey(keyName, ProviderName), JsonOptions)); }
        catch (Exception ex) { result["persisted_minimal"] = Describe(ex); result["persisted_signing"] = "NOT-RUN"; }
        File.WriteAllText(Path.Combine(dir, "diagnostic.json"), JsonSerializer.Serialize(result, JsonOptions)); Console.WriteLine(JsonSerializer.Serialize(result, JsonOptions)); return 0;
    }

    static int DiagnosticReopen(string[] args)
    {
        Require(args, 2); string path = Path.Combine(args[1], "diagnostic-key.json"); if (!File.Exists(path)) { Console.WriteLine("separate_process_reopen=NOT-RUN"); return 0; }
        var d = JsonSerializer.Deserialize<DiagnosticKey>(File.ReadAllText(path))!; try { using var key = CngKey.Open(d.KeyName, new CngProvider(d.Provider)); using var ecdsa = new ECDsaCng(key); _ = ecdsa.SignData(new byte[] { 3 }, HashAlgorithmName.SHA256); Console.WriteLine("separate_process_reopen=PASS"); return 0; } catch (Exception ex) { Console.WriteLine($"separate_process_reopen=FAIL:{Describe(ex)}"); return 0; }
    }

    static int DiagnosticCleanup(string[] args)
    {
        Require(args, 2); string path = Path.Combine(args[1], "diagnostic-key.json"); if (!File.Exists(path)) return 0; var d = JsonSerializer.Deserialize<DiagnosticKey>(File.ReadAllText(path))!; try { using var key = CngKey.Open(d.KeyName, new CngProvider(d.Provider)); key.Delete(); } catch { } return 0;
    }

    static string Describe(Exception ex) => $"{ex.GetType().FullName}; HRESULT=0x{ex.HResult:X8}; message={ex.Message}; inner={DescribeInner(ex.InnerException)}";
    static string DescribeInner(Exception? ex) => ex is null ? "none" : $"{ex.GetType().FullName}; HRESULT=0x{ex.HResult:X8}; message={ex.Message}; inner={DescribeInner(ex.InnerException)}";
    static bool ValidateServerCertificate(X509Certificate2? presented, System.Net.Security.SslPolicyErrors errors, string expectedSpki)
    {
        string actual = presented is null ? "null" : Convert.ToHexString(SHA256.HashData(ExportSpki(presented)));
        bool valid = presented is not null && presented.NotBefore.ToUniversalTime() <= DateTime.UtcNow && presented.NotAfter.ToUniversalTime() >= DateTime.UtcNow && actual == expectedSpki;
        if (!valid) Console.Error.WriteLine($"server_certificate_rejected errors={errors}; subject={presented?.Subject ?? "null"}; actual_spki={actual}; expected_spki={expectedSpki}; not_before={presented?.NotBefore:o}; not_after={presented?.NotAfter:o}");
        return valid;
    }

    static void Require(string[] args, int count) { if (args.Length < count) throw new ArgumentException("insufficient arguments"); }
    static byte[] ExportSpki(X509Certificate2 certificate)
    {
        using ECDsa? key = certificate.GetECDsaPublicKey();
        return key?.ExportSubjectPublicKeyInfo() ?? throw new InvalidOperationException("certificate is not an ECDSA certificate");
    }
    record CredentialRecord(string KeyName, string Thumbprint, DateTime NotBefore, DateTime NotAfter, string Provider, string CertificateStore, string PrivateKeyExport, string PrivateKeyPersistence);
    record Fixture(string Nonce, string Ciphertext, string Tag, string Aad, string? Plaintext = null);
    record EnvelopeFields(string TrustDomainRef, string ChannelRef, int FormatVersion, int ChannelEpoch, int Revision, string EnvelopeRef, string PreviousEnvelopeRef);
    record DiagnosticKey(string KeyName, string Provider);
}
