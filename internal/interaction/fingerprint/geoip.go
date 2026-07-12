package fingerprint

import (
	"net"
	"os"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

// GeoIP handles optional geographic IP lookup
type GeoIP struct {
	db   *maxminddb.Reader
	mu   sync.RWMutex
	path string
}

type asnRecord struct {
	ASN          uint   `maxminddb:"autonomous_system_number"`
	Organization string `maxminddb:"autonomous_system_organization"`
}

// NewGeoIP opens a GeoLite2-ASN.mmdb file and returns a GeoIP reader.
func NewGeoIP(mmdbPath string) (*GeoIP, error) {
	if _, err := os.Stat(mmdbPath); os.IsNotExist(err) {
		return nil, err
	}
	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		return nil, err
	}
	return &GeoIP{db: db, path: mmdbPath}, nil
}

// Lookup returns the ASN, organization, and country for the given IP.
// If the IP is not found or an error occurs, it returns zero values.
func (g *GeoIP) Lookup(ip net.IP) (uint, string, string) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var record asnRecord
	err := g.db.Lookup(ip, &record)
	if err != nil || record.ASN == 0 {
		return 0, "", ""
	}
	return record.ASN, record.Organization, ""
}

// Reload reloads the GeoIP database from disk.
func (g *GeoIP) Reload() error {
	db, err := maxminddb.Open(g.path)
	if err != nil {
		return err
	}
	g.mu.Lock()
	old := g.db
	g.db = db
	g.mu.Unlock()
	old.Close()
	return nil
}

// Close closes the GeoIP database.
func (g *GeoIP) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.db.Close()
}
