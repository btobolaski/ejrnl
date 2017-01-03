# ejrnl

[![Build Status](https://travis-ci.org/btobolaski/ejrnl.svg?branch=master)](https://travis-ci.org/btobolaski/ejrnl)

ejrnl is a journaling application for the privacy concious. It stores all of your entries in an
encrypted format.

## Installing

Install the go tool chain and then run `go get -u github.com/btobolaski/ejrnl/cmd/ejrnl`. You can
then update it in the same way.

## Usage

Create your journal using ejrnl init. You can create new entries using ejrnl new. It will create a 
temporary document and open your editor. When you're done editing, the file will be encrypted and 
added to your journal. The format is similar to jekyll's, it has a yaml front mattera with the body 
being markdown devided by `---`.

```yaml
id: 12345
date: 2016-12-25T11:23:45Z
tags:
- example
---
This is the post's body
```

The id and date are not required. If they are not specified, an id will automatically be generated and 
the current time will be used for the timestamp. The date must be an iso8601 timestamp. Tags are 
optional. While the intent is that the body will be markdown, it is not currently processed in any 
way.

If you'd like to set a new password, you can use `ejrnl rekey` to decrypt and then reencrypt every file
with the new password. The unencrypted files are never written to disk.

There is also an http server which, you can access using `ejrnl server`. It listens on port 3000 by
default and is protected with basic auth. It is definitely the least secure way to use ejrnl but it is
by far the most convenient

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
