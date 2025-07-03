package step

import (
	"testing"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-steplib/bitrise-step-authenticate-wth-gcp/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSupportedVersionIsPreinstalled(t *testing.T) {
	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedOutput").Return("{\"Google Cloud SDK\": \"528.0.0\"}", nil).Once()

	commandFactory := new(mocks.Factory)
	commandFactory.On("Create", "gcloud", []string{"version", "--format", "json", "--quiet"}, mock.Anything).Return(cmd).Once()

	authenticator := NewStep(
		stepconf.NewInputParser(env.NewRepository()),
		commandFactory,
		export.NewExporter(commandFactory),
		log.NewLogger(),
	)
	err := authenticator.InstallDependencies()
	require.NoError(t, err)

	cmd.AssertExpectations(t)
	commandFactory.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	versionCmd := new(mocks.Command)
	versionCmd.On("RunAndReturnTrimmedOutput").Return("{\"Google Cloud SDK\": \"428.0.0\"}", nil).Once()

	updateCmd := new(mocks.Command)
	updateCmd.On("Run").Return(nil).Once()

	commandFactory := new(mocks.Factory)
	commandFactory.On("Create", "gcloud", []string{"version", "--format", "json", "--quiet"}, mock.Anything).Return(versionCmd).Once()
	commandFactory.On("Create", "gcloud", []string{"components", "update", "--version=528.0.0"}, mock.Anything).Return(updateCmd).Once()

	authenticator := NewStep(
		stepconf.NewInputParser(env.NewRepository()),
		commandFactory,
		export.NewExporter(commandFactory),
		log.NewLogger(),
	)
	err := authenticator.InstallDependencies()
	require.NoError(t, err)

	versionCmd.AssertExpectations(t)
	updateCmd.AssertExpectations(t)
	commandFactory.AssertExpectations(t)
}
