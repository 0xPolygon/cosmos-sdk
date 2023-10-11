package ante

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper          AccountKeeper
	BankKeeper             types.BankKeeper
	ExtensionOptionChecker ExtensionOptionChecker
	FeegrantKeeper         FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
	TxFeeChecker           TxFeeChecker
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
// TODO CHECK HEIMDALL-V2 reconcile with heimdall's auth/ante.go (see commented code below)
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewValidateBasicDecorator(),
		NewTxTimeoutHeightDecorator(),
		NewValidateMemoDecorator(options.AccountKeeper),
		NewConsumeGasForTxSizeDecorator(options.AccountKeeper), // TODO CHECK HEIMDALL-V2 this was removed in heimdall's auth/ante.go (original ancestor's method `newCtx.GasMeter().ConsumeGas`)
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		NewSetPubKeyDecorator(options.AccountKeeper),        // SetPubKeyDecorator must be called before all signature verification decorators
		NewValidateSigCountDecorator(options.AccountKeeper), // TODO CHECK HEIMDALL-V2 this was removed in heimdall's auth/ante.go (original ancestor's method `ValidateSigCount`)
		NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// TODO CHECK HEIMDALL-V2 reconcile with previous method (create decorators?). Impact on x/auth/ante/sigverify.go and x/auth/ante/validator_tx_fee.go
//func NewAnteHandler(
//	ak AccountKeeper,
//	chainKeeper chainmanager.Keeper,
//	feeCollector FeeCollector,
//	contractCaller helper.IContractCaller,
//	sigGasConsumer SignatureVerificationGasConsumer,
//) sdk.AnteHandler {
//	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
//		// get module address
//		if addr := feeCollector.GetModuleAddress(authTypes.FeeCollectorName); addr.Empty() {
//			return newCtx, sdk.ErrInternal(fmt.Sprintf("%s module account has not been set", authTypes.FeeCollectorName)).Result(), true
//		}
//
//		// all transactions must be of type auth.StdTx
//		stdTx, ok := tx.(authTypes.StdTx)
//		if !ok {
//			// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
//			// during runTx.
//			newCtx = SetGasMeter(simulate, ctx, 0)
//			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
//		}
//
//		//Check whether the chain has reached the hard fork length to execute milestone msgs
//		if ctx.BlockHeight() < helper.GetAalborgHardForkHeight() && (stdTx.Msg.Type() == checkpointTypes.EventTypeMilestone || stdTx.Msg.Type() == checkpointTypes.EventTypeMilestoneTimeout) {
//			newCtx = SetGasMeter(simulate, ctx, 0)
//			return newCtx, sdk.ErrTxDecode("error decoding transaction").Result(), true
//		}
//
//		// get account params
//		params := ak.GetParams(ctx)
//
//		// gas for tx
//		gasForTx := params.MaxTxGas // stdTx.Fee.Gas
//
//		amount, ok := sdk.NewIntFromString(params.TxFees)
//		if !ok {
//			return newCtx, sdk.ErrInternal("Invalid param tx fees").Result(), true
//		}
//
//		feeForTx := sdk.Coins{sdk.Coin{Denom: authTypes.FeeToken, Amount: amount}} // stdTx.Fee.Amount
//
//		// new gas meter
//		newCtx = SetGasMeter(simulate, ctx, gasForTx)
//
//		// AnteHandlers must have their own defer/recover in order for the BaseApp
//		// to know how much gas was used! This is because the GasMeter is created in
//		// the AnteHandler, but if it panics the context won't be set properly in
//		// runTx's recover call.
//		defer func() {
//			if r := recover(); r != nil {
//				switch rType := r.(type) {
//				case sdk.ErrorOutOfGas:
//					log := fmt.Sprintf(
//						"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
//						rType.Descriptor, gasForTx, newCtx.GasMeter().GasConsumed(),
//					)
//					res = sdk.ErrOutOfGas(log).Result()
//
//					res.GasWanted = gasForTx
//					res.GasUsed = newCtx.GasMeter().GasConsumed()
//					abort = true
//				default:
//					panic(r)
//				}
//			}
//		}()
//
//		// validate tx
//		if err := tx.ValidateBasic(); err != nil {
//			return newCtx, err.Result(), true
//		}
//
//		if res := ValidateMemo(stdTx, params); !res.IsOK() {
//			return newCtx, res, true
//		}
//
//		// stdSigs contains the sequence number, account number, and signatures.
//		// When simulating, this would just be a 0-length slice.
//		signerAddrs := stdTx.GetSigners()
//
//		if len(signerAddrs) == 0 {
//			return newCtx, sdk.ErrNoSignatures("no signers").Result(), true
//		}
//
//		if len(signerAddrs) > 1 {
//			return newCtx, sdk.ErrUnauthorized("wrong number of signers").Result(), true
//		}
//
//		isGenesis := ctx.BlockHeight() == 0
//
//		// fetch first signer, who's going to pay the fees
//		signerAcc, res := GetSignerAcc(newCtx, ak, types.AccAddressToHeimdallAddress(signerAddrs[0]))
//		if !res.IsOK() {
//			return newCtx, res, true
//		}
//
//		// deduct the fees
//		if !feeForTx.IsZero() {
//			res = DeductFees(feeCollector, newCtx, signerAcc, feeForTx)
//			if !res.IsOK() {
//				return newCtx, res, true
//			}
//
//			// reload the account as fees have been deducted
//			signerAcc = ak.GetAccount(newCtx, signerAcc.GetAddress())
//		}
//
//		// stdSigs contains the sequence number, account number, and signatures.
//		// When simulating, this would just be a 0-length slice.
//		stdSigs := stdTx.GetSignatures()
//
//		// check signature, return account with incremented nonce
//		signBytes := GetSignBytes(ctx, newCtx.ChainID(), stdTx, signerAcc, isGenesis)
//
//		signerAcc, res = processSig(newCtx, signerAcc, stdSigs[0], signBytes, simulate, params, sigGasConsumer)
//		if !res.IsOK() {
//			return newCtx, res, true
//		}
//
//		ak.SetAccount(newCtx, signerAcc)
//
//		return newCtx, sdk.Result{GasWanted: gasForTx}, false // continue...
//	}
//}
