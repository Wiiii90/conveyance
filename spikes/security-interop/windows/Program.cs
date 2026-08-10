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
                "aes-fixture" => AesFixture(args),
                "aes-open" => AesOpen(args),
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
        request.CertificateExtensions.Add(new X509BasicConstraintsExtension(false, false, 0, false));
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
        Require(args, 5); string dir = args[1]; string logicalName = args[2]; string serverCertPath = args[3]; string url = args[4]; string installationRef = args.Length > 5 ? args[5] : "00000000-0000-0000-0000-000000000103";
        var record = JsonSerializer.Deserialize<CredentialRecord>(File.ReadAllText(Path.Combine(dir, $"{logicalName}.json"))) ?? throw new InvalidOperationException("credential record missing");
        using var store = new X509Store(StoreName.My, StoreLocation.CurrentUser); store.Open(OpenFlags.ReadOnly);
        using X509Certificate2? cert = store.Certificates.Find(X509FindType.FindByThumbprint, record.Thumbprint, false).OfType<X509Certificate2>().FirstOrDefault();
        if (cert is null || !cert.HasPrivateKey) throw new InvalidOperationException("reloaded certificate has no private key");
        using var serverCert = X509CertificateLoader.LoadCertificateFromFile(serverCertPath);
        string expectedSpki = Convert.ToHexString(SHA256.HashData(ExportSpki(serverCert)));
        using var handler = new HttpClientHandler(); handler.ClientCertificates.Add(cert);
        handler.ServerCertificateCustomValidationCallback = (_, presented, _, _) =>
            presented is not null && presented.NotBefore <= DateTime.UtcNow && presented.NotAfter >= DateTime.UtcNow &&
            Convert.ToHexString(SHA256.HashData(ExportSpki(presented))) == expectedSpki;
        using var client = new HttpClient(handler); using var response = client.GetAsync($"{url}?trust_domain_ref=00000000-0000-0000-0000-000000000101&channel_ref=00000000-0000-0000-0000-000000000102&installation_ref={installationRef}").GetAwaiter().GetResult();
        string body = response.Content.ReadAsStringAsync().GetAwaiter().GetResult(); Console.WriteLine(JsonSerializer.Serialize(new { status = (int)response.StatusCode, body }, JsonOptions));
        return response.IsSuccessStatusCode ? 0 : 2;
    }

    static int AesFixture(string[] args)
    {
        Require(args, 2); Directory.CreateDirectory(args[1]);
        byte[] key = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        byte[] nonce = Convert.FromHexString("a0a1a2a3a4a5a6a7a8a9aaab");
        byte[] plaintext = Encoding.UTF8.GetBytes("{\"fixture\":\"conveyance-security-interop-v1\"}");
        string aad = "conveyance-current-object-envelope-v1\ntrust_domain_ref=00000000-0000-0000-0000-000000000101\nchannel_ref=00000000-0000-0000-0000-000000000102\nenvelope_format_version=1\nchannel_epoch=7\nrevision=11\nenvelope_ref=00000000-0000-0000-0000-000000000105\nprevious_envelope_ref=00000000-0000-0000-0000-000000000104";
        byte[] ciphertext = new byte[plaintext.Length], tag = new byte[16]; using (var aes = new AesGcm(key, 16)) aes.Encrypt(nonce, plaintext, ciphertext, tag, Encoding.UTF8.GetBytes(aad));
        var result = new { nonce = Convert.ToBase64String(nonce), ciphertext = Convert.ToBase64String(ciphertext), tag = Convert.ToBase64String(tag), aad, plaintext = Encoding.UTF8.GetString(plaintext) };
        File.WriteAllText(Path.Combine(args[1], "windows-aes.json"), JsonSerializer.Serialize(result, JsonOptions)); Console.WriteLine(JsonSerializer.Serialize(result, JsonOptions)); return 0;
    }

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

    static void Require(string[] args, int count) { if (args.Length < count) throw new ArgumentException("insufficient arguments"); }
    static byte[] ExportSpki(X509Certificate2 certificate)
    {
        using ECDsa? key = certificate.GetECDsaPublicKey();
        return key?.ExportSubjectPublicKeyInfo() ?? throw new InvalidOperationException("certificate is not an ECDSA certificate");
    }
    record CredentialRecord(string KeyName, string Thumbprint, DateTime NotBefore, DateTime NotAfter, string Provider, string CertificateStore, string PrivateKeyExport, string PrivateKeyPersistence);
    record Fixture(string Nonce, string Ciphertext, string Tag, string Aad, string? Plaintext = null);
}
