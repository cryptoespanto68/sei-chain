package main

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/utils"
)

type Config struct {
	ChainID       string                `json:"chain_id"`
	TxsPerBlock   uint64                `json:"txs_per_block"`
	MsgsPerTx     uint64                `json:"msgs_per_tx"`
	Rounds        uint64                `json:"rounds"`
	MessageType   string                `json:"message_type"`
	PriceDistr    NumericDistribution   `json:"price_distribution"`
	QuantityDistr NumericDistribution   `json:"quantity_distribution"`
	MsgTypeDistr  MsgTypeDistribution   `json:"message_type_distribution"`
	ContractDistr ContractDistributions `json:"contract_distribution"`
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	// NOTE: this field will be renamed to Codec
	Marshaler codec.Codec
	TxConfig  client.TxConfig
	Amino     *codec.LegacyAmino
}

type NumericDistribution struct {
	Min         sdk.Dec `json:"min"`
	Max         sdk.Dec `json:"max"`
	NumDistinct int64   `json:"number_of_distinct_values"`
}

func (d *NumericDistribution) Sample() sdk.Dec {
	steps := sdk.NewDec(rand.Int63n(d.NumDistinct))
	return d.Min.Add(d.Max.Sub(d.Min).QuoInt64(d.NumDistinct).Mul(steps))
}

type MsgTypeDistribution struct {
	LimitOrderPct      sdk.Dec `json:"limit_order_percentage"`
	MarketOrderPct     sdk.Dec `json:"market_order_percentage"`
	DelegatePct        sdk.Dec `json:"delegate_percentage"`
	UndelegatePct      sdk.Dec `json:"undelegate_percentage"`
	BeginRedelegatePct sdk.Dec `json:"begin_redelegate_percentage"`
}

func (d *MsgTypeDistribution) SampleDexMsgs() string {
	if !d.LimitOrderPct.Add(d.MarketOrderPct).Equal(sdk.OneDec()) {
		panic("Distribution percentages must add up to 1")
	}
	randNum := sdk.MustNewDecFromStr(fmt.Sprintf("%f", rand.Float64()))
	if randNum.LT(d.LimitOrderPct) {
		return "limit"
	}
	return "market"
}

func (d *MsgTypeDistribution) SampleStakingMsgs() string {
	if !d.DelegatePct.Add(d.UndelegatePct).Add(d.BeginRedelegatePct).Equal(sdk.OneDec()) {
		panic("Distribution percentages must add up to 1")
	}
	randNum := sdk.MustNewDecFromStr(fmt.Sprintf("%f", rand.Float64()))
	if randNum.LT(d.DelegatePct) {
		return "delegate"
	} else if randNum.LT(d.DelegatePct.Add(d.UndelegatePct)) {
		return "undelegate"
	}
	return "begin_redelegate"
}

type ContractDistributions []ContractDistribution

func (d *ContractDistributions) Sample() string {
	if !utils.Reduce(*d, func(i ContractDistribution, o sdk.Dec) sdk.Dec { return o.Add(i.Percentage) }, sdk.ZeroDec()).Equal(sdk.OneDec()) {
		panic("Distribution percentages must add up to 1")
	}
	randNum := sdk.MustNewDecFromStr(fmt.Sprintf("%f", rand.Float64()))
	cumPct := sdk.ZeroDec()
	for _, dist := range *d {
		cumPct = cumPct.Add(dist.Percentage)
		if randNum.LTE(cumPct) {
			return dist.ContractAddr
		}
	}
	panic("this should never be triggered")
}

type ContractDistribution struct {
	ContractAddr string  `json:"contract_address"`
	Percentage   sdk.Dec `json:"percentage"`
}