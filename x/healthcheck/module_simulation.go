package healthcheck

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"healthcheck/testutil/sample"
	healthchecksimulation "healthcheck/x/healthcheck/simulation"
	"healthcheck/x/healthcheck/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = healthchecksimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgCreateChain = "op_weight_msg_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateChain int = 100

	opWeightMsgUpdateChain = "op_weight_msg_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateChain int = 100

	opWeightMsgDeleteChain = "op_weight_msg_chain"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteChain int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	healthcheckGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		ChainList: []types.Chain{
			{
				Creator: sample.AccAddress(),
				ChainId: "0",
			},
			{
				Creator: sample.AccAddress(),
				ChainId: "1",
			},
		},
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&healthcheckGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateChain int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateChain, &weightMsgCreateChain, nil,
		func(_ *rand.Rand) {
			weightMsgCreateChain = defaultWeightMsgCreateChain
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateChain,
		healthchecksimulation.SimulateMsgCreateChain(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateChain int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateChain, &weightMsgUpdateChain, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateChain = defaultWeightMsgUpdateChain
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateChain,
		healthchecksimulation.SimulateMsgUpdateChain(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteChain int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteChain, &weightMsgDeleteChain, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteChain = defaultWeightMsgDeleteChain
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteChain,
		healthchecksimulation.SimulateMsgDeleteChain(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
