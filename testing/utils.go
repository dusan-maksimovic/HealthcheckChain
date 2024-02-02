package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	appmonitored "healthcheck/app/monitored"
	appregistry "healthcheck/app/registry"

	registrytypes "healthcheck/x/healthcheck/types"
	commontypes "healthcheck/x/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

func SetupRegistryTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	encCdc := appregistry.MakeEncodingConfig()
	app := appregistry.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		appregistry.DefaultNodeHome,
		5,
		encCdc,
		simapp.EmptyAppOptions{},
	)
	return app, appregistry.NewDefaultGenesisState(encCdc.Marshaler)
}

func SetupMonitoredTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	encCdc := appmonitored.MakeEncodingConfig()
	app := appmonitored.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		appmonitored.DefaultNodeHome,
		5,
		encCdc,
		simapp.EmptyAppOptions{},
	)
	return app, appmonitored.NewDefaultGenesisState(encCdc.Marshaler)
}

type HealthcheckTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	registryChain  *ibctesting.TestChain
	monitoredChain *ibctesting.TestChain

	registryApp  *appregistry.App
	monitoredApp *appmonitored.App

	packetSniffers map[*ibctesting.TestChain]*packetSniffer

	path *ibctesting.Path
}

func (suite *HealthcheckTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 0)

	ibctesting.DefaultTestingAppInit = SetupRegistryTestingApp
	suite.registryChain = ibctesting.NewTestChain(suite.T(), suite.coordinator, appregistry.Name)
	suite.coordinator.Chains[appregistry.Name] = suite.registryChain
	suite.registryApp = suite.registryChain.App.(*appregistry.App)
	suite.registerPacketSniffer(suite.registryChain)

	ibctesting.DefaultTestingAppInit = SetupMonitoredTestingApp
	suite.monitoredChain = ibctesting.NewTestChain(suite.T(), suite.coordinator, appmonitored.Name)
	suite.coordinator.Chains[appmonitored.Name] = suite.monitoredChain
	suite.monitoredApp = suite.monitoredChain.App.(*appmonitored.App)
	suite.registerPacketSniffer(suite.monitoredChain)

	suite.path = ibctesting.NewPath(suite.monitoredChain, suite.registryChain)

	suite.coordinator.SetupClients(suite.path)
	suite.coordinator.CreateConnections(suite.path)

	suite.path.EndpointA.ChannelConfig.PortID = commontypes.MonitoredPortID
	suite.path.EndpointA.ChannelConfig.Order = types.ORDERED
	suite.path.EndpointA.ChannelConfig.Version = commontypes.Version

	suite.path.EndpointB.ChannelConfig.PortID = commontypes.HealthcheckPortID
	suite.path.EndpointB.ChannelConfig.Order = types.ORDERED
	suite.path.EndpointB.ChannelConfig.Version = commontypes.Version

	suite.registryApp.HealthcheckKeeper.SetChain(suite.registryContext(),
		registrytypes.Chain{
			ChainId:      appmonitored.Name,
			ConnectionId: suite.path.EndpointA.ConnectionID,
		})

	suite.coordinator.CreateChannels(suite.path)
}

func (s *HealthcheckTestSuite) registerPacketSniffer(chain *ibctesting.TestChain) {
	if s.packetSniffers == nil {
		s.packetSniffers = make(map[*ibctesting.TestChain]*packetSniffer)
	}
	p := newPacketSniffer()
	chain.App.GetBaseApp().SetStreamingService(p)
	s.packetSniffers[chain] = p
}

func (suite *HealthcheckTestSuite) registryContext() sdk.Context {
	return suite.registryChain.GetContext()
}

func (suite *HealthcheckTestSuite) monitoredContext() sdk.Context {
	return suite.monitoredChain.GetContext()
}

func (suite *HealthcheckTestSuite) relayCommittedPackets(
	srcChain *ibctesting.TestChain,
	path *ibctesting.Path,
	portID string,
	channelID string,
	expectedPackets int,
) {
	commitments := srcChain.App.GetIBCKeeper().ChannelKeeper.GetAllPacketCommitmentsAtChannel(
		suite.monitoredContext(),
		portID,
		channelID,
	)

	suite.Require().Equal(
		expectedPackets,
		len(commitments),
		fmt.Sprintf("expected %d packets, got: %d", expectedPackets, len(commitments)),
	)

	for _, commitment := range commitments {
		packet, found := suite.getSentPacket(srcChain, commitment.Sequence, commitment.ChannelId)

		suite.Require().True(found, "packet not found")

		err := path.RelayPacket(packet)
		suite.Require().NoError(err, "failed to relay packet")
	}
}

func (s *HealthcheckTestSuite) getSentPacket(chain *ibctesting.TestChain, sequence uint64, channelID string) (types.Packet, bool) {
	key := getSentPacketKey(sequence, channelID)
	packet, found := s.packetSniffers[chain].packets[key]

	return packet, found
}

var _ baseapp.StreamingService = &packetSniffer{}

type packetSniffer struct {
	packets map[string]types.Packet
}

func newPacketSniffer() *packetSniffer {
	return &packetSniffer{
		packets: make(map[string]types.Packet),
	}
}

func (ps *packetSniffer) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	packets := ParsePacketsFromEvents(ABCIToSDKEvents(res.GetEvents()))
	for _, packet := range packets {
		ps.packets[getSentPacketKey(packet.Sequence, packet.SourceChannel)] = packet
	}
	return nil
}

// getSentPacketKey returns a key for accessing a sent packet,
// given an ibc sequence number and the channel ID for the source endpoint.
func getSentPacketKey(sequence uint64, channelID string) string {
	return fmt.Sprintf("%s-%d", channelID, sequence)
}

func (*packetSniffer) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	return nil
}

func (*packetSniffer) ListenCommit(ctx context.Context, res abci.ResponseCommit) error {
	return nil
}

func (*packetSniffer) ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	return nil
}
func (*packetSniffer) Close() error                                                  { return nil }
func (*packetSniffer) Listeners() map[storetypes.StoreKey][]storetypes.WriteListener { return nil }
func (*packetSniffer) Stream(wg *sync.WaitGroup) error                               { return nil }

func ABCIToSDKEvents(abciEvents []abci.Event) sdk.Events {
	var events sdk.Events
	for _, evt := range abciEvents {
		var attributes []sdk.Attribute
		for _, attr := range evt.GetAttributes() {
			attributes = append(attributes, sdk.NewAttribute(string(attr.Key), string(attr.Value))) // TODO: is string() ok???
		}

		events = events.AppendEvent(sdk.NewEvent(evt.GetType(), attributes...))
	}

	return events
}

func ParsePacketsFromEvents(events []sdk.Event) (packets []types.Packet) {
	for i, ev := range events {
		if ev.Type == types.EventTypeSendPacket {
			packet, err := ibctesting.ParsePacketFromEvents(events[i:])
			if err != nil {
				panic(err)
			}
			packets = append(packets, packet)
		}
	}
	return
}
