package cli_test

import (
	"encoding/json"
	"testing"

	milestonestypes "github.com/0xPolygon/heimdall-v2/x/milestone/types"
	staketypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"
)

func TestSetGenesisValidator(t *testing.T) {
	// Generate a public key for testing
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	// Call the function to test
	rawMsg, err := cli.SetGenesisValidator(pubKey)
	require.NoError(t, err)
	require.NotNil(t, rawMsg)

	// Verify the returned data is valid RawMessage
	require.IsType(t, json.RawMessage{}, rawMsg)

	// Unmarshal the JSON result
	var genesisState staketypes.GenesisState
	err = json.Unmarshal(rawMsg, &genesisState)
	require.NoError(t, err)

	// Validate the current validator set
	require.Len(t, genesisState.CurrentValidatorSet.Validators, 1)
	currentValidator := genesisState.CurrentValidatorSet.Validators[0]

	// Validate the validators' list
	require.Len(t, genesisState.Validators, 1)
	validator := genesisState.Validators[0]

	// Test current validator set fields
	require.Equal(t, uint64(1), currentValidator.ValId)
	require.Equal(t, int64(1000), currentValidator.VotingPower)
	require.Equal(t, pubKey.Bytes(), currentValidator.PubKey)
	require.Equal(t, pubKey.Address().String(), currentValidator.Signer)

	// Test validators list fields
	require.Equal(t, uint64(1), validator.ValId)
	require.Equal(t, uint64(0), validator.StartEpoch)
	require.Equal(t, uint64(1000000), validator.EndEpoch)
	require.Equal(t, int64(1000), validator.VotingPower)
	require.Equal(t, pubKey.Bytes(), validator.PubKey)
	require.Equal(t, pubKey.Address().String(), validator.Signer)
}

// TestSetGenesisValidatorJSON tests the JSON structure of the stake output
func TestSetGenesisValidatorJSON(t *testing.T) {
	// Generate a public key with known output for testing
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	// Get the expected address string
	expectedSigner := pubKey.Address().String()

	// Call the function to test
	rawMsg, err := cli.SetGenesisValidator(pubKey)
	require.NoError(t, err)

	// Convert to map to check JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(rawMsg, &result)
	require.NoError(t, err)

	// Check top-level structure
	currentValidatorSet, ok := result["current_validator_set"].(map[string]interface{})
	require.True(t, ok, "current_validator_set should be a JSON object")

	validators, ok := result["validators"].([]interface{})
	require.True(t, ok, "validators should be a JSON array")
	require.Len(t, validators, 1)

	// Check current validator set structure
	currentValidators, ok := currentValidatorSet["validators"].([]interface{})
	require.True(t, ok, "current_validator_set.validators should be a JSON array")
	require.Len(t, currentValidators, 1)

	// Get the current validator
	currentValidator, ok := currentValidators[0].(map[string]interface{})
	require.True(t, ok)

	// Check only the fields that exist in the JSON
	valId, ok := currentValidator["val_id"].(float64)
	require.True(t, ok, "val_id should be a number")
	require.Equal(t, float64(1), valId)

	votingPower, ok := currentValidator["voting_power"].(float64)
	require.True(t, ok, "voting_power should be a number")
	require.Equal(t, float64(1000), votingPower)

	_, ok = currentValidator["pub_key"]
	require.True(t, ok, "pub_key should exist")

	signer, ok := currentValidator["signer"].(string)
	require.True(t, ok, "signer should be a string")
	require.Equal(t, expectedSigner, signer)

	// Get and check the validator from validators array
	validator, ok := validators[0].(map[string]interface{})
	require.True(t, ok, "validator should be a JSON object")

	// Check fields that exist in this validator object
	valId, ok = validator["val_id"].(float64)
	require.True(t, ok, "val_id should be a number")
	require.Equal(t, float64(1), valId)

	endEpoch, ok := validator["end_epoch"].(float64)
	require.True(t, ok, "end_epoch should be a number")
	require.Equal(t, float64(1000000), endEpoch)

	votingPower, ok = validator["voting_power"].(float64)
	require.True(t, ok, "voting_power should be a number")
	require.Equal(t, float64(1000), votingPower)

	_, ok = validator["pub_key"]
	require.True(t, ok, "pub_key should exist")

	signer, ok = validator["signer"].(string)
	require.True(t, ok, "signer should be a string")
	require.Equal(t, expectedSigner, signer)
}

// TestSetGenesisValidator_ErrorHandling tests error handling in SetGenesisValidator
func TestSetGenesisValidator_ErrorHandling(t *testing.T) {
	// Passing a nil public key to check if error handling works
	rawMsg, err := cli.SetGenesisValidator(nil)
	require.Error(t, err)
	require.Nil(t, rawMsg)
	// Validate the error message
	require.Equal(t, "invalid public key: nil", err.Error())
}

// TestBuildEmptyMilestoneGenesis tests the JSON structure of the milestones output
func TestBuildEmptyMilestoneGenesis(t *testing.T) {
	rawMsg, err := cli.BuildEmptyMilestoneGenesis()
	require.NoError(t, err)

	// Convert to map to check JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(rawMsg, &result)
	require.NoError(t, err)

	params := milestonestypes.DefaultParams()

	// Check top-level structure
	milestonesArray, ok := result["milestones"].([]interface{})
	require.True(t, ok, "milestones should be a JSON array")
	milestonesParams, ok := result["params"].(map[string]interface{})
	require.True(t, ok, "params should be a JSON object")

	require.Len(t, milestonesArray, 0)
	require.NoError(t, err)
	require.Equal(t, float64(params.MaxMilestonePropositionLength), milestonesParams["max_milestone_proposition_length"])
	require.Equal(t, float64(params.FfMilestoneThreshold), milestonesParams["ff_milestone_threshold"])
	require.Equal(t, float64(params.FfMilestoneBlockInterval), milestonesParams["ff_milestone_block_interval"])
}
