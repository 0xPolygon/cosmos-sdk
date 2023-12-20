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
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
// TODO CHECK HEIMDALL-V2 is this enough to reconcile with heimdall's auth/ante.go? (
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil { // TODO CHECK HEIMDALL-V2: what is the signing mode for heimdall? SignDoc?
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.FeeCollector == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "fee collector has not been set")
	}

	anteDecorators := []sdk.AnteDecorator{
		NewSetUpContextDecorator(),                                   // outermost AnteDecorator. SetUpContext must be called first // TODO CHECK HEIMDALL-V2 this should be ok
		NewExtensionOptionsDecorator(options.ExtensionOptionChecker), // TODO CHECK HEIMDALL-V2 this should be ok
		NewValidateBasicDecorator(),                                  // TODO CHECK HEIMDALL-V2 this should be ok
		// NewTxTimeoutHeightDecorator(),                                	// TODO CHECK HEIMDALL-V2 this is not present in heimdall
		NewValidateMemoDecorator(options.AccountKeeper), // TODO CHECK HEIMDALL-V2 this should be ok
		// NewConsumeGasForTxSizeDecorator(options.AccountKeeper),       	// TODO CHECK HEIMDALL-V2 this was removed in heimdall's auth/ante.go (original ancestor's method `newCtx.GasMeter().ConsumeGas`)
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker, options.FeeCollector), // TODO CHECK HEIMDALL-V2 heavily changed
		NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators // TODO CHECK HEIMDALL-V2 it should be ok (or we could remove the multiSig support)
		// NewValidateSigCountDecorator(options.AccountKeeper),				// TODO CHECK HEIMDALL-V2 this was removed in heimdall's auth/ante.go (original ancestor's method `ValidateSigCount`)
		NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),    // TODO CHECK HEIMDALL-V2 brand new, should be ok, to double check
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler), // TODO CHECK HEIMDALL-V2 brand new, should be ok, to double check
		NewIncrementSequenceDecorator(options.AccountKeeper),                        // TODO CHECK HEIMDALL-V2 this should be ok
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
