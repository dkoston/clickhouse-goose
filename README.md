# clickhouse-goose

[![Build Badge](https://travis-ci.org/dkoston/clickhouse-goose.svg?branch=master)](https://travis-ci.org/dkoston/clickhouse-goose)

Clickhouse allows for replicated tables but all data isn't fully replicated. As
such, migrations for schemas are run against each clickhouse server.

To do this, we parse the connection string and run goose as many times as needed.


## Example

```
clickhouse-goose tcp://clickhouse1:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=clickhouse2:9000,clickhouse3:9000
```

The above will run goose against clickhouse1, clickhouse2, and clickhouse3 using ./db/dbconf.yml

## Development

### Pre-requisites

- go 1.11.2+
- docker

### Releases

Releases can be built (via docker) with `./release.sh`