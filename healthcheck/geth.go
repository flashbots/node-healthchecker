package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

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

func Geth(ctx context.Context, cfg *config.HealthcheckGeth) *Result {
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_syncing
	// https://github.com/ethereum/go-ethereum/blob/v1.14.8/interfaces.go#L98-L127

	const query = `{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":0}`

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		cfg.BaseURL,
		bytes.NewReader([]byte(query)),
	)
	if err != nil {
		return &Result{Err: err}
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

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

	var status gethIsNotSyncing
	if err := json.Unmarshal(body, &status); err != nil {
		var status gethIsSyncing
		if err2 := json.Unmarshal(body, &status); err2 != nil {
			return &Result{
				Err: fmt.Errorf("failed to parse JSON body '%s': %w",
					string(body),
					errors.Join(err, err2),
				),
			}
		}
		return &Result{
			Err: fmt.Errorf("geth is still syncing: Current:=%s, Highest=%s",
				status.Result.CurrentBlock,
				status.Result.HighestBlock,
			),
		}
	}
	if status.Result { // i.e. it's syncing
		return &Result{Err: errors.New("geth is (still) syncing")}
	}

	return &Result{Ok: true}
}
