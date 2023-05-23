package digest

import (
	"os"
	"io/ioutil"
	"testing"
)

func TestCreateCheckSHA256Digest(t *testing.T) {
	content := []byte("test content")
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	tmpfile2, err := ioutil.TempFile("", "test2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile2.Name())

	if _, err := tmpfile2.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal(err)
	}

	var (
		file = tmpfile.Name()
		sameFile = tmpfile2.Name()
		digest = tmpfile.Name()+".sha256"
	)

	testCreateSHA256DigestOK(t, file, digest)
	testCheckSHA256DigestOK(t, file, digest)
	testCheckSHA256DigestOK(t, sameFile, digest)
}

func testCreateSHA256DigestOK(t *testing.T, file, digest string) {
	err := CreateSHA256Digest(file, digest)
	if err != nil {
		t.Errorf("got error %v", err)
	}	
}

func testCheckSHA256DigestOK(t *testing.T, file, digest string) {
ok, err := CheckSHA256Digest(file, digest)
if err != nil {
	t.Errorf("got error %v", err)
}
if !ok {
	t.Error("got verification failed")
}
}