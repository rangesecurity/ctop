# ctop
Provides a top-like service for monitoring consensus data of cosmos chains


## Architecture

### Event Detection

Event detection consists of two services, the main subscription service is responsible for subscribing to events, and pushing received events into a redis queue, while the secondary service streams events from the redis queue persisting the events into Postgres for further analysis. The following event subscriptions exist:

* `NewRound`
* `NewRoundStep`
* `Vote`

### Validator Indexing

The validator indexing service maintains a list of validators which are in the active set. This is used for implementing analysis tools like the missing vote analyzer.

## Analyzer

The analyzer provides analysis tools that analyzes indexed events in an attempt to provide insights into conditions that may result in chain halts or other liveliness issues.

### Missing Vote Analyzer

The missing vote analyzer is used to provide alerts when validators in the active set fail to vote in at least one round for the most recent block.

## Usage

### Requirements

* Postgres
* Redis

### Event Monitoring

To launch the event subscription service which monitors `Vote`, `NewRound` and `NewRoundStep` events, run the following command.

You must specify at minimum one pair of `chain_name` and `chain_rpc`, but you can specify any number of pairs to monitor more than one network

```shell
$> ./ctop event-subscription-service --networks <chain_name>,<chain_rpc> --redis.url <redis_url>
```

Example:

```shell
$>  ./ctop event-subscription-service --networks osmosis,tcp://osmosis.example.com:8080 --redis.url localhost:6379
```

### Redis Event Stream

To stream events that are stored in redis, and persist them into postgres run the following command

```shell
$>  ./ctop redis-event-stream --redis.url <redis_url> --db.url <db_url> --networks <networks>
```

Example:

```shell
$> ./ctop redis-event-stream --redis.url localhost:6379 --db.url postgres://postgres:password123@localhost:5432/ctop --networks osmosis
```

### Missing Vote Analyzer

The missing vote analyzer requires two components the validator indexer and the analyzer itself.

#### Validator Indexer

In order for the missing vote analyzer, the active validator set must be indexed. To do this on a periodic basis you can run the following command

```shell
$> ./ctop validator-indexer --poll.frequency <frequency> --db.url <db_url> --networks <chain_name>,<chain_rpc>
```

Example:

```shell
$> ./ctop validator-indexer --poll.frequency 60s --db.url postgres://postgres:password123@localhost:5432/ctop --networks osmosis,tcp://osmosis.example.com:8080
```

### Running The Analyzer

To run the analyzer run the following command. To provide up to date information ensure the redis event stream and event monitoring services are running


```shell
$> ./ctop analyzer  missing-votes --poll.frequency <frequency> --network <network>
```

Example:

```shell
$> ./ctop analyzer missing-votes --poll.frequency 30s --network osmosis
```