package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flashbots/node-healthchecker/config"
)

// opNodeSyncStatus is a snapshot of the op-node's driver.
//
// Values may be zeroed if not yet initialized.
type opNodeSyncStatus struct {
	Result struct {
		// CurrentL1 is the L1 block that the derivation process is last idled
		// at.
		//
		// This may not be fully derived into L2 data yet.
		//
		// The safe L2 blocks were produced/included fully from the L1 chain up
		// to and including this L1 block.
		//
		// If the node is synced, this matches the HeadL1, minus the verifier
		// confirmation distance.
		CurrentL1 opNodeL1BlockRef `json:"current_l1"`

		// HeadL1 is the perceived head of the L1 chain, no confirmation
		// distance.
		//
		// The head is not guaranteed to build on the other L1 sync status
		// fields, as the node may be in progress of resetting to adapt to a L1
		// reorg.
		HeadL1 opNodeL1BlockRef `json:"head_l1"`

		SafeL1 opNodeL1BlockRef `json:"safe_l1"`

		FinalizedL1 opNodeL1BlockRef `json:"finalized_l1"`

		// UnsafeL2 is the absolute tip of the L2 chain, pointing to block data
		// that has not been submitted to L1 yet.
		//
		// The sequencer is building this, and verifiers may also be ahead of
		// the SafeL2 block if they sync blocks via p2p or other offchain
		// sources.
		//
		// This is considered to only be local-unsafe post-interop, see
		// CrossUnsafe for cross-L2 guarantees.
		UnsafeL2 opNodeL2BlockRef `json:"unsafe_l2"`

		// SafeL2 points to the L2 block that was derived from the L1 chain.
		//
		// This point may still reorg if the L1 chain reorgs.
		//
		// This is considered to be cross-safe post-interop, see LocalSafe to
		// ignore cross-L2 guarantees.
		SafeL2 opNodeL2BlockRef `json:"safe_l2"`

		// FinalizedL2 points to the L2 block that was derived fully from
		// finalized L1 information, thus irreversible.
		FinalizedL2 opNodeL2BlockRef `json:"finalized_l2"`

		// PendingSafeL2 points to the L2 block processed from the batch, but
		// not consolidated to the safe block yet.
		PendingSafeL2 opNodeL2BlockRef `json:"pending_safe_l2"`

		// CrossUnsafeL2 is an unsafe L2 block, that has been verified to match
		// cross-L2 dependencies.
		//
		// Pre-interop every unsafe L2 block is also cross-unsafe.
		CrossUnsafeL2 opNodeL2BlockRef `json:"cross_unsafe_l2"`

		// LocalSafeL2 is an L2 block derived from L1, not yet verified to have
		// valid cross-L2 dependencies.
		LocalSafeL2 opNodeL2BlockRef `json:"local_safe_l2"`
	} `json:"result"`
}

type opNodeL1BlockRef struct {
	Hash       string `json:"hash"`
	Number     uint64 `json:"number"`
	ParentHash string `json:"parentHash"`
	Time       uint64 `json:"timestamp"`
}

type opNodeL2BlockRef struct {
	Hash           string `json:"hash"`
	Number         uint64 `json:"number"`
	ParentHash     string `json:"parentHash"`
	SequenceNumber uint64 `json:"sequenceNumber"` // distance to first block of epoch
	Time           uint64 `json:"timestamp"`

	L1Origin struct {
		Hash   string `json:"hash"`
		Number uint64 `json:"number"`
	} `json:"l1origin"`
}

func OpNode(ctx context.Context, cfg *config.HealthcheckOpNode) (healthcheck *Result) {
	healthcheck = &Result{Source: SourceOpNode}

	{ // optimism_syncStatus

		// https://docs.optimism.io/builders/node-operators/json-rpc#optimism_syncstatus
		// https://github.com/ethereum-optimism/optimism/blob/v1.9.1/op-service/eth/sync_status.go#L5-L34

		const optimismSyncStatus = `{"jsonrpc":"2.0","method":"optimism_syncStatus","params":[],"id":0}`

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			cfg.BaseURL,
			bytes.NewReader([]byte(optimismSyncStatus)),
		)
		if err != nil {
			healthcheck.Err = err
			return
		}
		req.Header.Set("accept", "application/json")
		req.Header.Set("content-type", "application/json")

		now := time.Now()
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			healthcheck.Err = err
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			healthcheck.Err = err
			return
		}

		if res.StatusCode != http.StatusOK {
			healthcheck.Err = fmt.Errorf("unexpected HTTP status '%d': %s",
				res.StatusCode,
				string(body),
			)
			return
		}

		var status opNodeSyncStatus
		err = json.Unmarshal(body, &status)
		if err != nil {
			healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
				string(body),
				err,
			)
			return
		}

		if status.Result.CurrentL1.Number > status.Result.HeadL1.Number {
			dist := status.Result.CurrentL1.Number - status.Result.HeadL1.Number
			if dist == 1 {
				healthcheck.Ok = true
				return
			}

			healthcheck.Ok = true
			healthcheck.Err = fmt.Errorf("current l1 block (number: %d, hash: %s) is greater than head (number: %d, hash %s): %d - %d = %d",
				status.Result.CurrentL1.Number, status.Result.CurrentL1.Hash,
				status.Result.HeadL1.Number, status.Result.HeadL1.Hash,
				status.Result.CurrentL1.Number, status.Result.HeadL1.Number, dist,
			)
			return
		}

		dist := status.Result.HeadL1.Number - status.Result.CurrentL1.Number
		if dist > cfg.ConfirmationDistance {
			healthcheck.Err = fmt.Errorf("current l1 block (number: %d, hash: %s) is behind the l1 head (number: %d, hash: %s) for more than confirmation distance: %d > %d",
				status.Result.CurrentL1.Number, status.Result.CurrentL1.Hash,
				status.Result.HeadL1.Number, status.Result.HeadL1.Hash,
				dist, cfg.ConfirmationDistance,
			)
			return
		}

		if cfg.BlockAgeThreshold != 0 {
			timestamp := time.Unix(int64(status.Result.UnsafeL2.Time), 0)
			age := now.Sub(timestamp)

			if age > cfg.BlockAgeThreshold {
				healthcheck.Err = fmt.Errorf("latest l2 unsafe timestamp %d is too old: %s > %s",
					status.Result.UnsafeL2.Time,
					age,
					cfg.BlockAgeThreshold,
				)
				return
			}
		}
	}

	healthcheck.Ok = true
	return
}
