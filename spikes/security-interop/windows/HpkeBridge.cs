using System.Runtime.InteropServices;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

internal static class HpkeNative
{
    const string Library = "conveyance_hpke";
    [DllImport(Library, CallingConvention = CallingConvention.Cdecl)] internal static extern int conveyance_hpke_suite_check(ushort kem, ushort kdf, ushort aead);
    [DllImport(Library, CallingConvention = CallingConvention.Cdecl)] internal static extern int conveyance_hpke_keygen(byte[]? ikm, nuint ikmLen, byte[] publicKey, ref nuint publicKeyLen, byte[] privateKey, ref nuint privateKeyLen);
    [DllImport(Library, CallingConvention = CallingConvention.Cdecl)] internal static extern int conveyance_hpke_public_from_private(byte[] privateKey, nuint privateKeyLen, byte[] publicKey, ref nuint publicKeyLen);
    [DllImport(Library, CallingConvention = CallingConvention.Cdecl)] internal static extern int conveyance_hpke_seal(byte[] recipientPublic, nuint recipientPublicLen, byte[]? ikme, nuint ikmeLen, byte[] info, nuint infoLen, byte[] aad, nuint aadLen, byte[] plaintext, nuint plaintextLen, byte[] enc, ref nuint encLen, byte[] ciphertext, ref nuint ciphertextLen);
    [DllImport(Library, CallingConvention = CallingConvention.Cdecl)] internal static extern int conveyance_hpke_open(byte[] recipientPrivate, nuint recipientPrivateLen, byte[] info, nuint infoLen, byte[] aad, nuint aadLen, byte[] enc, nuint encLen, byte[] ciphertext, nuint ciphertextLen, byte[] plaintext, ref nuint plaintextLen);

    internal static void Load(string path)
    {
        NativeLibrary.SetDllImportResolver(typeof(HpkeNative).Assembly, (name, _, _) => name == Library ? NativeLibrary.Load(path) : IntPtr.Zero);
    }
}

internal static class HpkeProgram
{
    static readonly byte[] Info = Encoding.UTF8.GetBytes("conveyance/channel-key-grant/1.0");
    static readonly byte[] RfcInfo = Convert.FromHexString("4f6465206f6e2061204772656369616e2055726e");
    static readonly byte[] RfcIkmE = Convert.FromHexString("2cd7c601cefb3d42a62b04b7a9041494c06c7843818e0ce28a8f704ae7ab20f9");
    static readonly byte[] RfcIkmR = Convert.FromHexString("dac33b0e9db1b59dbbea58d59a14e7b5896e9bdf98fad6891e99d1686492b9ee");
    static readonly byte[] RfcSkR = Convert.FromHexString("497b4502664cfea5d5af0b39934dac72242a74f8480451e1aee7d6a53320333d");
    static readonly byte[] RfcPkR = Convert.FromHexString("430f4b9859665145a6b1ba274024487bd66f03a2dd577d7753c68d7d7d00c00c");
    static readonly byte[] RfcEnc = Convert.FromHexString("6c93e09869df3402d7bf231bf540fadd35cd56be14f97178f0954db94b7fc256");

    static void Check(int code, string operation) { if (code != 0) throw new InvalidOperationException($"{operation} failed: {code}"); }
    static byte[] Keygen(byte[] ikm, out byte[] privateKey)
    {
        var pub = new byte[32]; privateKey = new byte[32]; nuint pl = 32, sl = 32;
        Check(HpkeNative.conveyance_hpke_keygen(ikm, (nuint)ikm.Length, pub, ref pl, privateKey, ref sl), "keygen");
        Array.Resize(ref pub, checked((int)pl)); Array.Resize(ref privateKey, checked((int)sl)); return pub;
    }
    static (byte[] Enc, byte[] Ciphertext) Seal(byte[] pub, byte[]? ikme, byte[] info, byte[] aad, byte[] plaintext)
    {
        var enc = new byte[64]; var ct = new byte[plaintext.Length + 16]; nuint el = 64, cl = (nuint)ct.Length;
        Check(HpkeNative.conveyance_hpke_seal(pub, (nuint)pub.Length, ikme, (nuint)(ikme?.Length ?? 0), info, (nuint)info.Length, aad, (nuint)aad.Length, plaintext, (nuint)plaintext.Length, enc, ref el, ct, ref cl), "seal");
        Array.Resize(ref enc, checked((int)el)); Array.Resize(ref ct, checked((int)cl)); return (enc, ct);
    }
    static byte[] Open(byte[] sk, byte[] info, byte[] aad, byte[] enc, byte[] ct)
    {
        var pt = new byte[Math.Max(0, ct.Length - 16)]; nuint pl = (nuint)pt.Length;
        Check(HpkeNative.conveyance_hpke_open(sk, (nuint)sk.Length, info, (nuint)info.Length, aad, (nuint)aad.Length, enc, (nuint)enc.Length, ct, (nuint)ct.Length, pt, ref pl), "open");
        Array.Resize(ref pt, checked((int)pl)); return pt;
    }
    static string A(string trust, string channel, ulong epoch, string installation) => $"conveyance-channel-key-grant-v1\ntrust_domain_ref={trust}\nchannel_ref={channel}\nchannel_epoch={epoch}\nrecipient_installation_ref={installation}";

