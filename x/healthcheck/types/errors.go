package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/healthcheck module sentinel errors
var (
	ErrSample                   = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout     = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion           = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrInvalidChannelFlow       = sdkerrors.Register(ModuleName, 1502, "invalid message sent to channel end")
	ErrInvalidHandshakeMetadata = sdkerrors.Register(ModuleName, 1503, "invalid monitored handshake metadata")
	ErrInvalidConnectionHops    = sdkerrors.Register(ModuleName, 1504, "invalid connection hops")
	ErrChainNotRegistered       = sdkerrors.Register(ModuleName, 1505, "chain is not registered")
	ErrUnexpectedConnectionID   = sdkerrors.Register(ModuleName, 1506, "unexpected connection ID")
	ErrChainAlreadyTracked      = sdkerrors.Register(ModuleName, 1507, "chain is already tracked through another channel")
)
