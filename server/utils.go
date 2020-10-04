package server

import (
	"crypto/rand"
	"encoding/binary"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func getSecuritySeed() string {
	var x uint64
	binary.Read(rand.Reader, binary.BigEndian, &x)
	return strconv.FormatUint(x, 10)
}

func genRandomToken() string {
	return genRandomString(64)
}

func genShortId() string {
	return genRandomString(12)
}

func genRandomString(n int) string {
	var p = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = p[int(b[i])%len(p)]
	}
	return string(b)
}

func makePassword(pass string) string {
	newpass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(newpass)
}

func comparePassword(pass, hashpass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(pass))
}

func isWeakPass(pass string) bool {
	return len(pass) < 6
}

func customQuote(s string) string {
	return `'` + s + `'`
}

func parseDomain(name, root string) (prefix, shortId string, rebind bool) {
	//r.u3yszl9nidbsx8p9.example.com.
	index := strings.Index(name, "."+root)
	if index <= 0 {
		return
	}

	//prefix = r.u3yszl9nidbsx8p9
	prefix = name[:index]
	lastIdx := strings.LastIndex(prefix, ".")
	if lastIdx <= 0 {
		shortId = prefix
		prefix = ""
		return
	}

	if lastIdx != len(prefix) {
		//shortId = u3yszl9nidbsx8p9
		shortId = prefix[lastIdx+1:]
	}

	//prefix = r
	prefix = prefix[:lastIdx]
	rebind = prefix == "r" || strings.HasSuffix(prefix, ".r")
	return
}