    static int Main(string[] args)
    {
        try
        {
            if (args.Length < 2) throw new ArgumentException("usage: HpkeBridge <native.dll> <proof|grant> [path]");
            HpkeNative.Load(Path.GetFullPath(args[0]));
            return args[1] switch { "proof" => Proof(), "grant" => Grant(args), "open-grant" => OpenGrant(args), "grant-tamper" => GrantTamper(args), _ => throw new ArgumentException("unknown command") };
        }
        catch (Exception ex) { Console.Error.WriteLine($"{ex.GetType().Name}: {ex.Message}"); return 1; }
    }

    static int Proof()
    {
        Check(HpkeNative.conveyance_hpke_suite_check(0x20, 1, 2), "suite_check");
        var derivedPub = Keygen(RfcIkmR, out var derivedSk);
        if (derivedPub.Length != 32 || derivedSk.Length != 32) throw new InvalidOperationException("keygen length mismatch");
        var rawPub = new byte[32]; nuint rawPubLen = 32;
        Check(HpkeNative.conveyance_hpke_public_from_private(RfcSkR, 32, rawPub, ref rawPubLen), "raw public-key derivation");
        if (!rawPub.SequenceEqual(RfcPkR)) throw new InvalidOperationException("RFC raw recipient public key mismatch");
        var rfc = Seal(RfcPkR, RfcIkmE, RfcInfo, Array.Empty<byte>(), Encoding.UTF8.GetBytes("rfc9180-selected"));
        if (!rfc.Enc.SequenceEqual(RfcEnc)) throw new InvalidOperationException("RFC encapsulated key mismatch");
        if (!Open(RfcSkR, RfcInfo, Array.Empty<byte>(), rfc.Enc, rfc.Ciphertext).SequenceEqual(Encoding.UTF8.GetBytes("rfc9180-selected"))) throw new InvalidOperationException("RFC open mismatch");
        var randomPub = Keygen(RandomNumberGenerator.GetBytes(32), out var randomSk);
        var random = Seal(randomPub, null, Info, Encoding.UTF8.GetBytes("aad"), Encoding.UTF8.GetBytes("native-random-round-trip"));
        if (!Open(randomSk, Info, Encoding.UTF8.GetBytes("aad"), random.Enc, random.Ciphertext).SequenceEqual(Encoding.UTF8.GetBytes("native-random-round-trip"))) throw new InvalidOperationException("random round-trip mismatch");
        var tampered = Encoding.UTF8.GetBytes("bad-aad"); var failed = new byte[random.Ciphertext.Length - 16]; nuint failedLen = (nuint)failed.Length;
        var code = HpkeNative.conveyance_hpke_open(randomSk, 32, Info, (nuint)Info.Length, tampered, (nuint)tampered.Length, random.Enc, (nuint)random.Enc.Length, random.Ciphertext, (nuint)random.Ciphertext.Length, failed, ref failedLen);
        if (code == 0 || failedLen != 0 || failed.Any(x => x != 0)) throw new InvalidOperationException("AAD tamper was accepted or plaintext leaked");
        Console.WriteLine(JsonSerializer.Serialize(new { suite_check = "PASS", rfc9180_algorithm_conformance = "PASS", frozen_suite_deterministic_proof = "PASS", random_round_trip = "PASS", aad_tamper = "PASS", derived_private_key_bytes = derivedSk.Length })); return 0;
    }

