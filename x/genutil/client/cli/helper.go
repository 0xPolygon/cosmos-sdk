package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	staketypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func SetGenesisValidator(valPubKey cryptotypes.PubKey) (json.RawMessage, error) {
	if valPubKey == nil {
		return nil, fmt.Errorf("invalid public key: nil")
	}

	genesisState := staketypes.GenesisState{
		CurrentValidatorSet: staketypes.ValidatorSet{
			Validators: []*staketypes.Validator{
				{
					EndEpoch:         0,
					ValId:            1,
					StartEpoch:       0,
					Nonce:            0,
					VotingPower:      1000,
					PubKey:           valPubKey.Bytes(),
					Signer:           strings.ToUpper(valPubKey.Address().String()),
					LastUpdated:      "",
					Jailed:           false,
					ProposerPriority: 0,
				},
			},
		},
		Validators: []*staketypes.Validator{
			{
				ValId:            1,
				StartEpoch:       0,
				EndEpoch:         1000000,
				Nonce:            0,
				VotingPower:      1000,
				PubKey:           valPubKey.Bytes(),
				Signer:           valPubKey.Address().String(),
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

	return data, nil
}
