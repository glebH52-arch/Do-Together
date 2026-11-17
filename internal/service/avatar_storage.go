package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"uuid"
)

const MaxAvatarBytes int64 = 5 << 20

var (
	ErrAvatarEmpty      = errors.New("avatar is empty")
	ErrAvatarTooLarge   = errors.New("avatar is too large")
	ErrAvatarFormat     = errors.New("avatar format must be JPEG or PNG")
	ErrAvatarDimensions = errors.New("avatar dimensions are invalid")
)

type AvatarStorage struct {
	directory string
}

func NewAvatarStorage(directory string) (*AvatarStorage, error) {
	if directory == "" {
		return nil, errors.New("avatar directory is empty")
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return nil, fmt.Errorf("create avatar directory: %w", err)
	}
	return &AvatarStorage{directory: directory}, nil
}

func (s *AvatarStorage) Save(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxAvatarBytes+1))
	if err != nil {
		return "", fmt.Errorf("read avatar: %w", err)
	}
	if len(data) == 0 {
		return "", ErrAvatarEmpty
	}
	if int64(len(data)) > MaxAvatarBytes {
		return "", ErrAvatarTooLarge
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		return "", ErrAvatarFormat
	}
	if config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 || int64(config.Width)*int64(config.Height) > 16_000_000 {
		return "", ErrAvatarDimensions
	}
	extension := ".png"
	if format == "jpeg" {
		extension = ".jpg"
	}
	filename := uuid.New().String() + extension
	temporary, err := os.CreateTemp(s.directory, ".avatar-*")
	if err != nil {
		return "", fmt.Errorf("create avatar file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return "", fmt.Errorf("write avatar file: %w", err)
	}
	if err := temporary.Chmod(0o640); err != nil {
		temporary.Close()
		return "", fmt.Errorf("set avatar permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close avatar file: %w", err)
	}
	if err := os.Rename(temporaryName, filepath.Join(s.directory, filename)); err != nil {
		return "", fmt.Errorf("publish avatar file: %w", err)
	}
	return filename, nil
}

func (s *AvatarStorage) Delete(filename string) error {
	if filename == "" || filepath.Base(filename) != filename {
		return nil
	}
	err := os.Remove(filepath.Join(s.directory, filename))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete avatar: %w", err)
	}
	return nil
}

func (s *AvatarStorage) Path(filename string) (string, bool) {
	if filename == "" || filepath.Base(filename) != filename {
		return "", false
	}
	return filepath.Join(s.directory, filename), true
}
