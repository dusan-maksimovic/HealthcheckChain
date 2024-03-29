package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/monitored module sentinel errors
var (
	ErrSample                       = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout         = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion               = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrInvalidChannelFlow           = sdkerrors.Register(ModuleName, 1502, "invalid message sent to channel end")
	ErrHealthcheckChannelAlreadySet = sdkerrors.Register(ModuleName, 1503, "channel for sending healthcheck updates is already set")
	ErrUnexpectedChannelID          = sdkerrors.Register(ModuleName, 1504, "unexpected channel ID")
)
