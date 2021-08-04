package server

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	binaryIPExp = regexp.MustCompile(`(?:0b)?((?:0|1){25,32})`)
	hexIPExp    = regexp.MustCompile(`(?i:0x)?([0-f]?[0-f]{7})`)
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

func parsePrefix(prefix string) (pprefix, mode string) {
	index := strings.LastIndex(prefix, ".")
	if index <= 0 {
		pprefix = prefix
		return
	}
	mode = prefix[index+1:]
	pprefix = prefix[:index]
	return
}

func parseDomain(name, root string) (prefix, shortId string, rebind bool) {
	//r.u3yszl9nidbsx8p9.example.com.
	//abc.r.u3yszl9nidbsx8p9.example.com.
	//127.0.0.1-100.100.100.cr.u3yszl9nidbsx8p9.example.com.
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

func parseQuestionName(name, root string) (q string) {
	index := strings.Index(name, "."+root)
	if index <= 0 {
		return
	}

	q = name[:index]
	return
}

func parseBinaryIP(ip string) (net.IP, error) {
	uip, err := strconv.ParseUint(ip, 2, 64)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(fmt.Sprintf("%d.%d.%d.%d",
		(uip>>24)&0xFF, (uip>>16)&0xFF, (uip>>8)&0xFF, uip&0xFF)), nil
}

func parseHexIP(ip string) (net.IP, error) {
	uip, err := strconv.ParseUint(ip, 16, 32)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(fmt.Sprintf("%d.%d.%d.%d",
		(uip>>24)&0xFF, (uip>>16)&0xFF, (uip>>8)&0xFF, uip&0xFF)), nil
}

func parseIP(ip string) (net.IP, error) {
	{
		subs := hexIPExp.FindAllStringSubmatch(ip, 1)
		if len(subs) == 1 {
			return parseHexIP(subs[0][1])
		}
	}
	{
		subs := binaryIPExp.FindAllStringSubmatch(ip, 1)
		if len(subs) == 1 {
			return parseHexIP(subs[0][1])
		}
	}
	return net.ParseIP(ip), nil
}
