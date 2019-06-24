package randomer

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// RandString contains random texts and methods to retreive them
type RandString struct {
	mutex   sync.RWMutex
	strings []string
	len     int
}

// NewRandString returns RandString
func NewRandString() *RandString {
	rand.Seed(time.Now().UnixNano())
	return new(RandString)
}

// Add adds new string to pool
func (rs *RandString) Add(in string) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	rs.strings = append(rs.strings, in)
	rs.len++
}

// String returns random string
func (rs *RandString) String() string {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	if rs.len == 0 {
		return ""
	}
	rnd := rand.Intn(rs.len)
	return strings.ReplaceAll(rs.strings[rnd], "{rnd}", randomString(6))
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
