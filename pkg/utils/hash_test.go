package utils

import (
	"os"
	"testing"
)

func TestGenerateFileHash_SameContent(t *testing.T) {
	// Two files with identical content should produce the same hash
	content := []byte("election result data")

	f1, _ := os.CreateTemp("", "hash_test_1")
	defer os.Remove(f1.Name())
	f1.Write(content)
	f1.Close()

	f2, _ := os.CreateTemp("", "hash_test_2")
	defer os.Remove(f2.Name())
	f2.Write(content)
	f2.Close()

	h1, err := GenerateFileHash(f1.Name())
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	h2, err := GenerateFileHash(f2.Name())
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected same hash for identical content: %s != %s", h1, h2)
	}
}

func TestGenerateFileHash_DifferentContent(t *testing.T) {
	f1, _ := os.CreateTemp("", "hash_test_diff_1")
	defer os.Remove(f1.Name())
	f1.Write([]byte("content A"))
	f1.Close()

	f2, _ := os.CreateTemp("", "hash_test_diff_2")
	defer os.Remove(f2.Name())
	f2.Write([]byte("content B"))
	f2.Close()

	h1, _ := GenerateFileHash(f1.Name())
	h2, _ := GenerateFileHash(f2.Name())

	if h1 == h2 {
		t.Error("expected different hashes for different content")
	}
}

func TestGenerateFileHash_MissingFile(t *testing.T) {
	_, err := GenerateFileHash("/non/existent/file.jpg")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
