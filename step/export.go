package step

const (
	authTokenKey       = "GOOGLE_AUTH_TOKEN"
	credentialsPathKey = "GOOGLE_APPLICATION_CREDENTIALS"
)

func (s *Step) Export(result Result) error {
	s.logger.Println()
	s.logger.Infof("Exporting outputs:")

	values := map[string]string{
		authTokenKey:       result.AuthToken,
		credentialsPathKey: result.CredentialsPath,
	}

	for k, v := range values {
		if err := s.exporter.ExportOutput(k, v); err != nil {
			s.logger.Warnf("Failed to export: %s: %s", k, err)
		} else {
			s.logger.Donef("Exported %s", k)
		}
	}

	return nil
}
