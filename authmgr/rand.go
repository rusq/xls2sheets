package authmgr

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randString generates a random string with length of `n`.
func randString(n int) string {
	var sb strings.Builder
	sb.Grow(n)
	for i := n - 1; i >= 0; i-- {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}
	return sb.String()
}
