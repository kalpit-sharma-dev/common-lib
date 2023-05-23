package digest

import (
	"crypto/sha256"
	"hash"
	"io"
	"io/ioutil"
	"os"
)

// CreateSHA256Digest creates a digest for the file using SHA256 hash.
func CreateSHA256Digest(file, digest string) error {
	return CreateDigest(file, digest, sha256.New())
}

// CheckSHA256Digest verifies the file against the digest using SHA256 hash.
func CheckSHA256Digest(file, digest string) (ok bool, err error) {
	return CheckDigest(file, digest, sha256.New())
}

// CreateDigest creates a digest for the file using the h hash.
// Could be wrapped for the particular hash algorythm.
func CreateDigest(file, digest string, h hash.Hash) error {
	var (
		f, d *os.File
		err  error
	)

	if f, err = os.Open(file); err != nil {
		return err
	}
	defer f.Close()

	if d, err = os.Create(digest); err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	if _, err = d.Write(h.Sum(nil)); err != nil {
		return err
	}

	return nil
}

// CheckDigest verifies the file against the digest using the h hash.
// Could be wrapped for the particular hash algorythm.
func CheckDigest(file, digest string, h hash.Hash) (ok bool, err error) {
	var (
		f                    *os.File
		actDigest, expDigest []byte
	)

	if f, err = os.Open(file); err != nil {
		return
	}
	defer f.Close()

	if _, err = io.Copy(h, f); err != nil {
		return
	}

	actDigest = h.Sum(nil)
	if expDigest, err = ioutil.ReadFile(digest); err != nil {
		return
	}

	if len(actDigest) == 0 || len(expDigest) == 0 || len(actDigest) != len(expDigest) {
		return
	}

	for i := range actDigest {
		if actDigest[i] != expDigest[i] {
			return
		}
	}

	return true, nil
}
