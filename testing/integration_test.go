package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	appmonitored "healthcheck/app/monitored"
	registrytypes "healthcheck/x/healthcheck/types"
	commontypes "healthcheck/x/types"
)

func TestHealthcheck(t *testing.T) {
	healthcheckSuite := new(HealthcheckTestSuite)
	suite.Run(t, healthcheckSuite)
}

func (s *HealthcheckTestSuite) TestHappyPath() {
	monitoredChain1 := GetMonitoredChain(s, appmonitored.Name)
	latestReportedBlock := monitoredChain1.Block
	s.Require().Equal(uint64(0), latestReportedBlock)

	// one packet was already sent by the monitored chain, since chain was moved forward
	// by the coordinator during channel handshake
	s.relayCommittedPackets(s.monitoredChain, s.path, commontypes.MonitoredPortID, s.path.EndpointA.ChannelID, 1)

	monitoredChain1 = GetMonitoredChain(s, appmonitored.Name)
	s.Require().Greater(monitoredChain1.Block, latestReportedBlock)
	latestReportedBlock = monitoredChain1.Block

	// advance the monitored chain, but not too much, so we get only one healthcheck packet
	s.coordinator.CommitNBlocks(s.monitoredChain, 3)
	s.relayCommittedPackets(s.monitoredChain, s.path, commontypes.MonitoredPortID, s.path.EndpointA.ChannelID, 1)

	monitoredChain1 = GetMonitoredChain(s, appmonitored.Name)
	s.Require().Greater(monitoredChain1.Block, latestReportedBlock)
}

func GetMonitoredChain(s *HealthcheckTestSuite, chainID string) registrytypes.Chain {
	monitoredChain1, found := s.registryApp.HealthcheckKeeper.GetChain(s.registryContext(), chainID)
	s.Require().True(found, fmt.Sprintf("chain with id: '%s' not found", appmonitored.Name))

	return monitoredChain1
}
