package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

func Lighthouse(ctx context.Context, cfg *config.HealthcheckLighthouse) *Result {
	// https://lighthouse-book.sigmaprime.io/api-lighthouse.html#lighthousesyncing
	// https://github.com/sigp/lighthouse/blob/v4.5.0/beacon_node/lighthouse_network/src/types/sync_state.rs#L6-L27

	_url, err := url.JoinPath(cfg.BaseURL, "lighthouse/syncing")
	if err != nil {
		return &Result{Err: err}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		_url,
		nil,
	)
	if err != nil {
		return &Result{Err: err}
	}
	req.Header.Set("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Result{Err: err}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &Result{Err: err}
	}

	if res.StatusCode != http.StatusOK {
		return &Result{Err: fmt.Errorf("unexpected HTTP status '%d': %s",
			res.StatusCode,
			string(body),
		)}
	}

	var state lighthouseStateAsString
	if err := json.Unmarshal(body, &state); err != nil {
		var state lighthouseStateAsStruct
		if err2 := json.Unmarshal(body, &state); err2 != nil {
			return &Result{Err: fmt.Errorf("failed to parse JSON body '%s': %w",
				string(body),
				errors.Join(err, err2),
			)}
		}
		switch {
		case state.Data.BackFillSyncing != nil:
			//
			// BackBillSyncing is "ok" because that's the state lighthouse
			// switches to after checkpoint sync is complete.
			//
			// See: https://lighthouse-book.sigmaprime.io/checkpoint-sync.html#backfilling-blocks
			//
			return &Result{
				Ok: true,
				Err: fmt.Errorf("lighthouse is in 'BackFillSyncing' state (completed: %d, remaining: %d)",
					state.Data.BackFillSyncing.Completed,
					state.Data.BackFillSyncing.Remaining,
				),
			}
		case state.Data.SyncingFinalized != nil:
			return &Result{
				Err: fmt.Errorf("lighthouse is in 'SyncingFinalized' state (start_slot: '%s', target_slot: '%s')",
					state.Data.SyncingFinalized.StartSlot,
					state.Data.SyncingFinalized.TargetSlot,
				),
			}
		case state.Data.SyncingHead != nil:
			return &Result{
				Err: fmt.Errorf("lighthouse is in 'SyncingHead' state (start_slot: '%s', target_slot: '%s')",
					state.Data.SyncingHead.StartSlot,
					state.Data.SyncingHead.TargetSlot,
				),
			}
		default:
			return &Result{
				Err: fmt.Errorf("lighthouse is in unrecognised state: %s",
					string(body),
				),
			}
		}
	}
	if state.Data != "Synced" {
		return &Result{
			Err: fmt.Errorf("lighthouse is not in synced state: %s",
				state.Data,
			),
		}
	}

	return &Result{Ok: true}
}
