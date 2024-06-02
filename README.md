# ctop
Provides a top-like service for monitoring consensus data of cosmos chains


# Usage

## Requirements

* Postgres
* Redis

## Event Monitoring

To launch the event subscription service which monitors `Vote`, `NewRound` and `NewRoundStep` events, run the following command.

You must specify at minimum one pair of `chain_name` and `chain_rpc`, but you can specify any number of pairs to monitor more than one network

```shell
$>  ./ctop subscription-service <redis-url> <chain_name> <chain_rpc>
```

## Persisting Monitored Events

Monitored events are temporarily stored in Redis, and ultimately need to be pushed to postgres for ctop to function. To do so run the following command:

You must specify at minimum one `chain_name`, however you may specify multiple if you are monitoring more than one network.

```shell
$> ./ctop redis-event-stream <redis-url> <chain_name>
```