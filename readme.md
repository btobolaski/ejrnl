# ejrnl

ejrnl is a journaling application for the privacy concious. It stores all of your entries in an
encrypted format.

## Encryption details

ejrnl uses the go standard library implementations whenever possible. ejrnl uses AEAD (GCM
specifically) with AES-128 as the cipher. The key is generated using scrypt from your specified
password and a salt that is generated on your first use of ejrnl. The exact storage format for the
encrypted files is as follows:

`{{nonce}}{{file}}`
