package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flashbots/node-healthchecker/config"
)

// rethIsNotSyncing is the status reported by reth when it's not syncing.
type rethIsNotSyncing struct {
	Result bool `json:"result"`
}

// rethIsSyncing is the status reported by reth when it is in syncing state.
type rethIsSyncing struct {
	Result struct {
		// StartingBlock is a starting block.
		StartingBlock string `json:"startingBlock"`

		// CurrentBlock is a current block.
		CurrentBlock string `json:"currentBlock"`

		// HighestBlock is the highest block seen so far.
		HighestBlock string `json:"highestBlock"`

		// WarpChunksAmount is a warp-sync snapshot chunks total.
		WarpChunksAmount *string `json:"warpChunksAmount,omitempty"`

		// WarpChunksProcessed is a warp-sync snapshot chunks processed.
		WarpChunksProcessed *string `json:"warpChunksProcessed,omitempty"`

		/// Stages contains the details of the sync-stages.
		Stages []struct {
			// Name of the sync-stage.
			Name string `json:"name"`

			// Block indicates the progress of the sync-stage.
			Block string `json:"block"`
		} `json:"stages"`
	} `json:"result"`
}

func Reth(ctx context.Context, cfg *config.HealthcheckReth) *Result {
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_syncing
	// https://github.com/alloy-rs/alloy/blob/v0.3.5/crates/rpc-types-eth/src/syncing.rs#L8-L36

	const query = `{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":0}`

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.BaseURL,
		bytes.NewReader([]byte(query)),
	)
	if err != nil {
		return &Result{Err: err}
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	req.Header.Set("content-type", "application/json; charset=utf-8")

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
		return &Result{
			Err: fmt.Errorf("unexpected HTTP status '%d': %s",
				res.StatusCode,
				string(body),
			),
		}
	}

	var status rethIsNotSyncing
	if err := json.Unmarshal(body, &status); err != nil {
		var status rethIsSyncing
		if err2 := json.Unmarshal(body, &status); err2 != nil {
			return &Result{
				Err: fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					errors.Join(err, err2),
				),
			}
		}
		stages := make([]string, 0, len(status.Result.Stages))
		for idx, stage := range status.Result.Stages {
			stages = append(stages, fmt.Sprintf("%s(%d)=%s", stage.Name, idx, stage.Block))
		}
		return &Result{
			Err: fmt.Errorf("reth is still syncing: Current=%s, Highest=%s, %s",
				status.Result.CurrentBlock,
				status.Result.HighestBlock,
				strings.Join(stages, ", "),
			),
		}
	}
	if status.Result { // i.e. it's syncing
		return &Result{Err: errors.New("reth is (still) syncing")}
	}

	return &Result{Ok: true}
}
