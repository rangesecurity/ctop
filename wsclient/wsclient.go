package wsclient

import (
	"context"
	"net/http"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
)

type WsClient struct {
	client *rpcclient.HTTP
}

func NewClient(url string) (*WsClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	client, err := rpcclient.NewWithClient(url, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	if err := client.Start(); err != nil {
		return nil, err
	}
	return &WsClient{client}, nil
}

func (ws *WsClient) SubscribeVotes(ctx context.Context) (<-chan coretypes.ResultEvent, error) {
	return ws.client.Subscribe(ctx, "votesub", types.EventQueryVote.String(), 256)

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
