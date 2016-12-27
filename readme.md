# ejrnl

[![Build Status](https://travis-ci.org/btobolaski/ejrnl.svg?branch=master)](https://travis-ci.org/btobolaski/ejrnl)

ejrnl is a journaling application for the privacy concious. It stores all of your entries in an
encrypted format.

## Installing

Install the go tool chain and then run `go get -u github.com/btobolaski/ejrnl/cmd/ejrnl`. You can
then update it in the same way.

## Developing

You'll need additional tools:

- make
- [glide][1]

[1]: https://glide.sh/

Make any changes and then run `make` to run the tests and then compile ejrnl.

## Encryption details

ejrnl uses the go standard library implementations whenever possible. ejrnl uses AEAD (GCM
specifically) with AES-128 as the cipher. The key is generated using scrypt from your specified
password and a salt that is generated on your first use of ejrnl. The exact storage format for the
encrypted files is as follows:

`{{nonce}}{{file}}`
