package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateChain = "create_chain"
	TypeMsgUpdateChain = "update_chain"
	TypeMsgDeleteChain = "delete_chain"
)

var _ sdk.Msg = &MsgCreateChain{}

func NewMsgCreateChain(
	creator string,
	chainId string,
	connectionId string,

) *MsgCreateChain {
	return &MsgCreateChain{
		Creator:      creator,
		ChainId:      chainId,
		ConnectionId: connectionId,
	}
}

func (msg *MsgCreateChain) Route() string {
	return RouterKey
}

func (msg *MsgCreateChain) Type() string {
	return TypeMsgCreateChain
}

func (msg *MsgCreateChain) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateChain) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateChain) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateChain{}

func NewMsgUpdateChain(
	creator string,
	chainId string,
	connectionId string,

) *MsgUpdateChain {
	return &MsgUpdateChain{
		Creator:      creator,
		ChainId:      chainId,
		ConnectionId: connectionId,
	}
}

func (msg *MsgUpdateChain) Route() string {
	return RouterKey
}

func (msg *MsgUpdateChain) Type() string {
	return TypeMsgUpdateChain
}

func (msg *MsgUpdateChain) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateChain) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateChain) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

var _ sdk.Msg = &MsgDeleteChain{}

func NewMsgDeleteChain(
	creator string,
	chainId string,

) *MsgDeleteChain {
	return &MsgDeleteChain{
		Creator: creator,
		ChainId: chainId,
	}
}
func (msg *MsgDeleteChain) Route() string {
	return RouterKey
}

func (msg *MsgDeleteChain) Type() string {
	return TypeMsgDeleteChain
}

func (msg *MsgDeleteChain) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteChain) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteChain) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
