package healthcheck

import (
	"fmt"

	"healthcheck/x/healthcheck/keeper"
	"healthcheck/x/healthcheck/types"
	commontypes "healthcheck/x/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
)

type IBCModule struct {
	keeper keeper.Keeper
}

func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return "", sdkerrors.Wrapf(types.ErrInvalidChannelFlow, "channel handshake must be initiated by monitored chain")
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	if order != channeltypes.ORDERED {
		return "", sdkerrors.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.ORDERED, order)
	}

	// Require portID is the portID module is bound to
	boundPort := im.keeper.GetPort(ctx)
	if boundPort != portID {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	if counterparty.PortId != commontypes.MonitoredPortID {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid counterparty port: %s, expected %s", counterparty.PortId, commontypes.MonitoredPortID)
	}

	metadata := &commontypes.HandshakeMetadata{}
	if err := metadata.Unmarshal([]byte(counterpartyVersion)); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidHandshakeMetadata,
			"error unmarshalling ibc-try metadata: \n%v; \nmetadata: %v", err, counterpartyVersion)
	}

	if metadata.Version != commontypes.Version {
		return "", sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, commontypes.Version)
	}

	// Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// If module can already authenticate the capability then module already owns it so we don't need to claim
	// Otherwise, module does not have channel capability and we must claim it from IBC
	if !im.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		// Only claim channel capability passed back by IBC module if we do not already own it
		if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return "", err
		}
	}

	if len(connectionHops) != 1 {
		return "", sdkerrors.Wrapf(types.ErrInvalidConnectionHops, "expected only one connection hop.")
	}

	monitoredChainID, err := im.keeper.GetCounterpartyChainIDFromConnection(ctx, connectionHops[0])
	if err != nil {
		return "", err
	}

	monitoredChain, found := im.keeper.GetChain(ctx, monitoredChainID)
	if !found {
		return "", sdkerrors.Wrapf(types.ErrChainNotRegistered, "chain with the chain ID %s isn't registered yet", monitoredChainID)
	}

	if monitoredChain.ConnectionId != connectionHops[0] {
		return "", sdkerrors.Wrapf(types.ErrUnexpectedConnectionID, "unexpected connection for chain with chain ID %s, expected: %s, got: %s", monitoredChainID, monitoredChain.ConnectionId, connectionHops[0])
	}

	// TODO: if channel is closed we could check the timeout and allow a new channel to be opened even if this two fields were set
	if monitoredChain.UpdateInterval != 0 && monitoredChain.TimeoutInterval != 0 {
		return "", types.ErrChainAlreadyTracked
	}

	if metadata.UpdateInterval == 0 {
		metadata.UpdateInterval = types.DefaultUpdateInterval
	}

	if metadata.TimeoutInterval == 0 {
		metadata.TimeoutInterval = types.DefaultTimeoutInterval
	}

	monitoredChain.TimeoutInterval = metadata.TimeoutInterval
	monitoredChain.UpdateInterval = metadata.UpdateInterval

	im.keeper.SetChain(ctx, monitoredChain)

	return commontypes.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_,
	counterpartyVersion string,
) error {
	return sdkerrors.Wrapf(types.ErrInvalidChannelFlow, "channel handshake must be initiated by monitored chain")
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	monitoredChainID, err := im.keeper.GetCounterpartyChainIDFromChannel(ctx, portID, channelID)
	if err != nil {
		return err
	}

	monitoredChain, found := im.keeper.GetChain(ctx, monitoredChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrChainNotRegistered, "chain with the chain ID %s isn't registered yet", monitoredChainID)
	}

	monitoredChain.ChannelId = channelID
	im.keeper.SetChain(ctx, monitoredChain)

	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	// this line is used by starport scaffolding # oracle/packet/module/recv

	var modulePacketData commontypes.HealthcheckPacketData
	if err := types.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &modulePacketData); err != nil {
		return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error()))
	}

	// connection ID is checked during handshake, so whatever chain ID is returned here we know it matches the right connection ID
	chainID, err := im.keeper.GetCounterpartyChainIDFromChannel(ctx, modulePacket.DestinationPort, modulePacket.DestinationChannel)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	monitoredChain, found := im.keeper.GetChain(ctx, chainID)
	if !found {
		return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(types.ErrChainNotRegistered, "chain with the chain ID %s isn't registered yet", chainID))
	}

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {
	case *commontypes.HealthcheckPacketData_Data:
		if monitoredChain.Timestamp > packet.Data.Timestamp ||
			monitoredChain.Block > packet.Data.Block {
			err := fmt.Errorf("newer healthcheck update has already been submitted for chain with chain ID %s", monitoredChain.ChainId)
			return channeltypes.NewErrorAcknowledgement(err)
		}

		monitoredChain.Status = uint64(types.Active)
		monitoredChain.Timestamp = packet.Data.Timestamp
		monitoredChain.Block = packet.Data.Block
		monitoredChain.RegistryBlockHeight = uint64(ctx.BlockHeight())
		im.keeper.SetChain(ctx, monitoredChain)

	// this line is used by starport scaffolding # ibc/packet/module/recv
	default:
		err := fmt.Errorf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return fmt.Errorf("registry chain does not send packets; no acknowledgements are expected")
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return fmt.Errorf("registry chain does not send packets; no timeouts are expected")
}
