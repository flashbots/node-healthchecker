package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// rethLatestBlock is the latest block as reported by reth
type rethLatestBlock struct {
	Result struct {
		Timestamp string `json:"timestamp"`
	} `json:"result"`
}

func Reth(ctx context.Context, cfg *config.HealthcheckReth) (healthcheck *Result) {
	healthcheck = &Result{Source: SourceReth}

	{ // eth_syncing

		// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_syncing
		// https://github.com/alloy-rs/alloy/blob/v0.3.5/crates/rpc-types-eth/src/syncing.rs#L8-L36

		const ethSyncing = `{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":0}`

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			cfg.BaseURL,
			bytes.NewReader([]byte(ethSyncing)),
		)
		if err != nil {
			healthcheck.Err = err
			return
		}
		req.Header.Set("accept", "application/json; charset=utf-8")
		req.Header.Set("content-type", "application/json; charset=utf-8")

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

		var status rethIsNotSyncing
		if err := json.Unmarshal(body, &status); err != nil {
			var status rethIsSyncing
			if err2 := json.Unmarshal(body, &status); err2 != nil {
				healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					errors.Join(err, err2),
				)
				return
			}
			stages := make([]string, 0, len(status.Result.Stages))
			for idx, stage := range status.Result.Stages {
				stages = append(stages, fmt.Sprintf("%s(%d)=%s", stage.Name, idx, stage.Block))
			}
			healthcheck.Err = fmt.Errorf("still syncing (current: %s, highest: %s): %s",
				status.Result.CurrentBlock,
				status.Result.HighestBlock,
				strings.Join(stages, ", "),
			)
			return
		}
		if status.Result { // i.e. it's syncing
			healthcheck.Err = errors.New("still syncing")
			return
		}
	}

	{ // eth_getBlockByNumber
		const ethGetBlockByNumber = `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":0}`

		if cfg.BlockAgeThreshold != 0 {
			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodPost,
				cfg.BaseURL,
				bytes.NewReader([]byte(ethGetBlockByNumber)),
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

			var latestBlock rethLatestBlock
			if err := json.Unmarshal(body, &latestBlock); err != nil {
				healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					err,
				)
				return
			}

			epoch, err := strconv.ParseInt(
				strings.TrimPrefix(latestBlock.Result.Timestamp, "0x"),
				16, 64,
			)
			if err != nil {
				healthcheck.Err = fmt.Errorf("failed to parse hex timestamp '%s': %w",
					latestBlock.Result.Timestamp,
					err,
				)
				return
			}

			timestamp := time.Unix(epoch, 0)
			age := now.Sub(timestamp)

			if age > cfg.BlockAgeThreshold {
				healthcheck.Err = fmt.Errorf("latest block's timestamp '%s' is too old: %s > %s",
					latestBlock.Result.Timestamp,
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
