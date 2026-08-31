package tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 90 * time.Second}

func downloadGH() error {
	osName, arch, ext := ghAssetParts()
	if osName == "" {
		return fmt.Errorf("unsupported platform %s/%s", goos, goarch)
	}
	url, err := latestGHAsset(osName, arch, ext)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "gh-download-*"+ext)
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := httpGetFile(url, tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	destDir, err := userLocalBin()
	if err != nil {
		return err
	}
	binName := "gh"
	if goos == "windows" {
		binName = "gh.exe"
	}
	dest := filepath.Join(destDir, binName)
	if ext == ".zip" {
		if err := extractZipFile(tmpName, binName, dest); err != nil {
			return err
		}
	} else {
		if err := extractTarGzFile(tmpName, "bin/gh", dest); err != nil {
			return err
		}
	}
	_ = os.Chmod(dest, 0755)
	EnsureToolPaths()
	if !GHInstalled() {
		return fmt.Errorf("gh extracted to %s but is not on PATH", dest)
	}
	return nil
}

func ghAssetParts() (osName, arch, ext string) {
	switch goos {
	case "linux":
		osName, ext = "linux", ".tar.gz"
	case "darwin":
		osName, ext = "macOS", ".zip"
	case "windows":
		osName, ext = "windows", ".zip"
	default:
		return "", "", ""
	}
	switch goarch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return "", "", ""
	}
	return osName, arch, ext
}

func latestGHAsset(osName, arch, ext string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/cli/cli/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "TaxiCheck")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github releases: HTTP %d", resp.StatusCode)
	}
	var body struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", err
	}
	want := fmt.Sprintf("_%s_%s%s", osName, arch, ext)
	for _, a := range body.Assets {
		if strings.Contains(a.Name, want) && !strings.Contains(a.Name, ".sha256") {
			return a.URL, nil
		}
	}
	return "", fmt.Errorf("no GitHub CLI build for %s/%s", osName, arch)
}

func httpGetFile(url string, w io.Writer) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "TaxiCheck")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	_, err = io.Copy(w, io.LimitReader(resp.Body, 80<<20))
	return err
}

func extractTarGzFile(archive, inner, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.ToSlash(hdr.Name)
		if !strings.HasSuffix(name, inner) && filepath.Base(name) != filepath.Base(inner) {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, io.LimitReader(tr, 80<<20))
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	return fmt.Errorf("gh binary not found in archive")
}

func extractZipFile(archive, binName, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(f.Name) != binName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, io.LimitReader(rc, 80<<20))
		rc.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	return fmt.Errorf("gh binary not found in archive")
}
