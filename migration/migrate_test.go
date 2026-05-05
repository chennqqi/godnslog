package migration

import (
	"testing"
	"time"

	newmodels "github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
)

func TestFromTblDns(t *testing.T) {
	testTime := time.Now()

	dns := &oldmodels.TblDns{
		Id:     1,
		Uid:    100,
		Domain: "test.example.com",
		Var:    "abc123",
		Ip:     "192.168.1.1",
		Ctime:  testTime,
		Atime:  testTime,
	}

	interaction := newmodels.FromTblDns(dns)

	if interaction.Type != newmodels.InteractionTypeDNS {
		t.Errorf("Expected type %s, got %s", newmodels.InteractionTypeDNS, interaction.Type)
	}

	if interaction.Domain == nil || *interaction.Domain != dns.Domain {
		t.Errorf("Expected domain %s, got %v", dns.Domain, interaction.Domain)
	}

	if interaction.Token == nil || *interaction.Token != dns.Var {
		t.Errorf("Expected token %s, got %v", dns.Var, interaction.Token)
	}

	if interaction.SourceIP != dns.Ip {
		t.Errorf("Expected source IP %s, got %s", dns.Ip, interaction.SourceIP)
	}
}

func TestFromTblHttp(t *testing.T) {
	testTime := time.Now()

	http := &oldmodels.TblHttp{
		Id:     1,
		Uid:    100,
		Ip:     "192.168.1.1",
		Var:    "abc123",
		Path:   "/test",
		Method: "GET",
		Data:   "test data",
		Ctype:  "application/json",
		Ua:     "Mozilla/5.0",
		Ctime:  testTime,
		Atime:  testTime,
	}

	interaction := newmodels.FromTblHttp(http)

	if interaction.Type != newmodels.InteractionTypeHTTP {
		t.Errorf("Expected type %s, got %s", newmodels.InteractionTypeHTTP, interaction.Type)
	}

	if interaction.Method == nil || *interaction.Method != http.Method {
		t.Errorf("Expected method %s, got %v", http.Method, interaction.Method)
	}

	if interaction.Path == nil || *interaction.Path != http.Path {
		t.Errorf("Expected path %s, got %v", http.Path, interaction.Path)
	}

	if interaction.SourceIP != http.Ip {
		t.Errorf("Expected source IP %s, got %s", http.Ip, interaction.SourceIP)
	}
}
