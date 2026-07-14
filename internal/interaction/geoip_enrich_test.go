package interaction

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/interaction/fingerprint"
)

func TestEnhanceInteraction_PersistsASN(t *testing.T) {
	fp := fingerprint.NewFingerprinter("")
	domain := "token.example.com"
	ip := "1.2.3.4"
	inter := &models.Interaction{
		ID:        "test-id",
		Type:      "dns",
		Timestamp: time.Now(),
		SourceIP:  ip,
		Domain:    &domain,
	}
	enhanceInteraction(inter, fp)
	if inter.ASN != nil {
		t.Errorf("expected nil ASN with empty fingerprinter, got %v", *inter.ASN)
	}
}
