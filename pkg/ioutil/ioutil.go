package ioutil

import (
	"github.com/joakim-ribier/go-utils/pkg/cryptosutil"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
)

// Load loads data from {filename}, decryptes (if secret is defined) and decodes them to a {T} type.
func Load[T any](filename, secret string) (T, error) {
	var out T

	bytes, err := iosutil.Load(filename)
	if err != nil {
		return out, err
	}

	if secret != "" {
		if decrypted, err := cryptosutil.Decrypt(bytes, secret); err != nil {
			return out, err
		} else {
			bytes = decrypted
		}
	}

	out, err = jsonsutil.Unmarshal[T](bytes)
	if err != nil {
		return out, err
	}

	return out, nil
}

// Write encodes {data} to JSON, encryptes (if secret is defined) and writes them in {filename}.
func Write[T any](data T, filename, secret string) error {
	bytes, err := jsonsutil.Marshal[T](data)
	if err != nil {
		return err
	}

	if secret != "" {
		if encrypted, err := cryptosutil.Encrypt(bytes, secret); err != nil {
			return err
		} else {
			bytes = encrypted
		}
	}

	return iosutil.Write(bytes, filename)
}
