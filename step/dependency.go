package step

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"slices"

	"github.com/bitrise-io/go-utils/v2/command"
)

const (
	supportedCLIVersion = "528.0.0"
	baseURL             = "https://storage.googleapis.com/cloud-sdk-release/google-cloud-cli-%s.tar.gz"
)

func (s *Step) InstallDependencies() error {
	s.logger.Println()
	s.logger.Infof("Getting CLI version:\n")

	version, ok, err := s.installedCLIVersion()
	if err != nil {
		return s.installCLI()
	}

	if ok && version == supportedCLIVersion {
		s.logger.Printf("CLI version %s is already installed.\n", version)
		return nil
	}

	if !ok {
		s.logger.Printf("Installing CLI version: %s\n", version)
		return s.installCLI()
	}

	s.logger.Printf("Updating CLI version to: %s\n", supportedCLIVersion)
	return s.updateCLI()
}

func (s *Step) installedCLIVersion() (string, bool, error) {
	cmd := s.commandFactory.Create("gcloud", []string{"version", "--format", "json", "--quiet"}, nil)
	output, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return "", false, err
	}

	var versions map[string]string
	err = json.Unmarshal([]byte(output), &versions)
	if err != nil {
		return "", false, err
	}

	sdkVersion, ok := versions["Google Cloud SDK"]
	if !ok {
		return "", false, err
	}

	return sdkVersion, true, nil
}

func (s *Step) installCLI() error {
	platform, err := getPlatform()
	if err != nil {
		return fmt.Errorf("failed to get platform: %w", err)
	}

	architecture, err := getArchitecture()
	if err != nil {
		return fmt.Errorf("failed to get architecture: %w", err)
	}

	url := fmt.Sprintf(baseURL, fmt.Sprintf("%s-%s-%s", supportedCLIVersion, platform, architecture))
	tarPath, err := downloadSDK(url)
	if err != nil {
		return fmt.Errorf("failed to download SDK from %s: %w", url, err)
	}
	defer os.Remove(tarPath)

	sdkPath, err := s.extractSDK(tarPath)
	if err != nil {
		return fmt.Errorf("failed to extract SDK from %s: %w", sdkPath, err)
	}

	err = s.setupSDK(sdkPath)
	if err != nil {
		return fmt.Errorf("failed to setup SDK: %w", err)
	}

	return nil
}

func (s *Step) updateCLI() error {
	cmd := s.commandFactory.Create("gcloud", []string{"components", "update", fmt.Sprintf("--version=%s", supportedCLIVersion)}, nil)
	return cmd.Run()
}

func getPlatform() (string, error) {
	supportedPlatforms := []string{"darwin", "linux"}
	if slices.Contains(supportedPlatforms, runtime.GOOS) {
		return runtime.GOOS, nil
	}

	return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func getArchitecture() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64", nil
	case "arm64":
		return "arm", nil
	}

	return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
}

func downloadSDK(url string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	tarPath := path.Join(tmpDir, "google-cloud-sdk.tar.gz")
	out, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return tarPath, err
}

func (s *Step) extractSDK(tarPath string) (string, error) {
	extractedPath := path.Join(path.Dir(tarPath), "gougle-cloud-sdk")
	if err := os.MkdirAll(extractedPath, os.ModePerm); err != nil {
		return "", err
	}

	cmd := s.commandFactory.Create("tar", []string{"-xf", tarPath, "-C", extractedPath}, nil)
	if output, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		s.logger.Errorf("tar failure: %s", output)
		return "", err
	}

	return extractedPath, nil
}

func (s *Step) setupSDK(path string) error {
	cmd := s.commandFactory.Create("./install.sh", []string{"--path-update", "true", "--quiet"}, &command.Opts{
		Dir: path,
	})
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = s.commandFactory.Create("./bin/gcloud", []string{"init"}, &command.Opts{
		Dir: path,
	})
	return cmd.Run()
}
