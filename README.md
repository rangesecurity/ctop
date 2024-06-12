# ctop
Provides a top-like service for monitoring consensus data of cosmos chains


## Architecture

### Event Subscription Service

The event subscription service connects to tendermint RPC's and streams the following events into a redis queue:

* `NewRound`
* `NewRoundStep`
* `Vote`

### Redis Event Stream Service

The redis event stream service connects to the redis queue and streams events in real-time, persisting them into a postgres database for further analysis.

### Validator Indexing Service

The validator indexing service connects to tendermint RPC's and retrieves the list of validators which are in the active set and persists this information into postgres; Any changes to the active set (adding/removal of validators) is recorded.

### Analyzer Service

The analyzer service provides analysis tools that analyzes indexed data in an attempt to provide insights into conditions that may result in chain halts or other liveliness issues. The following analyzer tools are available:

* Missing Vote Analyzer
  * Provides alerts when validators in the active set fail to vote in at least one round for the most recent block

## Usage

### Requirements

* Postgres
* Redis
* Golang

### Building CLI

To build the cli run the following command from the root folder of the repository

```shell
$> go build
```

### Database Preparation

To prepare the database you need to intialize the migrations table, and run a migration. This is required in order for ctop to be used.

```shell
# initialize migration table
$> ./ctop db --db.url <db_url> init
# run migrations
$> ./ctop db --db.url <db_url> migrate
```

### Event Subscription

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

To launch the redis event stream service which is responsible for streaming events from redis and persisting them into postgres, run the following command.

```shell
$>  ./ctop redis-event-stream --redis.url <redis_url> --db.url <db_url> --networks <networks>
```

Example:

```shell
$> ./ctop redis-event-stream --redis.url localhost:6379 --db.url postgres://postgres:password123@localhost:5432/ctop --networks osmosis
```

### Missing Vote Analyzer

The missing vote analyzer requires two components the validator indexer and the analyzer itself. This service is dependent on the event subscription and redis event stream services.

#### Validator Indexer

In order for the missing vote analyzer to work, the active validator set must be indexed. While it's not needed to constantly run the validator indexer service, it is recommended that you do this so that any changes to the active validator set are recorded.

```shell
$> ./ctop validator-indexer --poll.frequency <frequency> --db.url <db_url> --networks <chain_name>,<chain_rpc>
```

Example:

```shell
$> ./ctop validator-indexer --poll.frequency 60s --db.url postgres://postgres:password123@localhost:5432/ctop --networks osmosis,tcp://osmosis.example.com:8080
```

#### Running The Analyzer

To run the analyzer use the following command. To provide up to date information ensure the redis event stream and event monitoring services are running.


```shell
$> ./ctop analyzer  missing-votes --poll.frequency <frequency> --network <network>
```

Example:

```shell
$> ./ctop analyzer missing-votes --poll.frequency 30s --network osmosis
```