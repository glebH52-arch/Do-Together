package service

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"strings"
	"testing"
)

func TestAvatarStorageSaveAndDelete(t *testing.T) {
	storage, err := NewAvatarStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var source bytes.Buffer
	if err := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	filename, err := storage.Save(&source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(filename, ".png") {
		t.Fatalf("expected canonical png filename, got %q", filename)
	}
	path, ok := storage.Path(filename)
	if !ok {
		t.Fatal("saved filename was rejected")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if err := storage.Delete(filename); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected deleted file, got %v", err)
	}
}

func TestAvatarStorageRejectsContentAndPaths(t *testing.T) {
	storage, err := NewAvatarStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Save(strings.NewReader("not an image")); err != ErrAvatarFormat {
		t.Fatalf("expected format error, got %v", err)
	}
	if _, ok := storage.Path("../secret.png"); ok {
		t.Fatal("path traversal must be rejected")
	}
}
