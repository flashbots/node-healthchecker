package healthchecker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (h *Healthchecker) checkLighthouse(ctx context.Context, url string) error {
	// https://lighthouse-book.sigmaprime.io/api-lighthouse.html#lighthousesyncing

	type isNotSyncing struct {
		Data string // `json:"data"`
	}

	type isSyncing struct {
		Data struct {
			SyncingFinalized struct {
				StartSlot  string // `json:"start_slot"`
				TargetSlot string // `json:"target_slot"`
			} // `json:"SyncingFinalized"`
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
		return err
	}
	req.Header.Set("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected HTTP status '%d': %s",
			res.StatusCode,
			string(body),
		)
	}

	var status isNotSyncing
	if err := json.Unmarshal(body, &status); err != nil {
		var status isSyncing
		if err2 := json.Unmarshal(body, &status); err2 != nil {
			return fmt.Errorf(
				"failed to parse JSON body '%s': %w",
				string(body),
				errors.Join(err, err2),
			)
		}
		return fmt.Errorf(
			"lighthouse is still syncing (start: %s, target: %s)",
			status.Data.SyncingFinalized.StartSlot,
			status.Data.SyncingFinalized.TargetSlot,
		)
	}
	if status.Data != "Synced" {
		return fmt.Errorf("lighthouse is not synced: %s", status.Data)
	}

	return nil
}
