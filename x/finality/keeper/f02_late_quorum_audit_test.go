package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	testutil "github.com/babylonlabs-io/babylon/v4/testutil/btcstaking-helper"
	"github.com/babylonlabs-io/babylon/v4/testutil/datagen"
	btclctypes "github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
	btcstktypes "github.com/babylonlabs-io/babylon/v4/x/btcstaking/types"
	ftypes "github.com/babylonlabs-io/babylon/v4/x/finality/types"
)

type f02AuditFixture struct {
	h             *testutil.Helper
	r             *rand.Rand
	covenantSKs   []*btcec.PrivateKey
	fp            *btcstktypes.FinalityProvider
	stakingTxHash string
	stakingValue  uint64
	expiryHeight  uint32
	quorum        uint32
	tip           *btclctypes.BTCHeaderInfo
}

func newF02AuditFixture(t *testing.T, seed int64) *f02AuditFixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	btclcKeeper := btcstktypes.NewMockBTCLightClientKeeper(ctrl)
	btccKeeper := btcstktypes.NewMockBtcCheckpointKeeper(ctrl)
	h := testutil.NewHelper(t, btclcKeeper, btccKeeper, nil)

	hooks := h.FinalityHooks.(*ftypes.MockFinalityHooks)
	hooks.EXPECT().AfterBtcDelegationActivated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	hooks.EXPECT().AfterBtcDelegationUnbonded(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	hooks.EXPECT().AfterBbnFpEntersActiveSet(gomock.Any(), gomock.Any()).AnyTimes()
	hooks.EXPECT().AfterBbnFpRemovedFromActiveSet(gomock.Any(), gomock.Any()).AnyTimes()

	r := rand.New(rand.NewSource(seed))
	covenantSKs, _ := h.GenAndApplyParams(r)
	fpSK, fpPK, fp := h.CreateFinalityProvider(r)

	stakingValue := int64(2 * 10e8)
	delSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	stakingTxHash, _, del, _, _, _, err := h.CreateDelegationWithBtcBlockHeight(
		r,
		delSK,
		fpPK,
		stakingValue,
		1000,
		0,
		0,
		false,
		false,
		10,
		30,
	)
	require.NoError(t, err)

	h.CommitPubRandList(r, fpSK, fp, 1, 100, true)

	params := h.BTCStakingKeeper.GetParams(h.Ctx)
	expiryHeight := del.EndHeight - del.UnbondingTime
	require.Greater(t, expiryHeight, uint32(30))
	expiryEvents := h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(h.Ctx, expiryHeight, expiryHeight)
	require.Len(t, expiryEvents, 1)
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, expiryEvents[0].GetBtcDelStateUpdate().NewState)

	// Keep one mutable tip object so every keeper call observes the height
	// selected by each simulated Babylon block and each covenant transaction.
	tip := &btclctypes.BTCHeaderInfo{Height: 30}
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(tip).AnyTimes()

	return &f02AuditFixture{
		h:             h,
		r:             r,
		covenantSKs:   covenantSKs,
		fp:            fp,
		stakingTxHash: stakingTxHash,
		stakingValue:  uint64(stakingValue),
		expiryHeight:  expiryHeight,
		quorum:        params.CovenantQuorum,
		tip:           tip,
	}
}

func (f *f02AuditFixture) beginBlock(babylonHeight uint64, btcHeight uint32) {
	f.h.SetCtxHeight(babylonHeight)
	f.tip.Height = btcHeight
	f.h.BeginBlocker()
}

func (f *f02AuditFixture) storedDelegation(t *testing.T) *btcstktypes.BTCDelegation {
	t.Helper()
	del, err := f.h.BTCStakingKeeper.GetBTCDelegation(f.h.Ctx, f.stakingTxHash)
	require.NoError(t, err)
	return del
}

func (f *f02AuditFixture) derivedStatus(t *testing.T, btcHeight uint32) btcstktypes.BTCDelegationStatus {
	t.Helper()
	del := f.storedDelegation(t)
	status, err := f.h.BTCStakingKeeper.BtcDelStatus(f.h.Ctx, del, f.quorum, btcHeight)
	require.NoError(t, err)
	return status
}

func (f *f02AuditFixture) submitQuorum(t *testing.T) {
	t.Helper()
	del := f.storedDelegation(t)
	messages := f.h.GenerateCovenantSignaturesMessages(f.r, f.covenantSKs, del)
	require.GreaterOrEqual(t, len(messages), int(f.quorum))
	for _, msg := range messages[:int(f.quorum)] {
		_, err := f.h.MsgServer.AddCovenantSigs(f.h.Ctx, msg)
		require.NoError(t, err)
	}
}

func (f *f02AuditFixture) assertPersistedPower(t *testing.T, babylonHeight uint64, expected uint64) {
	t.Helper()
	cache := f.h.FinalityKeeper.GetVotingPowerDistCache(f.h.Ctx, babylonHeight)
	table := f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, babylonHeight)
	require.Equal(t, expected, f.h.FinalityKeeper.GetVotingPower(f.h.Ctx, f.fp.BtcPk.MustMarshal(), babylonHeight))
	if expected == 0 {
		// An empty protobuf cache marshals to no bytes and is therefore absent
		// from the KV store. Treat absence as the persisted empty distribution.
		if cache != nil {
			require.Zero(t, cache.TotalVotingPower)
			require.Zero(t, cache.NumActiveFps)
			require.Empty(t, cache.FinalityProviders)
		}
		require.Empty(t, table)
		return
	}

	require.NotNil(t, cache, "a non-empty distribution cache must be present in keeper storage")
	require.Equal(t, expected, cache.TotalVotingPower)
	require.Equal(t, uint32(1), cache.NumActiveFps)
	require.Len(t, cache.FinalityProviders, 1)
	require.Equal(t, expected, cache.FinalityProviders[0].TotalBondedSat)
	require.True(t, cache.FinalityProviders[0].IsTimestamped)
	require.Equal(t, map[string]uint64{f.fp.BtcPk.MarshalHex(): expected}, table)
}

