package wsclient

import (
	"context"

	rpcclient "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
)

type WsClient struct {
	client *rpcclient.HTTP
}

func NewClient(url string) (*WsClient, error) {
	client, err := rpcclient.New(url, "/websocket")
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
