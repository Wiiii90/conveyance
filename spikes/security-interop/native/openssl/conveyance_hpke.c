#include "conveyance_hpke.h"

#include <openssl/evp.h>
#include <openssl/hpke.h>

#include <limits.h>
#include <string.h>

static const OSSL_HPKE_SUITE suite = {
    OSSL_HPKE_KEM_ID_X25519,
    OSSL_HPKE_KDF_ID_HKDF_SHA256,
    OSSL_HPKE_AEAD_ID_AES_GCM_256
};

static int valid_bytes(const uint8_t *value, size_t length)
{
    return length == 0 || value != NULL;
}

static void clear_bytes(uint8_t *value, size_t length)
{
    if (value != NULL && length != 0)
        OPENSSL_cleanse(value, length);
}

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_suite_check(
    uint16_t kem_id, uint16_t kdf_id, uint16_t aead_id)
{
    OSSL_HPKE_SUITE requested = { kem_id, kdf_id, aead_id };
    return OSSL_HPKE_suite_check(requested) == 1 ? CONV_HPKE_OK : CONV_HPKE_CRYPTO_FAILED;
}

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_keygen(
    const uint8_t *ikm, size_t ikm_len,
    uint8_t *public_key, size_t *public_key_len,
    uint8_t *private_key, size_t *private_key_len)
{
    EVP_PKEY *key = NULL;
    size_t pub_len = 32, priv_len = 32;
    int result = CONV_HPKE_CRYPTO_FAILED;

    if (public_key_len == NULL || private_key_len == NULL ||
        public_key == NULL || private_key == NULL ||
        *public_key_len < pub_len || *private_key_len < priv_len ||
        !valid_bytes(ikm, ikm_len) || ikm_len > OSSL_HPKE_MAX_PARMLEN)
        return CONV_HPKE_BAD_INPUT;

    if (OSSL_HPKE_keygen(suite, public_key, &pub_len, &key,
                         ikm, ikm_len, NULL, NULL) != 1 || key == NULL)
        goto done;
    if (EVP_PKEY_get_raw_private_key(key, private_key, &priv_len) != 1)
        goto done;
    *public_key_len = pub_len;
    *private_key_len = priv_len;
    result = CONV_HPKE_OK;

done:
    if (result != CONV_HPKE_OK) {
        clear_bytes(public_key, *public_key_len);
        clear_bytes(private_key, *private_key_len);
        *public_key_len = 0;
        *private_key_len = 0;
    }
    EVP_PKEY_free(key);
    return result;
}

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_seal(
    const uint8_t *recipient_public_key, size_t recipient_public_key_len,
    const uint8_t *ikme, size_t ikme_len,
    const uint8_t *info, size_t info_len,
    const uint8_t *aad, size_t aad_len,
    const uint8_t *plaintext, size_t plaintext_len,
    uint8_t *enc, size_t *enc_len,
    uint8_t *ciphertext, size_t *ciphertext_len)
{
    OSSL_HPKE_CTX *ctx = NULL;
    size_t enc_capacity, ct_capacity;
    int result = CONV_HPKE_CRYPTO_FAILED;

    if (enc_len == NULL || ciphertext_len == NULL || enc == NULL || ciphertext == NULL ||
        recipient_public_key == NULL || recipient_public_key_len != 32 ||
        !valid_bytes(ikme, ikme_len) || ikme_len > OSSL_HPKE_MAX_PARMLEN ||
        !valid_bytes(info, info_len) || info_len > OSSL_HPKE_MAX_INFOLEN ||
        !valid_bytes(aad, aad_len) || !valid_bytes(plaintext, plaintext_len) ||
        plaintext_len > SIZE_MAX - 16)
        return CONV_HPKE_BAD_INPUT;

    enc_capacity = *enc_len;
    ct_capacity = *ciphertext_len;
    if (enc_capacity < 32 || ct_capacity < plaintext_len + 16)
        return CONV_HPKE_OUTPUT_TOO_SMALL;

    ctx = OSSL_HPKE_CTX_new(OSSL_HPKE_MODE_BASE, suite,
                            OSSL_HPKE_ROLE_SENDER, NULL, NULL);
    if (ctx == NULL)
        goto done;
    if (ikme != NULL && OSSL_HPKE_CTX_set1_ikme(ctx, ikme, ikme_len) != 1)
        goto done;
    if (OSSL_HPKE_encap(ctx, enc, enc_len, recipient_public_key,
                        recipient_public_key_len, info, info_len) != 1)
        goto done;
    if (OSSL_HPKE_seal(ctx, ciphertext, ciphertext_len, aad, aad_len,
                       plaintext, plaintext_len) != 1)
        goto done;
    result = CONV_HPKE_OK;

done:
    if (result != CONV_HPKE_OK) {
        clear_bytes(enc, enc_capacity);
        clear_bytes(ciphertext, ct_capacity);
        *enc_len = 0;
        *ciphertext_len = 0;
    }
    OSSL_HPKE_CTX_free(ctx);
    return result;
}

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_public_from_private(
    const uint8_t *private_key, size_t private_key_len,
    uint8_t *public_key, size_t *public_key_len)
{
    EVP_PKEY *key = NULL;
    size_t capacity;
    int result = CONV_HPKE_CRYPTO_FAILED;
    if (private_key == NULL || private_key_len != 32 || public_key == NULL || public_key_len == NULL)
        return CONV_HPKE_BAD_INPUT;
    capacity = *public_key_len;
    if (capacity < 32)
        return CONV_HPKE_OUTPUT_TOO_SMALL;
    key = EVP_PKEY_new_raw_private_key(EVP_PKEY_X25519, NULL, private_key, private_key_len);
    if (key != NULL && EVP_PKEY_get_raw_public_key(key, public_key, public_key_len) == 1)
        result = CONV_HPKE_OK;
    if (result != CONV_HPKE_OK) {
        clear_bytes(public_key, capacity);
        *public_key_len = 0;
    }
    EVP_PKEY_free(key);
    return result;
}

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_open(
    const uint8_t *recipient_private_key, size_t recipient_private_key_len,
    const uint8_t *info, size_t info_len,
    const uint8_t *aad, size_t aad_len,
    const uint8_t *enc, size_t enc_len,
    const uint8_t *ciphertext, size_t ciphertext_len,
    uint8_t *plaintext, size_t *plaintext_len)
{
    EVP_PKEY *key = NULL;
    OSSL_HPKE_CTX *ctx = NULL;
    size_t capacity;
    int result = CONV_HPKE_OPEN_FAILED;

    if (plaintext_len == NULL || plaintext == NULL ||
        recipient_private_key == NULL || recipient_private_key_len != 32 ||
        enc == NULL || enc_len != 32 || ciphertext == NULL || ciphertext_len < 16 ||
        !valid_bytes(info, info_len) || info_len > OSSL_HPKE_MAX_INFOLEN ||
        !valid_bytes(aad, aad_len))
        return CONV_HPKE_BAD_INPUT;
    capacity = *plaintext_len;
    if (capacity < ciphertext_len - 16)
        return CONV_HPKE_OUTPUT_TOO_SMALL;
    clear_bytes(plaintext, capacity);

    key = EVP_PKEY_new_raw_private_key(EVP_PKEY_X25519, NULL,
                                       recipient_private_key, recipient_private_key_len);
    if (key == NULL)
        goto done;
    ctx = OSSL_HPKE_CTX_new(OSSL_HPKE_MODE_BASE, suite,
                            OSSL_HPKE_ROLE_RECEIVER, NULL, NULL);
    if (ctx == NULL)
        goto done;
    if (OSSL_HPKE_decap(ctx, enc, enc_len, key, info, info_len) != 1) {
        result = CONV_HPKE_CRYPTO_FAILED;
        goto done;
    }
    if (OSSL_HPKE_open(ctx, plaintext, plaintext_len, aad, aad_len,
                       ciphertext, ciphertext_len) != 1) {
        *plaintext_len = 0;
        clear_bytes(plaintext, capacity);
        goto done;
    }
    result = CONV_HPKE_OK;

done:
    EVP_PKEY_free(key);
    OSSL_HPKE_CTX_free(ctx);
    if (result != CONV_HPKE_OK) {
        *plaintext_len = 0;
        clear_bytes(plaintext, capacity);
    }
    return result;
}
