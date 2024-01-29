package ante

import (
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, can the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.Tx, params types.Params) (sdk.Coins, int64, error) {
	amount, ok := sdkmath.NewIntFromString(params.GetTxFees())
	if !ok {
		return nil, 0, errorsmod.Wrap(sdkerrors.ErrInvalidTxFees, "must provide correct txFees")
	}

	// TODO HV2: gas is retrieved from Params as currently done in heimdall. Can this be changed
	gas := params.GetMaxTxGas()
	feeCoins := sdk.Coins{sdk.Coin{Denom: types.FeeToken, Amount: amount}}

	// TODO HV2: removed as not present in heimdall
	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	//if ctx.IsCheckTx() {
	//
	//	minGasPrices := ctx.MinGasPrices()
	//	if !minGasPrices.IsZero() {
	//		requiredFees := make(sdk.Coins, len(minGasPrices))
	//
	//		// Determine the required fees by multiplying each required minimum gas
	//		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	//		glDec := sdkmath.LegacyNewDec(int64(gas))
	//		for i, gp := range minGasPrices {
	//			fee := gp.Amount.Mul(glDec)
	//			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	//		}
	//
	//		if !feeCoins.IsAnyGTE(requiredFees) {
	//			return nil, 0, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
	//		}
	//	}
	//}

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritize as expected.
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
