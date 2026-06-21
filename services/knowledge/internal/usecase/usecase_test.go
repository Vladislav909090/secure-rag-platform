package usecase

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestSizeLimitReaderAllowsExactLimit(t *testing.T) {
	reader := &sizeLimitReader{r: strings.NewReader("abc"), max: 3}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(data) != "abc" {
		t.Fatalf("ReadAll() = %q, want abc", data)
	}
	if reader.read != 3 || reader.failed {
		t.Fatalf("unexpected reader state: read=%d failed=%v", reader.read, reader.failed)
	}
}

func TestSizeLimitReaderRejectsOverflowAndStaysFailed(t *testing.T) {
	reader := &sizeLimitReader{r: strings.NewReader("abcd"), max: 3}
	buf := make([]byte, 4)

	n, err := reader.Read(buf)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("Read() error = %v, want ErrFileTooLarge", err)
	}
	if n != 4 {
		t.Fatalf("Read() n = %d, want 4", n)
	}
	if !reader.failed {
		t.Fatalf("reader should be failed after overflow")
	}

	n, err = reader.Read(buf)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("second Read() error = %v, want ErrFileTooLarge", err)
	}
	if n != 0 {
		t.Fatalf("second Read() n = %d, want 0", n)
	}
}
