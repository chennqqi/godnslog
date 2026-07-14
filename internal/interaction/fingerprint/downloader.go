package fingerprint

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const mmdbLicenseSignupURL = "https://www.maxmind.com/en/geolite2/signup"

// mmdbName is the file we extract from inside the MaxMind tar.gz archive.
const mmdbName = "GeoLite2-ASN.mmdb"

// downloadMMDBTarGz fetches the tar.gz bytes from url with a 5-minute timeout.
func downloadMMDBTarGz(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download mmdb: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download mmdb: unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// extractMMDB walks the tar.gz, finds the entry ending in mmdbName,
// and writes its contents to destPath.
func extractMMDB(tarGzBytes []byte, destPath string) error {
	gzr, err := gzip.NewReader(bytes.NewReader(tarGzBytes))
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != mmdbName {
			continue
		}
		// Ensure parent dir exists.
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("mkdir dest: %w", err)
		}
		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create dest: %w", err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("write dest: %w", err)
		}
		return out.Close()
	}
	return fmt.Errorf("mmdb entry %q not found in archive", mmdbName)
}

// DownloadAndExtractMMDB downloads a MaxMind tar.gz from url and extracts
// GeoLite2-ASN.mmdb to destPath.
func DownloadAndExtractMMDB(url, destPath string) error {
	data, err := downloadMMDBTarGz(url)
	if err != nil {
		return err
	}
	return extractMMDB(data, destPath)
}

// EnsureMMDB ensures an mmdb file exists at path. If the file already exists,
// it is a no-op. If the file is missing and licenseKey is non-empty, it
// downloads from MaxMind. If both are missing, it returns an error that
// mentions the free signup URL.
func EnsureMMDB(path, licenseKey string) error {
	if path == "" {
		return errors.New("mmdb path is empty")
	}
	if _, err := os.Stat(path); err == nil {
		return nil // file already present
	}
	if licenseKey == "" {
		return fmt.Errorf("mmdb not found at %q and no license key configured; sign up for a free key at %s", path, mmdbLicenseSignupURL)
	}
	url := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz", licenseKey)
	if err := DownloadAndExtractMMDB(url, path); err != nil {
		return fmt.Errorf("auto-download mmdb failed: %w (sign up at %s)", err, mmdbLicenseSignupURL)
	}
	return nil
}
