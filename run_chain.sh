#!/bin/bash
set -eu

BINARY=$1
CHAIN_ID=$2
MONIKER=$3

$BINARY init $MONIKER --chain-id $CHAIN_ID
$BINARY keys add wallet1 --keyring-backend test
$BINARY keys add wallet2 --keyring-backend test
$BINARY add-genesis-account wallet1 1000000000000stake --keyring-backend test
$BINARY add-genesis-account wallet2 1000000000000stake --keyring-backend test
$BINARY gentx wallet1 1000000000stake --keyring-backend test --chain-id $CHAIN_ID
$BINARY collect-gentxs

$BINARY start