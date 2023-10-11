package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

// TODO CHECK HEIMDALL-V2 this is imported from heimdall > needs reimplementation?
// RegisterRoutes - Central function to define routes that get registered by the main application
// RegisterRoutes registers the auth module REST routes.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/auth/accounts/{address}", QueryAccountRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/auth/accounts/{address}/sequence", QueryAccountSequenceRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/auth/params", paramsHandlerFn(cliCtx)).Methods("GET")
}
