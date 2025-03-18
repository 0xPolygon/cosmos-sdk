package cli

import (
	"encoding/json"
	stakeType "github.com/0xPolygon/heimdall-v2/x/stake/types"
  cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

)

func SetGenesisValidator(valPubKey cryptotypes.PubKey) (json.RawMessage, error) {
	genesisState := stakeType.GenesisState{
		CurrentValidatorSet: stakeType.ValidatorSet{
			Validators: []*stakeType.Validator{
				{
					EndEpoch:         0,
					ValId:            1,
					StartEpoch:       0,
					Nonce:            0,
					VotingPower:      1000,
					PubKey:           valPubKey.Bytes(),
					Signer:           "0x" + valPubKey.Address().String(),
					LastUpdated:      "",
					Jailed:           false,
					ProposerPriority: 0,
				},
			},
		},
		Validators: []*stakeType.Validator{
			{
				ValId:            1,
				StartEpoch:       0,
				EndEpoch:         1000000,
				Nonce:            0,
				VotingPower:      1000,
				PubKey:           valPubKey.Bytes(),
				Signer:           "0x" + valPubKey.Address().String(),
				LastUpdated:      "",
				Jailed:           false,
				ProposerPriority: 0,
			},
		},
	}

	data, err := json.Marshal(genesisState)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

