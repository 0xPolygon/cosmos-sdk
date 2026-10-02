package server

import (
	"bytes"
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"
)

func TestCometLoggerWrapperWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := CometLoggerWrapper{Logger: log.NewLogger(&buf, log.OutputJSONOption())}

	logger.With("module", "p2p").Warn("peer misbehaved", "peer", "abc")

	var line map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &line))
	require.Equal(t, "warn", line["level"])
	require.Equal(t, "peer misbehaved", line["message"])
	require.Equal(t, "p2p", line["module"])
	require.Equal(t, "abc", line["peer"])
}
