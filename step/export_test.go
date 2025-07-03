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

func TestExport(t *testing.T) {
	result := Result{
		AuthToken:       "auth-token",
		CredentialsPath: "credentials-path",
	}

	cmd := new(mocks.Command)
	cmd.On("RunAndReturnTrimmedCombinedOutput").Return("", nil).Twice()

	commandFactory := new(mocks.Factory)
	commandFactory.On("Create", "envman", []string{"add", "--key", "GOOGLE_AUTH_TOKEN", "--value", result.AuthToken}, mock.Anything).Return(cmd)
	commandFactory.On("Create", "envman", []string{"add", "--key", "GOOGLE_APPLICATION_CREDENTIALS", "--value", result.CredentialsPath}, mock.Anything).Return(cmd)

	authenticator := NewStep(
		stepconf.NewInputParser(env.NewRepository()),
		commandFactory,
		export.NewExporter(commandFactory),
		log.NewLogger(),
	)
	err := authenticator.Export(result)
	require.NoError(t, err)

	commandFactory.AssertExpectations(t)
}
