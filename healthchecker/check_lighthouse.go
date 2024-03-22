package healthchecker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (h *Healthchecker) checkLighthouse(ctx context.Context, url string) *healthcheckResult {
	// https://lighthouse-book.sigmaprime.io/api-lighthouse.html#lighthousesyncing
	// https://github.com/sigp/lighthouse/blob/v4.5.0/beacon_node/lighthouse_network/src/types/sync_state.rs#L6-L27

	type stateString struct {
		Data string // `json:"data"`
	}

	type stateStruct struct {
		Data struct {
			BackFillSyncing *struct {
				Completed uint64 // `json:"completed"`
				Remaining uint64 // `json:"remaining"`
			} // `json:"BackFillSyncing"`

			SyncingFinalized *struct {
				StartSlot  string // `json:"start_slot"`
				TargetSlot string // `json:"target_slot"`
			} // `json:"SyncingFinalized"`

			SyncingHead *struct {
				StartSlot  string // `json:"start_slot"`
				TargetSlot string // `json:"target_slot"`
			} // `json:"SyncingHead"`
		} // `json:"data"`
	}

	//--------------------------------------------------------------------------

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return &healthcheckResult{err: err}
	}
	req.Header.Set("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &healthcheckResult{err: err}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &healthcheckResult{err: err}
	}

	if res.StatusCode != http.StatusOK {
		return &healthcheckResult{err: fmt.Errorf(
			"unexpected HTTP status '%d': %s",
			res.StatusCode,
			string(body),
		)}
	}

	var state stateString
	if err := json.Unmarshal(body, &state); err != nil {
		var state stateStruct
		if err2 := json.Unmarshal(body, &state); err2 != nil {
			return &healthcheckResult{err: fmt.Errorf(
				"failed to parse JSON body '%s': %w",
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
			return &healthcheckResult{ok: true, err: fmt.Errorf(
				"lighthouse is in 'BackFillSyncing' state (completed: %d, remaining: %d)",
				state.Data.BackFillSyncing.Completed,
				state.Data.BackFillSyncing.Remaining,
			)}
		case state.Data.SyncingFinalized != nil:
			return &healthcheckResult{err: fmt.Errorf(
				"lighthouse is in 'SyncingFinalized' state (start_slot: '%s', target_slot: '%s')",
				state.Data.SyncingFinalized.StartSlot,
				state.Data.SyncingFinalized.TargetSlot,
			)}
		case state.Data.SyncingHead != nil:
			return &healthcheckResult{err: fmt.Errorf(
				"lighthouse is in 'SyncingHead' state (start_slot: '%s', target_slot: '%s')",
				state.Data.SyncingHead.StartSlot,
				state.Data.SyncingHead.TargetSlot,
			)}
		default:
			return &healthcheckResult{err: fmt.Errorf(
				"lighthouse is in unrecognised state: %s",
				string(body),
			)}
		}
	}
	if state.Data != "Synced" {
		return &healthcheckResult{err: fmt.Errorf(
			"lighthouse is not in synced state: %s",
			state.Data,
		)}
	}

	return &healthcheckResult{ok: true}
}
