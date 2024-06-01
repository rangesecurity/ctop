package service

import (
	"context"

	"github.com/cometbft/cometbft/types"
	"github.com/rangesecurity/ctop/wsclient"
)

// connects to a single chain
type Connector struct {
	voteCh         chan types.EventDataVote
	newRoundCh     chan types.EventDataNewRound
	newRoundStepCh chan types.EventDataRoundState
	wsClient       *wsclient.WsClient
	network        string
	cancel         context.CancelFunc
	ctx            context.Context
}

func NewConnector(ctx context.Context, network string, url string) (*Connector, error) {
	ctx, cancel := context.WithCancel(ctx)
	client, err := wsclient.NewClient(url)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Connector{
		wsClient:       client,
		network:        network,
		newRoundCh:     make(chan types.EventDataNewRound, 256),
		voteCh:         make(chan types.EventDataVote, 1024),
		newRoundStepCh: make(chan types.EventDataRoundState, 256),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func (c *Connector) Start() error {
	votes, err := c.wsClient.SubscribeVotes(c.ctx)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg := <-votes:
				if voteInfo, ok := msg.Data.(types.EventDataVote); ok {
					c.voteCh <- voteInfo
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	newRounds, err := c.wsClient.SubscribeNewRound(c.ctx)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case msg := <-newRounds:
				if roundInfo, ok := msg.Data.(types.EventDataNewRound); ok {
					c.newRoundCh <- roundInfo
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	newRoundSteps, err := c.wsClient.SubscribeNewRoundStep(c.ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-newRoundSteps:
				if roundInfo, ok := msg.Data.(types.EventDataRoundState); ok {
					c.newRoundStepCh <- roundInfo
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (c *Connector) GetVotes() <-chan types.EventDataVote {
	return c.voteCh
}

func (c *Connector) GetNewRounds() <-chan types.EventDataNewRound {
	return c.newRoundCh
}

func (c *Connector) GetNewRoundSteps() <-chan types.EventDataRoundState {
	return c.newRoundStepCh
}

func (c *Connector) Network() string {
	return c.network
}

func (c *Connector) Close() {
	c.cancel()
}
