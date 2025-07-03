package step

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

func (s *Step) Run(config Config) (Result, error) {
	s.logger.Println()
	s.logger.Infof("Performing GCP authentication:")

	email, err := extractEmail(config.ServiceAccountKey)
	if err != nil {
		return Result{}, fmt.Errorf("failed to extract email from service account key: %w", err)
	}

	keyPath, err := save(config.ServiceAccountKey)
	if err != nil {
		return Result{}, fmt.Errorf("failed to save service account key: %w", err)
	}

	if err := s.authenticate(email, keyPath); err != nil {
		return Result{}, fmt.Errorf("failed to authenticate with service account: %w", err)
	}

	s.logger.Printf("GCP authentication successful")

	token, err := s.generateToken()
	if err != nil {
		return Result{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	s.logger.Printf("Access token generated")

	if config.DockerLogin {
		err = s.loginWithDocker(token, config.ArtifactRegistryLocations)
		if err != nil {
			return Result{}, fmt.Errorf("failed to login to Docker with GCP token: %w", err)
		}

		s.logger.Printf("Logged in with Docker")
	}

	return Result{
		AuthToken:       token,
		CredentialsPath: keyPath,
	}, nil
}

func extractEmail(data string) (string, error) {
	var values map[string]string
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return "", err
	}

	email, ok := values["client_email"]
	if !ok {
		return "", fmt.Errorf("no client_email found")
	}

	return email, nil
}

func save(serviceAccountKey string) (string, error) {
	file, err := os.CreateTemp("", "service_account_key_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for service account key: %w", err)
	}

	if _, err := file.WriteString(serviceAccountKey); err != nil {
		return "", fmt.Errorf("failed to write service account key to temporary file: %w", err)
	}

	return file.Name(), nil
}

func (s *Step) authenticate(email, keyPath string) error {
	cmd := s.commandFactory.Create("gcloud", []string{"auth", "activate-service-account", email, fmt.Sprintf("--key-file=%s", keyPath)}, nil)
	if output, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		s.logger.Errorf("GCP authentication output: %s", output)
		return err
	}

	return nil
}

func (s *Step) generateToken() (string, error) {
	cmd := s.commandFactory.Create("gcloud", []string{"auth", "print-access-token", "--quiet"}, nil)
	output, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", err
	}

	return output, nil
}

func (s *Step) loginWithDocker(token string, locations []string) error {
	for _, location := range locations {
		cmd := s.commandFactory.Create("docker", []string{"login", "-u", "oauth2accesstoken", "--password-stdin", fmt.Sprintf("https://%s", location)}, &command.Opts{
			Stdin: strings.NewReader(token),
		})
		if output, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
			s.logger.Errorf("Docker login output: %s", output)
			return err
		}
	}

	return nil
}
