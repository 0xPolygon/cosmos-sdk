package cli_test

import (
	"encoding/json"
	"testing"

	stakeType "github.com/0xPolygon/heimdall-v2/x/stake/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"
  "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

)

func TestSetGenesisValidator(t *testing.T) {

	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	rawMsg, err := cli.SetGenesisValidator(pubKey)
	require.NoError(t, err)
	require.NotNil(t, rawMsg)

	var genesisState stakeType.GenesisState
	err = json.Unmarshal(rawMsg, &genesisState)
	require.NoError(t, err)

	require.Len(t, genesisState.CurrentValidatorSet.Validators, 1)
	currentValidator := genesisState.CurrentValidatorSet.Validators[0]
	
	require.Len(t, genesisState.Validators, 1)
	validator := genesisState.Validators[0]

	// Test current validator set fields
	require.Equal(t, uint64(1), currentValidator.ValId)
	require.Equal(t, uint64(0), currentValidator.StartEpoch)
	require.Equal(t, uint64(0), currentValidator.EndEpoch)
	require.Equal(t, uint64(0), currentValidator.Nonce)
	require.Equal(t, int64(1000), currentValidator.VotingPower)
	require.Equal(t, pubKey.Bytes(), currentValidator.PubKey)
	require.Equal(t, "0x"+pubKey.Address().String(), currentValidator.Signer)
	require.False(t, currentValidator.Jailed)
	require.Equal(t, int64(0), currentValidator.ProposerPriority)

	// Test validators list fields
	require.Equal(t, uint64(1), validator.ValId)
	require.Equal(t, uint64(0), validator.StartEpoch)
	require.Equal(t, uint64(1000000), validator.EndEpoch)
	require.Equal(t, uint64(0), validator.Nonce)
	require.Equal(t, int64(1000), validator.VotingPower)
	require.Equal(t, pubKey.Bytes(), validator.PubKey)
	require.Equal(t, "0x"+pubKey.Address().String(), validator.Signer)
	require.False(t, validator.Jailed)
	require.Equal(t, int64(0), validator.ProposerPriority)

	// Test error case with invalid marshaling (this would require mocking json.Marshal)
	// This is an advanced test that would require dependency injection or mocking
}

// TestSetGenesisValidatorJSON tests the JSON structure of the output
func TestSetGenesisValidatorJSON(t *testing.T) {

	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	
	expectedSigner := "0x" + pubKey.Address().String()

	rawMsg, err := cli.SetGenesisValidator(pubKey)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(rawMsg), &result)
	require.NoError(t, err)

	currentValidatorSet, ok := result["current_validator_set"].(map[string]interface{})
	require.True(t, ok, "current_validator_set should be a JSON object")
	
	validators, ok := result["validators"].([]interface{})
	require.True(t, ok, "validators should be a JSON array")
	require.Len(t, validators, 1)

	// Check current validator set structure
	currentValidators, ok := currentValidatorSet["validators"].([]interface{})
	require.True(t, ok, "current_validator_set.validators should be a JSON array")
	require.Len(t, currentValidators, 1)

	// Get and check the current validator
	currentValidator := currentValidators[0].(map[string]interface{})
	require.Equal(t, float64(1), currentValidator["val_id"])
	require.Equal(t, float64(0), currentValidator["start_epoch"])
	require.Equal(t, float64(0), currentValidator["end_epoch"])
	require.Equal(t, float64(1000), currentValidator["voting_power"])
	require.Equal(t, expectedSigner, currentValidator["signer"])

	// Get and check the validator from validators array
	validator := validators[0].(map[string]interface{})
	require.Equal(t, float64(1), validator["val_id"])
	require.Equal(t, float64(0), validator["start_epoch"])
	require.Equal(t, float64(1000000), validator["end_epoch"])
	require.Equal(t, float64(1000), validator["voting_power"])
	require.Equal(t, expectedSigner, validator["signer"])

}


