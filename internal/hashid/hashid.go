package hashid

import (
	"crypto/sha256"
	"fmt"
)

// Stable returns a deterministic 8-byte hex ID derived from s.
func Stable(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum[:8])
}
