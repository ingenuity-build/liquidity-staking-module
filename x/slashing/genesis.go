package slashing

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/types"
	stakingtypes "github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, stakingKeeper types.StakingKeeper, data *types.GenesisState) {
	stakingKeeper.IterateValidators(ctx,
		func(index int64, validator stakingtypes.ValidatorI) bool {
			consPk, err := validator.ConsPubKey()
			if err != nil {
				panic(err)
			}
			keeper.AddPubkey(ctx, consPk)
			return false
		},
	)

	for _, info := range data.SigningInfos {
		address, err := sdk.ConsAddressFromBech32(info.Address)
		if err != nil {
			panic(err)
		}
		keeper.SetValidatorSigningInfo(ctx, address, info.ValidatorSigningInfo)
	}

	for _, array := range data.MissedBlocks {
		address, err := sdk.ConsAddressFromBech32(array.Address)
		if err != nil {
			panic(err)
		}
		for _, missed := range array.MissedBlocks {
			keeper.SetValidatorMissedBlockBitArray(ctx, address, missed.Index, missed.Missed)
		}
	}

	epochNumber := stakingKeeper.GetEpochNumber(ctx)

	for _, msg := range data.BufferedMsgs {
		keeper.RestoreEpochAction(ctx, epochNumber, msg)
	}

	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) (data *types.GenesisState) {
	params := keeper.GetParams(ctx)
	signingInfos := make([]types.SigningInfo, 0)
	missedBlocks := make([]types.ValidatorMissedBlocks, 0)
	keeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		bechAddr := address.String()
		signingInfos = append(signingInfos, types.SigningInfo{
			Address:              bechAddr,
			ValidatorSigningInfo: info,
		})

		localMissedBlocks := keeper.GetValidatorMissedBlocks(ctx, address)

		missedBlocks = append(missedBlocks, types.ValidatorMissedBlocks{
			Address:      bechAddr,
			MissedBlocks: localMissedBlocks,
		})

		return false
	})
	msgs := keeper.GetEpochActions(ctx)

	var anys []*codectypes.Any
	for _, msg := range msgs {
		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		anys = append(anys, any)
	}

	return types.NewGenesisState(params, signingInfos, missedBlocks, anys)
}