package step

import (
	"strings"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
)

type Input struct {
	ServiceAccountKey         string `env:"service_account_key,required"`
	DockerLogin               bool   `env:"docker_login,required,opt[true,false]"`
	ArtifactRegistryLocations string `env:"artifact_registry_locations,required"`
	Verbose                   bool   `env:"verbose,opt[true,false]"`
}

type Config struct {
	ServiceAccountKey         string
	DockerLogin               bool
	ArtifactRegistryLocations []string
}

type Result struct {
	AuthToken       string
	CredentialsPath string
}

type Step struct {
	inputParser    stepconf.InputParser
	commandFactory command.Factory
	exporter       export.Exporter
	logger         log.Logger
}

func NewStep(
	inputParser stepconf.InputParser,
	commandFactory command.Factory,
	exporter export.Exporter,
	logger log.Logger,
) Step {
	return Step{
		inputParser:    inputParser,
		commandFactory: commandFactory,
		exporter:       exporter,
		logger:         logger,
	}
}

func (s *Step) ProcessConfig() (*Config, error) {
	var input Input
	err := s.inputParser.Parse(&input)
	if err != nil {
		return &Config{}, err
	}

	stepconf.Print(input)
	s.logger.EnableDebugLog(input.Verbose)

	var locations []string
	for _, location := range strings.Split(input.ArtifactRegistryLocations, "\n") {
		if location == "" {
			continue
		}
		locations = append(locations, location)
	}

	return &Config{
		ServiceAccountKey:         input.ServiceAccountKey,
		DockerLogin:               input.DockerLogin,
		ArtifactRegistryLocations: locations,
	}, nil
}
