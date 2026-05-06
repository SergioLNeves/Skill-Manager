package github

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	tarballURL = "https://codeload.github.com/%s/tar.gz/%s"
	shaURL     = "https://api.github.com/repos/%s/commits/%s"
)

// Fetcher downloads GitHub repos via tarball and extracts skill directories.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a Fetcher with the default HTTP client.
func NewFetcher() *Fetcher {
	return &Fetcher{client: &http.Client{}}
}

// FetchResult is the outcome of a successful Fetch call.
type FetchResult struct {
	// ExtractDir is the root of the extracted repo content.
	ExtractDir string
	// SHA is the resolved commit hash.
	SHA string
}

// Fetch downloads the tarball for <repo>@<ref>, extracts it under <cacheDir>/<repo>/<sha>/,
// and returns the extraction directory and resolved SHA. If the directory already exists,
// it is treated as a cache hit and no download is performed.
func (f *Fetcher) Fetch(ctx context.Context, cacheDir, repo, ref string) (FetchResult, error) {
	sha, err := f.resolveRef(ctx, repo, ref)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: resolve ref %s@%s: %w", repo, ref, err)
	}

	extractDir := filepath.Join(cacheDir, repo, sha)
	if _, err := os.Stat(extractDir); err == nil {
		return FetchResult{ExtractDir: extractDir, SHA: sha}, nil // cache hit
	}

	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: mkdir %s: %w", extractDir, err)
	}

	if err := f.downloadAndExtract(ctx, repo, ref, extractDir); err != nil {
		_ = os.RemoveAll(extractDir) // clean up partial extraction
		return FetchResult{}, fmt.Errorf("fetcher: download %s@%s: %w", repo, ref, err)
	}

	return FetchResult{ExtractDir: extractDir, SHA: sha}, nil
}

// HashFile returns the SHA-256 hex digest of a file's content.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// resolveRef returns the commit SHA for <repo>@<ref> using the GitHub API.
// If ref is already a 40-char hex SHA, it is returned as-is.
func (f *Fetcher) resolveRef(ctx context.Context, repo, ref string) (string, error) {
	if len(ref) == 40 && isHex(ref) {
		return ref, nil
	}
	url := fmt.Sprintf(shaURL, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.sha")
	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s for %s", resp.Status, url)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(body))
	if sha == "" {
		return "", fmt.Errorf("empty SHA from GitHub API for %s@%s", repo, ref)
	}
	return sha, nil
}

func (f *Fetcher) downloadAndExtract(ctx context.Context, repo, ref, dst string) error {
	url := fmt.Sprintf(tarballURL, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}
	return extractTarGz(resp.Body, dst)
}

// extractTarGz extracts a .tar.gz stream, stripping the top-level directory.
func extractTarGz(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var stripPrefix string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}

		// The first entry is always the top-level dir (e.g. "owner-repo-sha/").
		if stripPrefix == "" {
			parts := strings.SplitN(filepath.ToSlash(hdr.Name), "/", 2)
			stripPrefix = parts[0] + "/"
		}

		rel := strings.TrimPrefix(filepath.ToSlash(hdr.Name), stripPrefix)
		if rel == "" || rel == "." {
			continue
		}

		target := filepath.Join(dst, filepath.FromSlash(rel))

		// Guard against path traversal
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

func isHex(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
