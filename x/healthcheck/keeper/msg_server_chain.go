package keeper

import (
	"context"

	"healthcheck/x/healthcheck/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateChain(goCtx context.Context, msg *types.MsgCreateChain) (*types.MsgCreateChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetChain(
		ctx,
		msg.ChainId,
	)
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "index already set")
	}

	var chain = types.Chain{
		Creator:      msg.Creator,
		ChainId:      msg.ChainId,
		ConnectionId: msg.ConnectionId,
	}

	k.SetChain(
		ctx,
		chain,
	)
	return &types.MsgCreateChainResponse{}, nil
}

func (k msgServer) UpdateChain(goCtx context.Context, msg *types.MsgUpdateChain) (*types.MsgUpdateChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetChain(
		ctx,
		msg.ChainId,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	var chain = types.Chain{
		Creator:      msg.Creator,
		ChainId:      msg.ChainId,
		ConnectionId: msg.ConnectionId,
	}

	k.SetChain(ctx, chain)

	return &types.MsgUpdateChainResponse{}, nil
}

func (k msgServer) DeleteChain(goCtx context.Context, msg *types.MsgDeleteChain) (*types.MsgDeleteChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetChain(
		ctx,
		msg.ChainId,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	k.RemoveChain(
		ctx,
		msg.ChainId,
	)

	return &types.MsgDeleteChainResponse{}, nil
}
