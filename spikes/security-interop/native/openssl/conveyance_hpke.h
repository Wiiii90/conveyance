#pragma once

#include <stddef.h>
#include <stdint.h>

#ifdef _WIN32
#define CONV_HPKE_API __declspec(dllexport)
#define CONV_HPKE_CALL __cdecl
#else
#define CONV_HPKE_API
#define CONV_HPKE_CALL
#endif

enum {
    CONV_HPKE_OK = 0,
    CONV_HPKE_BAD_INPUT = 1,
    CONV_HPKE_OPEN_FAILED = 2,
    CONV_HPKE_CRYPTO_FAILED = 3,
    CONV_HPKE_OUTPUT_TOO_SMALL = 4
};

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_suite_check(
    uint16_t kem_id, uint16_t kdf_id, uint16_t aead_id);

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_keygen(
    const uint8_t *ikm, size_t ikm_len,
    uint8_t *public_key, size_t *public_key_len,
    uint8_t *private_key, size_t *private_key_len);

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_public_from_private(
    const uint8_t *private_key, size_t private_key_len,
    uint8_t *public_key, size_t *public_key_len);

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_seal(
    const uint8_t *recipient_public_key, size_t recipient_public_key_len,
    const uint8_t *ikme, size_t ikme_len,
    const uint8_t *info, size_t info_len,
    const uint8_t *aad, size_t aad_len,
    const uint8_t *plaintext, size_t plaintext_len,
    uint8_t *enc, size_t *enc_len,
    uint8_t *ciphertext, size_t *ciphertext_len);

CONV_HPKE_API int CONV_HPKE_CALL conveyance_hpke_open(
    const uint8_t *recipient_private_key, size_t recipient_private_key_len,
    const uint8_t *info, size_t info_len,
    const uint8_t *aad, size_t aad_len,
    const uint8_t *enc, size_t enc_len,
    const uint8_t *ciphertext, size_t ciphertext_len,
    uint8_t *plaintext, size_t *plaintext_len);
