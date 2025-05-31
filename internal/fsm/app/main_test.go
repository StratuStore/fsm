package app

import (
	"github.com/StratuStore/fsm/internal/libs/config"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestValidateApp(t *testing.T) {
	err := fx.ValidateApp(CreateApp(&config.Config{}))
	require.NoError(t, err)
}
