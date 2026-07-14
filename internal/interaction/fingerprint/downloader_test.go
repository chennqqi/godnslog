package fingerprint

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// buildTarGz builds a .tar.gz whose sole entry is dir/GeoLite2-ASN.mmdb
// containing fakeContent, matching MaxMind's release archive layout.
func buildTarGz(t *testing.T, dir, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	filePath := filepath.Join(dir, name)
	hdr := &tar.Header{
		Name: filePath, Mode: 0644, Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func TestDownloadAndExtractMMDB_Success(t *testing.T) {
	tarBytes := buildTarGz(t, "GeoLite2-ASN_20260101", "GeoLite2-ASN.mmdb", "fake-mmdb-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(tarBytes)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "GeoLite2-ASN.mmdb")
	if err := DownloadAndExtractMMDB(srv.URL, dest); err != nil {
		t.Fatalf("DownloadAndExtractMMDB: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "fake-mmdb-content" {
		t.Errorf("extracted content = %q, want %q", got, "fake-mmdb-content")
	}
}

func TestDownloadAndExtractMMDB_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "GeoLite2-ASN.mmdb")
	err := DownloadAndExtractMMDB(srv.URL, dest)
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
}

func TestEnsureMMDB_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GeoLite2-ASN.mmdb")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	if err := EnsureMMDB(path, "any-key"); err != nil {
		t.Errorf("EnsureMMDB with existing file should return nil, got %v", err)
	}
}

func TestEnsureMMDB_NoLicenseKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GeoLite2-ASN.mmdb")
	err := EnsureMMDB(path, "")
	if err == nil {
		t.Fatal("expected error when no license key and no file, got nil")
	}
}
