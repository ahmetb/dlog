package dlog_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/ahmetalpbalkan/dlog"
)

func Test_tooShortForPrefix(t *testing.T) {
	r := dlog.NewReader(strings.NewReader("123"))
	_, err := ioutil.ReadAll(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(err.Error(), "dlog: corrupted prefix") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func Test_corruptPrefixInMiddle(t *testing.T) {
	b := append(msg(1, []byte("Hi!")), []byte{0x1, 0x0, 0x0, 0x0}...)
	r := dlog.NewReader(bytes.NewReader(b))
	_, err := ioutil.ReadAll(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(err.Error(), "dlog: corrupted prefix") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func Test_prefixReadFailure(t *testing.T) {
	r := dlog.NewReader(
		io.MultiReader(
			bytes.NewReader(msg(1, []byte("Hello!"))),
			&badReader{}))
	b, err := ioutil.ReadAll(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if expected := "dlog: error reading prefix: phony error"; err.Error() != expected {
		t.Fatalf("expected: %q got: %v", expected, err)
	}
	if expected := "Hello!"; string(b) != expected {
		t.Fatalf("wrong partially read part. expeected %q got %q", expected, string(b))
	}
}

func Test_unrecognizedStreamByte(t *testing.T) {
	r := dlog.NewReader(bytes.NewReader([]byte{0x03, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}))
	_, err := ioutil.ReadAll(r)
	if err == nil {
		t.Fatal("expected error")
	}
	if expected := "dlog: unexpected stream byte: 0x3"; err.Error() != expected {
		t.Fatalf("expected error: %q got:%v", expected, err)
	}
}

func Test_growsInitialBuffer(t *testing.T) {
	m := bytes.Repeat([]byte{'A'}, 3000) // intiial buf 2048
	r := dlog.NewReader(bytes.NewReader(msg(1, m)))
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !same(m, b) {
		t.Fatalf("wrong output: expected=(%d)%#v got=(%d)%#v", len(m), m, len(b), b)
	}
}

func Test_messageLimit_atLimit(t *testing.T) {
	b := append(msg(1, []byte("Hello!\n")),
		msg(1, bytes.Repeat([]byte{'A'}, 65536))...) // large msg but allowed 64k
	_, err := ioutil.ReadAll(dlog.NewReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_messageLimit_exceeds(t *testing.T) {
	b := append(msg(1, []byte("hello\n")),
		[]byte{0x1, 0x0, 0x0, 0x0 /* 65537 bytes */, 0x00, 0x01, 0x00, 0x01 /* rest must be ignored */, 0xff}...)
	_, err := ioutil.ReadAll(dlog.NewReader(bytes.NewReader(b)))
	if err == nil {
		t.Fatal("expected error")
	}
	if expected := "parsed msg too large"; !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q got: %v", expected, err)
	}
}

func Test_corruptMessage_missingBody(t *testing.T) {
	b := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05}
	out, err := ioutil.ReadAll(dlog.NewReader(bytes.NewReader(b)))
	if err == nil {
		t.Fatal("expected error", out)
	}
	if expected := "dlog: corrupt message read 0 out of 5 bytes: EOF"; err.Error() != expected {
		t.Fatalf("expected %q got: %v", expected, err)
	}
}

func Test_corruptMessage_partialBody(t *testing.T) {
	b := msg(1, []byte("helloworld"))
	b = b[:len(b)-1] // cut off last byte
	out, err := ioutil.ReadAll(dlog.NewReader(bytes.NewReader(b)))
	if err == nil {
		t.Fatal("expected error", out)
	}
	if expected := "dlog: corrupt message read 9 out of 10 bytes: unexpected EOF"; err.Error() != expected {
		t.Fatalf("expected %q got: %v", expected, err)
	}
}

func Test_messageReadFailure(t *testing.T) {
	r := io.MultiReader(
		bytes.NewReader([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05}), // prefix for msg len=5
		&badReader{})
	out, err := ioutil.ReadAll(dlog.NewReader(r))
	if err == nil {
		t.Fatal("expected error", out)
	}
	if expected := "dlog: failed to read message: phony error"; err.Error() != expected {
		t.Fatalf("expected %q got: %v", expected, err)
	}
}

func Test_twoSmallMessages_parsedCorrectly(t *testing.T) {
	b := append(msg(1, []byte("hello\n")), msg(2, []byte("world\n"))...)
	out, err := ioutil.ReadAll(dlog.NewReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}
	if expected := "hello\nworld\n"; string(out) != expected {
		t.Fatalf("wrong output: %q, expected: %q", string(out), expected)
	}
}

func msg(fd int8, b []byte) []byte {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(b)))
	v := []byte{byte(fd), 0x0, 0x0, 0x0}
	return append(append(v, size...), b...)
}

type badReader struct{}

func (_ *badReader) Read(p []byte) (int, error) {
	return 0, errors.New("phony error")
}

func same(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
