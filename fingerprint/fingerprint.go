package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"
	"strings"
)

// HashBytes returns the SHA-256 hex digest of data.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// HashFile returns the SHA-256 hex digest of the file at path.
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

// MerkleRoot computes an order-independent fingerprint from a slice of
// fingerprint strings. It sorts them lexicographically, joins with a null
// separator, and returns the SHA-256 hex digest. An empty input returns the
// hash of the empty string.
func MerkleRoot(fingerprints []string) string {
	sorted := make([]string, len(fingerprints))
	copy(sorted, fingerprints)
	sort.Strings(sorted)
	combined := strings.Join(sorted, "\x00")
	return HashBytes([]byte(combined))
}
