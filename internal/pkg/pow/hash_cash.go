package pow

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

const zeroByte = 48 // ascii code for '0'

type HashCashData struct {
	Version    int    `json:"version"`
	ZerosCount int    `json:"zerosCount"`
	Date       int64  `json:"s_date"`
	Resource   string `json:"resource"`
	Rand       string `json:"rand"`
	Counter    int    `json:"counter"`
}

// sha1Hash - calculates sha1 hash from given string
func sha1Hash(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// toString - prepare data for hashing
func (h HashCashData) toString() string {
	return fmt.Sprintf("%d:%d:%d:%s::%s:%d", h.Version, h.ZerosCount, h.Date, h.Resource, h.Rand, h.Counter)
}

func Verify(hash string, zerosCount int) bool {
	if zerosCount > len(hash) {
		return false
	}
	for _, ch := range hash[:zerosCount] {
		if ch != zeroByte {
			return false
		}
	}
	return true
}

func (h HashCashData) Compute(maxIterations int) (string, error) {
	for h.Counter <= maxIterations || maxIterations <= 0 {
		header := h.toString()
		hash := sha1Hash(header)
		if Verify(hash, h.ZerosCount) {
			byteData, err := json.Marshal(h)
			if err != nil {
				return "", fmt.Errorf("err marshal hashcash: %w", err)
			}
			return string(byteData), nil
		}
		h.Counter++
	}
	return "", fmt.Errorf("max iterations exceeded")
}
