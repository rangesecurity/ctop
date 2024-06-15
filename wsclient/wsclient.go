package wsclient

import (
	"context"
	"net/http"
	"strings"

	rpcclient "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
)

type WsClient struct {
	client *rpcclient.HTTP
}

func NewClient(url string, authToken string) (*WsClient, error) {
	var (
		client *rpcclient.HTTP
		err    error
	)
	if authToken != "" {
		authTransport := &AuthTransport{
			Transport: http.DefaultTransport,
			Token:     authToken,
		}
		client, err = rpcclient.NewWithClient(url, "/websocket", &http.Client{
			Transport: authTransport,
		})
	} else {
		client, err = rpcclient.New(url, "/websocket")
	}
	if err != nil {
		return nil, err
	}
	if err := client.Start(); err != nil {
		return nil, err
	}
	return &WsClient{client}, nil
}

func (ws *WsClient) SubscribeVotes(ctx context.Context) (<-chan coretypes.ResultEvent, error) {
	return ws.client.Subscribe(ctx, "votesub", types.EventQueryVote.String(), 1024)

}

func (ws *WsClient) UnsubscribeVotes(ctx context.Context) error {
	return ws.client.Unsubscribe(ctx, "", types.EventQueryVote.String())
}

func (ws *WsClient) SubscribeNewRound(ctx context.Context) (<-chan coretypes.ResultEvent, error) {
	return ws.client.Subscribe(ctx, "roundsub", types.EventQueryNewRound.String(), 256)

}

func (ws *WsClient) UnsubscribeNewRound(ctx context.Context) error {
	return ws.client.Unsubscribe(ctx, "", types.EventQueryNewRound.String())
}

func (ws *WsClient) SubscribeNewRoundStep(ctx context.Context) (<-chan coretypes.ResultEvent, error) {
	return ws.client.Subscribe(ctx, "roundstepsub", types.EventQueryNewRoundStep.String(), 256)

}

func (ws *WsClient) UnsubscribeNewRoundStep(ctx context.Context) error {
	return ws.client.Unsubscribe(ctx, "", types.EventQueryNewRoundStep.String())
}

func (ws *WsClient) Validators(ctx context.Context) ([]*types.Validator, error) {
	var (
		page       int
		perPage    int = 100
		validators     = make([]*types.Validator, 0, 200)
	)
	// question: is it sufficient to cap pages to 4? not aware of a cosmos chain with more than 200 validators
	for page = 1; page < 5; page++ {
		validatorRes, err := ws.client.Validators(ctx, nil, &page, &perPage)
		if err != nil {
			// if this error happens we have finished enumerating the validator set
			if strings.Contains(err.Error(), "page should be within") {
				break
			}
			return nil, err
		}
		validators = append(validators, validatorRes.Validators...)
	}
	return validators, nil
}
