package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

func ComputeSHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ComputeCRC32(data []byte) string {
	table := crc32.MakeTable(crc32.IEEE)
	checksum := crc32.Checksum(data, table)
	return hex.EncodeToString([]byte{byte(checksum >> 24), byte(checksum >> 16), byte(checksum >> 8), byte(checksum)})
}