func TestF02ControlQuorumBeforeExpiry(t *testing.T) {
	f := newF02AuditFixture(t, 2026101001)
	expiringHeight := f.expiryHeight

	f.beginBlock(2, 30)
	f.beginBlock(3, expiringHeight-1)
	require.Equal(t, btcstktypes.BTCDelegationStatus_PENDING, f.derivedStatus(t, expiringHeight-1))

	f.submitQuorum(t)
	require.Equal(t, btcstktypes.BTCDelegationStatus_ACTIVE, f.derivedStatus(t, expiringHeight-1))
	events := f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, expiringHeight-1, expiringHeight-1)
	require.Len(t, events, 1)
	require.Equal(t, btcstktypes.BTCDelegationStatus_ACTIVE, events[0].GetBtcDelStateUpdate().NewState)

	f.beginBlock(4, expiringHeight-1)
	f.assertPersistedPower(t, 4, f.stakingValue)
	t.Logf("control before expiry: state=%s BTC height=%d stake_sats=%d cache=%s table=%v",
		f.derivedStatus(t, expiringHeight-1), expiringHeight-1, f.stakingValue,
		f.h.FinalityKeeper.GetVotingPowerDistCache(f.h.Ctx, 4).String(),
		f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, 4))

	// Historical revisions define expiry with a strict `>` comparison against
	// EndHeight, so at expiringHeight the derived status can still be ACTIVE.
	// Advance one BTC block beyond the event threshold to make this control's
	// terminal-expiry assertion valid across both the historical and current
	// boundary semantics.
	f.beginBlock(5, expiringHeight+1)
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, expiringHeight+1))
	f.assertPersistedPower(t, 5, 0)
	require.Empty(t, f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, expiringHeight, expiringHeight))
	t.Logf("control after expiry: state=%s BTC height=%d voting_power=0 cache=%s table=%v",
		f.derivedStatus(t, expiringHeight+1), expiringHeight+1,
		f.cacheSummary(5),
		f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, 5))
}

func TestF02LateQuorumAfterExpiryHasNoVotingPower(t *testing.T) {
	f := newF02AuditFixture(t, 2026101002)
	expiringHeight := f.expiryHeight

	f.beginBlock(2, 30)
	f.beginBlock(3, expiringHeight-1)
	f.beginBlock(4, expiringHeight)

	// Without quorum, expiry is consumed but the status calculation still
	// reports PENDING because quorum is checked before the BTC timeout.
	require.Equal(t, btcstktypes.BTCDelegationStatus_PENDING, f.derivedStatus(t, expiringHeight))
	require.Empty(t, f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, expiringHeight, expiringHeight))
	f.assertPersistedPower(t, 4, 0)
	t.Logf("after EXPIRED event without quorum: derived_state=%s BTC height=%d voting_power=0",
		f.derivedStatus(t, expiringHeight), expiringHeight)

	// Advance another Babylon block before submitting the signatures, so the
	// expiration event has already been cleared from the real event store.
	f.beginBlock(5, expiringHeight+1)
	require.Empty(t, f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, expiringHeight, expiringHeight))

	f.submitQuorum(t)
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, expiringHeight+1))
	events := f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, expiringHeight+1, expiringHeight+1)
	require.Len(t, events, 1)
	require.Equal(t, btcstktypes.BTCDelegationStatus_ACTIVE, events[0].GetBtcDelStateUpdate().NewState)

	f.beginBlock(6, expiringHeight+1)
	cache := f.h.FinalityKeeper.GetVotingPowerDistCache(f.h.Ctx, 6)
	table := f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, 6)
	power := f.h.FinalityKeeper.GetVotingPower(f.h.Ctx, f.fp.BtcPk.MustMarshal(), 6)
	cachePower := uint64(0)
	if cache != nil {
		cachePower = cache.TotalVotingPower
	}
	require.Equal(t, power, cachePower)
	if power == 0 {
		if cache != nil {
			require.Zero(t, cache.NumActiveFps)
			require.Empty(t, cache.FinalityProviders)
		}
		require.Empty(t, table)
	} else {
		require.NotNil(t, cache, "a non-empty distribution cache must be present in keeper storage")
		require.Equal(t, uint32(1), cache.NumActiveFps)
		require.Len(t, cache.FinalityProviders, 1)
		require.Equal(t, power, cache.FinalityProviders[0].TotalBondedSat)
		require.True(t, cache.FinalityProviders[0].IsTimestamped)
		require.Equal(t, map[string]uint64{f.fp.BtcPk.MarshalHex(): power}, table)
	}

	t.Logf("after late quorum: derived_state=%s BTC height=%d voting_power=%d cache=%s table=%v",
		f.derivedStatus(t, expiringHeight+1), expiringHeight+1, power, f.cacheSummary(6), table)

	// Security invariant: the expired delegation must not be reintroduced to
	// the voting-power distribution after its EXPIRED event was consumed.
	require.Zero(t, power, "expired delegation regained voting power; cache=%s table=%v", f.cacheSummary(6), table)
}

func (f *f02AuditFixture) cacheSummary(babylonHeight uint64) string {
	cache := f.h.FinalityKeeper.GetVotingPowerDistCache(f.h.Ctx, babylonHeight)
	if cache == nil {
		return "<absent: empty serialized distribution>"
	}
	return cache.String()
}
