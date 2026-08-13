package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type target struct {
	name        string
	mainPackage string
	goos        string
	goarch      string
}

var (
	versionFlag string
	osFlag      string
	archFlag    string
	outFlag     string
)

func main() {
	flag.StringVar(&versionFlag, "version", "", "release version, for example v2.0.2")
	flag.StringVar(&versionFlag, "v", "", "alias for -version")
	flag.StringVar(&osFlag, "os", "linux,windows", "comma or space separated target systems")
	flag.StringVar(&osFlag, "p", "linux,windows", "alias for -os")
	flag.StringVar(&archFlag, "arch", "amd64", "comma or space separated target architectures")
	flag.StringVar(&archFlag, "a", "amd64", "alias for -arch")
	flag.StringVar(&outFlag, "out", "dist", "output directory")
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	version, err := resolveVersion(versionFlag)
	if err != nil {
		return err
	}
	commit, err := commandOutput("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("read git commit: %w", err)
	}

	operatingSystems, err := parseValues(osFlag, map[string]bool{"linux": true, "windows": true, "darwin": true})
	if err != nil {
		return fmt.Errorf("target systems: %w", err)
	}
	architectures, err := parseValues(archFlag, map[string]bool{"386": true, "amd64": true})
	if err != nil {
		return fmt.Errorf("target architectures: %w", err)
	}

	outDir, err := filepath.Abs(outFlag)
	if err != nil {
		return err
	}
	if err = os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clean output directory: %w", err)
	}
	if err = os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Join(outDir, ".build"))

	buildDate := time.Now().UTC().Format(time.RFC3339)
	ldflags := fmt.Sprintf("-w -X main.AppVersion=%s -X main.BuildDate=%s -X main.GitCommit=%s", version, buildDate, commit)
	projects := []struct {
		name        string
		mainPackage string
	}{
		{name: "gocron", mainPackage: "./cmd/gocron"},
		{name: "gocron-node", mainPackage: "./cmd/node"},
	}

	archives := make([]string, 0, len(projects)*len(operatingSystems)*len(architectures))
	for _, project := range projects {
		for _, goos := range operatingSystems {
			for _, goarch := range architectures {
				t := target{name: project.name, mainPackage: project.mainPackage, goos: goos, goarch: goarch}
				archive, buildErr := buildTarget(outDir, version, ldflags, t)
				if buildErr != nil {
					return buildErr
				}
				archives = append(archives, archive)
			}
		}
	}

	if err = writeChecksums(outDir, archives); err != nil {
		return err
	}
	fmt.Printf("release packages written to %s\n", outDir)
	return nil
}

func resolveVersion(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		var err error
		value, err = commandOutput("git", "describe", "--tags", "--always")
		if err != nil {
			return "", fmt.Errorf("resolve version: %w", err)
		}
	}
	if !strings.HasPrefix(value, "v") {
		value = "v" + value
	}
	for _, r := range strings.TrimPrefix(value, "v") {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isNumber := r >= '0' && r <= '9'
		if !isLetter && !isNumber && r != '.' && r != '-' && r != '_' {
			return "", fmt.Errorf("invalid release version %q", value)
		}
	}
	return value, nil
}

func parseValues(value string, supported map[string]bool) ([]string, error) {
	items := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	if len(items) == 0 {
		return nil, fmt.Errorf("no values supplied")
	}
	for _, item := range items {
		if !supported[item] {
			return nil, fmt.Errorf("unsupported value %q", item)
		}
	}
	return items, nil
}

func buildTarget(outDir, version, ldflags string, t target) (string, error) {
	rootName := fmt.Sprintf("%s-%s-%s", t.name, t.goos, t.goarch)
	stagingDir := filepath.Join(outDir, ".build", rootName)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return "", err
	}

	binaryName := t.name
	if t.goos == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(stagingDir, binaryName)
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", binaryPath, t.mainPackage)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+t.goos, "GOARCH="+t.goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("building %s/%s %s\n", t.goos, t.goarch, t.name)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build %s: %w", rootName, err)
	}
	if t.goos != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return "", err
		}
	}
	for _, name := range []string{"README.md", "LICENSE"} {
		if err := copyFile(name, filepath.Join(stagingDir, name), 0644); err != nil {
			return "", err
		}
	}

	archiveBase := fmt.Sprintf("new-%s-%s-%s-%s", t.name, version, t.goos, t.goarch)
	var archivePath string
	if t.goos == "windows" {
		archivePath = filepath.Join(outDir, archiveBase+".zip")
		if err := writeZip(archivePath, stagingDir, rootName, binaryName); err != nil {
			return "", err
		}
	} else {
		archivePath = filepath.Join(outDir, archiveBase+".tar.gz")
		if err := writeTarGz(archivePath, stagingDir, rootName, binaryName); err != nil {
			return "", err
		}
	}
	return archivePath, nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func writeZip(destination, sourceDir, rootName, binaryName string) error {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(file)
	for _, name := range []string{binaryName, "README.md", "LICENSE"} {
		data, readErr := ioutil.ReadFile(filepath.Join(sourceDir, name))
		if readErr != nil {
			writer.Close()
			file.Close()
			return readErr
		}
		header := &zip.FileHeader{Name: filepath.ToSlash(filepath.Join(rootName, name)), Method: zip.Deflate}
		header.SetMode(0644)
		entry, createErr := writer.CreateHeader(header)
		if createErr != nil {
			writer.Close()
			file.Close()
			return createErr
		}
		if _, createErr = entry.Write(data); createErr != nil {
			writer.Close()
			file.Close()
			return createErr
		}
	}
	if err = writer.Close(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func writeTarGz(destination, sourceDir, rootName, binaryName string) error {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	entries := []struct {
		name string
		mode int64
	}{
		{name: binaryName, mode: 0755},
		{name: "README.md", mode: 0644},
		{name: "LICENSE", mode: 0644},
	}
	for _, entry := range entries {
		data, readErr := ioutil.ReadFile(filepath.Join(sourceDir, entry.name))
		if readErr != nil {
			return closeTar(file, gzipWriter, tarWriter, readErr)
		}
		header := &tar.Header{
			Name:    filepath.ToSlash(filepath.Join(rootName, entry.name)),
			Mode:    entry.mode,
			Size:    int64(len(data)),
			ModTime: time.Now().UTC(),
		}
		if err = tarWriter.WriteHeader(header); err != nil {
			return closeTar(file, gzipWriter, tarWriter, err)
		}
		if _, err = tarWriter.Write(data); err != nil {
			return closeTar(file, gzipWriter, tarWriter, err)
		}
	}
	return closeTar(file, gzipWriter, tarWriter, nil)
}

func closeTar(file *os.File, gzipWriter *gzip.Writer, tarWriter *tar.Writer, currentErr error) error {
	if err := tarWriter.Close(); currentErr == nil {
		currentErr = err
	}
	if err := gzipWriter.Close(); currentErr == nil {
		currentErr = err
	}
	if err := file.Close(); currentErr == nil {
		currentErr = err
	}
	return currentErr
}

func writeChecksums(outDir string, archives []string) error {
	sort.Strings(archives)
	lines := make([]string, 0, len(archives))
	for _, archive := range archives {
		file, err := os.Open(archive)
		if err != nil {
			return err
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		file.Close()
		if copyErr != nil {
			return copyErr
		}
		lines = append(lines, fmt.Sprintf("%s  %s", hex.EncodeToString(hash.Sum(nil)), filepath.Base(archive)))
	}
	return ioutil.WriteFile(filepath.Join(outDir, "SHA256SUMS.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func commandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0")
	}
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}
