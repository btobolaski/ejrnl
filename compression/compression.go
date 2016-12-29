package compression

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/btobolaski/ejrnl"
	"github.com/btobolaski/ejrnl/crypto"
)

func CompressAndEncrypt(data, key []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	gWriter := gzip.NewWriter(buffer)
	if _, err := gWriter.Write(data); err != nil {
		return []byte{}, err
	}
	if err := gWriter.Close(); err != nil {
		return []byte{}, err
	}
	return crypto.Encrypt(buffer.Bytes(), key)
}

func DecryptAndDecompress(cyphertext, key []byte) ([]byte, error) {
	raw, err := crypto.Decrypt(cyphertext, key)
	if err != nil {
		return []byte{}, err
	}
	buffer := bytes.NewBuffer(raw)
	gReader, err := gzip.NewReader(buffer)
	if err == nil {
		uncompressed, err := ioutil.ReadAll(gReader)
		// we need to handle the case where the data isn't compressed at all. Therefore, if we
		// successfully uncompress the data, we return it. Otherwise, we test whether the data can be
		// parsed as json.
		if err == nil || err == io.EOF {
			return uncompressed, nil
		}
	}
	entry := &ejrnl.Entry{}
	err2 := json.Unmarshal(raw, &entry)
	if err2 != nil {
		return []byte{}, fmt.Errorf("Failed to decompress the data because %s and failed to parse the raw data as json because %s", err, err2)
	}
	return raw, nil
}
