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
	FeeCollector           FeeCollector
	// TODO HV2 import and enable the following
	// ChainKeeper 		   chainmanager.Keeper
	// ContractCaller 		   helper.IContractCaller
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.

// TODO HV2 double check this function and all the decorators
//
//	is this enough to reconcile with heimdall's auth/ante.go?
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil { // TODO HV2: what is the signing mode for heimdall? SignDoc?
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	// TODO HV2: heimdall is using this for the supply method `SendCoinsFromAccountToModule`.
	//  Upstream supply is merged with bank, so do we need this or BankKeeper is enough?
	if options.FeeCollector == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "fee collector has not been set")
	}

	anteDecorators := []sdk.AnteDecorator{
		NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewValidateBasicDecorator(),
		// NewTxTimeoutHeightDecorator(), // TODO HV2 this is not present in heimdall
		NewValidateMemoDecorator(options.AccountKeeper),
		// NewConsumeGasForTxSizeDecorator(options.AccountKeeper), // TODO HV2 this was removed in heimdall's auth/ante.go (original ancestor's method `newCtx.GasMeter().ConsumeGas`)
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker, options.FeeCollector), // TODO HV2 heavily changed
		NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		// NewValidateSigCountDecorator(options.AccountKeeper), // TODO HV2 this was removed in heimdall's auth/ante.go (original ancestor's method `ValidateSigCount`)
		NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
