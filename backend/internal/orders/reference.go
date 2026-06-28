package orders

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// GenerateReference returns a Paystack reference of the form "RUE-XXXXXXXX"
// where X is a random uppercase hex digit. ~4 billion combinations; collision
// probability is negligible at case-study scale and the DB UNIQUE constraint
// is the backstop.
func GenerateReference() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("RUE-%08X", binary.BigEndian.Uint32(b[:])), nil
}
