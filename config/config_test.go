package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadNotFound(t *testing.T) {
	if _, err := ReadFile("testdata/notfound.json"); !os.IsNotExist(err) {
		t.Errorf("expected opening config error, got %v", err)
	}
}

func TestLoadEmpty(t *testing.T) {
	want := "cannot load market configuration"
	if _, err := ReadFile("testdata/empty"); err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("expected reading config error, got %v", err)
	}
}

func TestReadFile(t *testing.T) {
	s, err := ReadFile("testdata/config.json")
	if err != nil {
		t.Errorf("expected cannot load market configuration, got %v", err)
	}

	if wantHTTPAddress := "localhost:8080"; s.HTTPAddress != wantHTTPAddress {
		t.Errorf("HTTPAddress = %q, wanted %q", s.HTTPAddress, wantHTTPAddress)
	}
	if wantHTTPCertFile := "cert.pem"; s.HTTPCertFile != wantHTTPCertFile {
		t.Errorf("HTTPCertFile = %q, wanted %q", s.HTTPCertFile, wantHTTPCertFile)
	}
	if wantHTTPKeyFile := "cert.key"; s.HTTPKeyFile != wantHTTPKeyFile {
		t.Errorf("HTTPKeyFile = %q, wanted %q", s.HTTPKeyFile, wantHTTPKeyFile)
	}
	if wantSQLDataSourceName := "market"; s.SQLDataSourceName != wantSQLDataSourceName {
		t.Errorf("SQLDataSourceName = %q, wanted %q", s.SQLDataSourceName, wantSQLDataSourceName)
	}
	if wantFileStorageHost := "localhost:9000"; s.FileStorageHost != wantFileStorageHost {
		t.Errorf("FileStorageHost = %q, wanted %q", s.FileStorageHost, wantFileStorageHost)
	}
	if wantFileStorageAPIKey := "key"; s.FileStorageAPIKey != wantFileStorageAPIKey {
		t.Errorf("FileStorageAPIKey = %q, wanted %q", s.FileStorageAPIKey, wantFileStorageAPIKey)
	}
	if wantFileStorageSecret := "secret"; s.FileStorageSecret != wantFileStorageSecret {
		t.Errorf("FileStorageSecret = %q, wanted %q", s.FileStorageSecret, wantFileStorageSecret)
	}
	if wantHTTPInspectionAddress := "localhost:7000"; s.HTTPInspectionAddress != wantHTTPInspectionAddress {
		t.Errorf("HTTPInspectionAddress = %q, wanted %q", s.HTTPInspectionAddress, wantHTTPInspectionAddress)
	}
}
