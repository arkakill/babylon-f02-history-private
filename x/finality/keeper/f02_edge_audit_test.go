package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	btcstktypes "github.com/babylonlabs-io/babylon/v4/x/btcstaking/types"
)

func f02DelegationEventsAt(f *f02AuditFixture, btcHeight uint32) []btcstktypes.BTCDelegationStatus {
	events := f.h.BTCStakingKeeper.GetAllPowerDistUpdateEvents(f.h.Ctx, btcHeight, btcHeight)
	states := make([]btcstktypes.BTCDelegationStatus, 0, len(events))
	for _, event := range events {
		update := event.GetBtcDelStateUpdate()
		if update != nil && update.StakingTxHash == f.stakingTxHash {
			states = append(states, update.NewState)
		}
	}
	return states
}

func TestF02QuorumAtExpirySameHeightLeavesNoVotingPower(t *testing.T) {
	f := newF02AuditFixture(t, 2026101101)
	e := f.expiryHeight

	// Record E-1 as the previous Babylon block's BTC tip, but do not process
	// another Babylon block before the quorum arrives at BTC height E.
	f.beginBlock(2, e-1)
	f.tip.Height = e
	f.submitQuorum(t)

	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, e))
	require.ElementsMatch(t,
		[]btcstktypes.BTCDelegationStatus{btcstktypes.BTCDelegationStatus_ACTIVE, btcstktypes.BTCDelegationStatus_EXPIRED},
		f02DelegationEventsAt(f, e),
		"ACTIVE and EXPIRED should both be queued at E before power distribution runs",
	)

	require.NotPanics(t, func() { f.beginBlock(3, e) }, "same-height ACTIVE/EXPIRED processing must not underflow or panic")
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, e))
	f.assertPersistedPower(t, 3, 0)
	require.Empty(t, f02DelegationEventsAt(f, e), "processed events at E must be consumed")
	t.Logf("same-height quorum: state=%s BTC height=%d persisted_power=0 cache=%s table=%v",
		f.derivedStatus(t, e), e, f.cacheSummary(3), f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, 3))
}

func TestF02ActiveAtEminus1AndExpiryAtEProcessedInCatchup(t *testing.T) {
	f := newF02AuditFixture(t, 2026101102)
	e := f.expiryHeight

	// Queue ACTIVE at E-1 after the prior block has recorded E-2. The next
	// Babylon block jumps directly to E, so the keeper catches up both heights.
	f.beginBlock(2, e-2)
	f.tip.Height = e - 1
	f.submitQuorum(t)
	require.Equal(t, btcstktypes.BTCDelegationStatus_ACTIVE, f.derivedStatus(t, e-1))
	require.Equal(t, []btcstktypes.BTCDelegationStatus{btcstktypes.BTCDelegationStatus_ACTIVE}, f02DelegationEventsAt(f, e-1))
	require.Equal(t, []btcstktypes.BTCDelegationStatus{btcstktypes.BTCDelegationStatus_EXPIRED}, f02DelegationEventsAt(f, e))

	require.NotPanics(t, func() { f.beginBlock(3, e) }, "catch-up over E-1 and E must not underflow or panic")
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, e))
	f.assertPersistedPower(t, 3, 0)
	require.Empty(t, f02DelegationEventsAt(f, e-1))
	require.Empty(t, f02DelegationEventsAt(f, e))

	// Replaying the same covenant message is rejected, and another Babylon
	// block at the same BTC tip must not reintroduce power after event cleanup.
	del := f.storedDelegation(t)
	msgs := f.h.GenerateCovenantSignaturesMessages(f.r, f.covenantSKs, del)
	require.NotEmpty(t, msgs)
	_, err := f.h.MsgServer.AddCovenantSigs(f.h.Ctx, msgs[0])
	require.ErrorIs(t, err, btcstktypes.ErrDuplicatedCovenantSig)
	f.beginBlock(4, e)
	require.Equal(t, btcstktypes.BTCDelegationStatus_EXPIRED, f.derivedStatus(t, e))
	f.assertPersistedPower(t, 4, 0)
	require.Empty(t, f02DelegationEventsAt(f, e-1))
	require.Empty(t, f02DelegationEventsAt(f, e))
	t.Logf("catch-up and replay: state=%s BTC height=%d persisted_power=0 cache=%s table=%v",
		f.derivedStatus(t, e), e, f.cacheSummary(4), f.h.FinalityKeeper.GetVotingPowerTable(f.h.Ctx, 4))
}
