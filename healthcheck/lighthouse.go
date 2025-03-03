package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/flashbots/node-healthchecker/config"
)

// lighthouseStateAsString represents sync-state of lighthouse as a string.
//
// Possible values:
//
//   - "Synced" means that lighthouse is up to date with all known peers and is
//     connected to at least one fully synced peer. In this state, parent
//     lookups are enabled.
//
//   - "Stalled" means that no useful peers are connected to lighthouse.
//     Long-range sync's cannot proceed and there are no useful peers to
//     download parents for. More peers need to be connected before lighthouse
//     can proceed.
//
//   - "SyncTransition" means that lighthouse has completed syncing a finalized
//     chain and is in the process of re-evaluating which sync state to progress
//     to.
type lighthouseStateAsString struct {
	Data string `json:"data"`
}

// lighthouseStateAsStruct represents sync-state of lighthouse as a struct.
type lighthouseStateAsStruct struct {
	Data struct {
		// BackFillSyncing means that lighthouse is undertaking a backfill sync.
		//
		// This occurs when a user has specified a trusted state. The node first
		// syncs "forward" by downloading blocks up to the current head as
		// specified by its peers. Once completed, the node enters this sync
		// state and attempts to download all required historical blocks.
		BackFillSyncing *struct {
			Completed uint64 `json:"completed"`
			Remaining uint64 `json:"remaining"`
		} `json:"BackFillSyncing"`

		// SyncingFinalized means that lighthouse is performing a long-range
		// (batch) sync over a finalized chain.
		//
		// In this state, parent lookups are disabled.
		SyncingFinalized *struct {
			StartSlot  string `json:"start_slot"`
			TargetSlot string `json:"target_slot"`
		} `json:"SyncingFinalized"`

		// SyncingHead means that lighthouse is performing a long-range (batch)
		// sync over one or many head chains.
		//
		// In this state parent lookups are disabled.
		SyncingHead *struct {
			StartSlot  string `json:"start_slot"`
			TargetSlot string `json:"target_slot"`
		} `json:"SyncingHead"`
	} `json:"data"`
}

type lighthouseBeaconBlocksHead struct {
	Data struct {
		Message struct {
			Slot string `json:"slot"`

			Body struct {
				ExecutionPayload struct {
					Timestamp string `json:"timestamp"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}

func Lighthouse(ctx context.Context, cfg *config.HealthcheckLighthouse) (healthcheck *Result) {
	healthcheck = &Result{Source: SourceLighthouse}

	{ // lighthouse/syncing

		// https://lighthouse-book.sigmaprime.io/api-lighthouse.html#lighthousesyncing
		// https://github.com/sigp/lighthouse/blob/v4.5.0/beacon_node/lighthouse_network/src/types/sync_state.rs#L6-L27

		_url, err := url.JoinPath(cfg.BaseURL, "lighthouse/syncing")
		if err != nil {
			healthcheck.Err = err
			return
		}

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			_url,
			nil,
		)
		if err != nil {
			healthcheck.Err = err
			return
		}
		req.Header.Set("accept", "application/json")

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

		var state lighthouseStateAsString
		if err := json.Unmarshal(body, &state); err != nil {
			var state lighthouseStateAsStruct
			if err2 := json.Unmarshal(body, &state); err2 != nil {
				healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					errors.Join(err, err2),
				)
				return
			}
			switch {
			case state.Data.BackFillSyncing != nil:
				//
				// BackBillSyncing is "ok" because that's the state lighthouse
				// switches to after checkpoint sync is complete.
				//
				// See: https://lighthouse-book.sigmaprime.io/checkpoint-sync.html#backfilling-blocks
				//
				healthcheck.Ok = true
				healthcheck.Err = fmt.Errorf("is in 'BackFillSyncing' state (completed: %d, remaining: %d)",
					state.Data.BackFillSyncing.Completed,
					state.Data.BackFillSyncing.Remaining,
				)
				return
			case state.Data.SyncingFinalized != nil:
				healthcheck.Err = fmt.Errorf("is in 'SyncingFinalized' state (start_slot: '%s', target_slot: '%s')",
					state.Data.SyncingFinalized.StartSlot,
					state.Data.SyncingFinalized.TargetSlot,
				)
				return
			case state.Data.SyncingHead != nil:
				healthcheck.Err = fmt.Errorf("is in 'SyncingHead' state (start_slot: '%s', target_slot: '%s')",
					state.Data.SyncingHead.StartSlot,
					state.Data.SyncingHead.TargetSlot,
				)
				return
			default:
				healthcheck.Err = fmt.Errorf("is in unrecognised state: %s",
					string(body),
				)
				return
			}
		}
		if state.Data != "Synced" {
			healthcheck.Err = fmt.Errorf("is not in synced state: %s",
				state.Data,
			)
			return
		}
	}

	{ // eth/v2/beacon/blocks/head
		if cfg.BlockAgeThreshold != 0 {
			// https://github.com/sigp/lighthouse/blob/v4.5.0/consensus/types/src/execution_payload.rs#L50-L86

			_url, err := url.JoinPath(cfg.BaseURL, "eth/v2/beacon/blocks/head")
			if err != nil {
				healthcheck.Err = err
				return
			}

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				_url,
				nil,
			)
			if err != nil {
				healthcheck.Err = err
				return
			}
			req.Header.Set("accept", "application/json")

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

			var head lighthouseBeaconBlocksHead
			if err := json.Unmarshal(body, &head); err != nil {
				healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					err,
				)
				return
			}

			epoch, err := strconv.Atoi(head.Data.Message.Body.ExecutionPayload.Timestamp)
			if err != nil {
				healthcheck.Err = fmt.Errorf("failed to parse timestamp '%s': %w",
					head.Data.Message.Body.ExecutionPayload.Timestamp,
					err,
				)
				return
			}
			timestamp := time.Unix(int64(epoch), 0)
			age := now.Sub(timestamp)

			if age > cfg.BlockAgeThreshold {
				healthcheck.Err = fmt.Errorf("beacon head timestamp '%s' (slot '%s') is too old: %s > %s",
					head.Data.Message.Body.ExecutionPayload.Timestamp,
					head.Data.Message.Slot,
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
