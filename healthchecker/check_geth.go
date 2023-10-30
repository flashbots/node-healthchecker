package healthchecker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (h *Healthchecker) checkGeth(ctx context.Context, url string) error {
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_syncing

	const query = `{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":0}`

	type isNotSyncing struct {
		Result bool // `json:"result"`
	}

	type isSyncing struct {
		Result struct {
			CurrentBlock string // `json:"currentBlock"`
			HighestBlock string // `json:"highestBlock"`
		} // `json:"result"`
	}

	//--------------------------------------------------------------------------

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		bytes.NewReader([]byte(query)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

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
			"geth is still syncing (current: %s, highest: %s)",
			status.Result.CurrentBlock,
			status.Result.HighestBlock,
		)
	}
	if status.Result { // i.e. it's syncing
		return errors.New("geth is (still) syncing")
	}

	return nil
}
