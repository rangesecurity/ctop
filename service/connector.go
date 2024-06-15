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

func NewConnector(ctx context.Context, network, url, authToken string) (*Connector, error) {
	ctx, cancel := context.WithCancel(ctx)
	client, err := wsclient.NewClient(url, authToken)
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

// Starts the connector event loop, which subscribes to Vote, NewRound, and NewRoundStep events
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

// Returns a channel that can be used to retrieve vote events
func (c *Connector) GetVotes() <-chan types.EventDataVote {
	return c.voteCh
}

// Returns a channel that can be used to retrieve NewRound events
func (c *Connector) GetNewRounds() <-chan types.EventDataNewRound {
	return c.newRoundCh
}

// Returns a channel that can be used to retrieve NewRoundStep events
func (c *Connector) GetNewRoundSteps() <-chan types.EventDataRoundState {
	return c.newRoundStepCh
}

// Returns the network this connector is for
func (c *Connector) Network() string {
	return c.network
}

// Returns all currently active validators for this network
func (c *Connector) Validators() ([]*types.Validator, error) {
	return c.wsClient.Validators(c.ctx)
}

func (c *Connector) Close() {
	c.cancel()
}
