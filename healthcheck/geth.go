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

// gethIsNotSyncing is the status reported by geth when it's not syncing.
type gethIsNotSyncing struct {
	Result bool `json:"result"`
}

// gethIsSyncing is the status reported by geth when it's in syncing state.
type gethIsSyncing struct {
	Result struct {
		// StartingBlock is the block number where sync began.
		StartingBlock string `json:"startingBlock"`

		// CurrentBlock is a current block number where sync is at.
		CurrentBlock string `json:"currentBlock"`

		// HighestBlock is the highest alleged block number in the chain.
		HighestBlock string `json:"highestBlock"`

		// SyncedAccounts is a number of accounts downloaded (snap sync).
		SyncedAccounts string `json:"syncedAccounts"`

		// Number of account trie bytes persisted to disk (snap sync).
		SyncedAccountBytes string `json:""`

		// SyncedBytecodes is a number of bytecodes downloaded (snap sync).
		SyncedBytecodes string `json:"syncedBytecodes"`

		// SyncedBytecodeBytes is a number of bytecode bytes downloaded (snap sync).
		SyncedBytecodeBytes string `json:"syncedBytecodeBytes"`

		// SyncedStorage is a number of storage slots downloaded (snap sync).
		SyncedStorage string `json:"syncedStorage"`

		// SyncedStorageBytes is a number of storage trie bytes persisted to disk (snap sync).
		SyncedStorageBytes string `json:"syncedStorageBytes"`

		HealedTrienodes     string `json:"healingTrienodes"`
		HealedTrienodeBytes string `json:"healedTrienodeBytes"`
		HealedBytecodes     string `json:"healedBytecodes"`
		HealedBytecodeBytes string `json:"healedBytecodeBytes"`

		HealingTrienodes string `json:"healedTrienodes"`
		HealingBytecode  string `json:"healingBytecode"`

		TxIndexFinishedBlocks  string `json:"txIndexFinishedBlocks"`
		TxIndexRemainingBlocks string `json:"txIndexRemainingBlocks"`
	} `json:"result"`
}

// gethLatestBlock is the latest block as reported by geth
type gethLatestBlock struct {
	Result struct {
		Timestamp string `json:"timestamp"`
	} `json:"result"`
}

func Geth(ctx context.Context, cfg *config.HealthcheckGeth) (healthcheck *Result) {
	healthcheck = &Result{Source: SourceGeth}

	{ // eth_syncing

		// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_syncing
		// https://github.com/ethereum/go-ethereum/blob/v1.14.8/interfaces.go#L98-L127

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
		req.Header.Set("accept", "application/json")
		req.Header.Set("content-type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			healthcheck.Err = err
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			healthcheck.Err = err
		}

		if res.StatusCode != http.StatusOK {
			healthcheck.Err = fmt.Errorf("unexpected HTTP status '%d': %s",
				res.StatusCode,
				string(body),
			)
			return
		}

		var status gethIsNotSyncing
		if err := json.Unmarshal(body, &status); err != nil {
			var status gethIsSyncing
			if err2 := json.Unmarshal(body, &status); err2 != nil {
				healthcheck.Err = fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					errors.Join(err, err2),
				)
				return
			}
			healthcheck.Err = fmt.Errorf("still syncing (current: '%s', highest: '%s')",
				status.Result.CurrentBlock,
				status.Result.HighestBlock,
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

			var latestBlock gethLatestBlock
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