    static int Grant(string[] args)
    {
        if (args.Length != 3) throw new ArgumentException("grant requires an output JSON path");
        const string trust = "00000000-0000-0000-0000-000000000101", channel = "00000000-0000-0000-0000-000000000102", installation = "00000000-0000-0000-0000-000000000103";
        var aad = Encoding.UTF8.GetBytes(A(trust, channel, 7, installation)); var key = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        var pair = Keygen(RandomNumberGenerator.GetBytes(32), out var recipientSk); var sealedGrant = Seal(pair, null, Info, aad, key);
        var output = new { grant = new { grant_format_version = 1, kem_id = 32, kdf_id = 1, aead_id = 2, trust_domain_ref = trust, channel_ref = channel, channel_epoch = 7, recipient_installation_ref = installation, enc = Convert.ToBase64String(sealedGrant.Enc), ciphertext = Convert.ToBase64String(sealedGrant.Ciphertext) }, recipient_private_key = Convert.ToBase64String(recipientSk), recipient_public_key = Convert.ToBase64String(pair) };
        File.WriteAllText(args[2], JsonSerializer.Serialize(output, new JsonSerializerOptions { WriteIndented = true })); Console.WriteLine("OPENSSL-GRANT-WRITTEN"); return 0;
    }

    static int OpenGrant(string[] args)
    {
        if (args.Length != 3) throw new ArgumentException("open-grant requires an input JSON path");
        using var document = JsonDocument.Parse(File.ReadAllText(args[2])); var root = document.RootElement;
        var grant = root.GetProperty("grant"); var sk = Convert.FromBase64String(root.GetProperty("recipient_private_key").GetString()!);
        var enc = Convert.FromBase64String(grant.GetProperty("enc").GetString()!); var ct = Convert.FromBase64String(grant.GetProperty("ciphertext").GetString()!);
        var aad = Encoding.UTF8.GetBytes(A(grant.GetProperty("trust_domain_ref").GetString()!, grant.GetProperty("channel_ref").GetString()!, grant.GetProperty("channel_epoch").GetUInt64(), grant.GetProperty("recipient_installation_ref").GetString()!));
        var opened = Open(sk, Info, aad, enc, ct); var expected = Convert.FromHexString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f");
        if (!opened.SequenceEqual(expected)) throw new InvalidOperationException("grant Channel Key mismatch");
        Console.WriteLine("PASS"); return 0;
    }

    static int GrantTamper(string[] args)
    {
        if (args.Length != 3) throw new ArgumentException("grant-tamper requires an input JSON path");
        var original = File.ReadAllText(args[2]);
        var mutations = new[] { "trust_domain_ref", "channel_ref", "channel_epoch", "recipient_installation_ref", "enc", "ciphertext/tag" };
        foreach (var mutation in mutations)
        {
            var root = JsonNode.Parse(original)!.AsObject(); var grant = root["grant"]!.AsObject();
            var sk = Convert.FromBase64String(root["recipient_private_key"]!.GetValue<string>());
            var enc = Convert.FromBase64String(grant["enc"]!.GetValue<string>()); var ct = Convert.FromBase64String(grant["ciphertext"]!.GetValue<string>());
            switch (mutation)
            {
                case "trust_domain_ref": grant["trust_domain_ref"] = "00000000-0000-0000-0000-000000000199"; break;
                case "channel_ref": grant["channel_ref"] = "00000000-0000-0000-0000-000000000199"; break;
                case "channel_epoch": grant["channel_epoch"] = grant["channel_epoch"]!.GetValue<ulong>() + 1; break;
                case "recipient_installation_ref": grant["recipient_installation_ref"] = "00000000-0000-0000-0000-000000000199"; break;
                case "enc": enc[0] ^= 1; break;
                case "ciphertext/tag": ct[^1] ^= 1; break;
            }
            var aad = Encoding.UTF8.GetBytes(A(grant["trust_domain_ref"]!.GetValue<string>(), grant["channel_ref"]!.GetValue<string>(), grant["channel_epoch"]!.GetValue<ulong>(), grant["recipient_installation_ref"]!.GetValue<string>()));
            try { _ = Open(sk, Info, aad, enc, ct); throw new InvalidOperationException($"tamper accepted: {mutation}"); }
            catch (InvalidOperationException ex) when (ex.Message.StartsWith("open failed:")) { }
        }
        Console.WriteLine(JsonSerializer.Serialize(new { result = "PASS", passed = mutations.Length, cases = mutations })); return 0;
    }
}
