# “Word of Wisdom” TCP-server
Status: ![](https://github.com/kkucherenkov/faraway-challenge/workflows/main/badge.svg)

## Task description
Design and implement "Word of Wisdom" TCP server.
- TCP server should be protected from DDOS attacks with the [Prof of Work](https://en.wikipedia.org/wiki/Proof_of_work), and the challenge-response protocol should be used.
- The choice of the POW algorithm should be explained.
- After Prof Of Work verification, the server should send one of the quotes from the "Word of Wisdom" book or any other collection of the quotes.
- Docker file should be provided both for the server and for the client that solves the POW challenge.

## The Proof Of Work algorithm
In this project implemented the Hashcash algorithm. It is a classic, well-documented algorithm that has several advantages:
- well documented
- simple validation
- a client can get all required data via one request
- a wide range of challenge complexity settings
In this project, some changes to the hashcash protocol implemented:
- replaced a random number with a UUID-string in the `rand` field of the challenge data. It should prevent the usage of precomputed challenges
- added issued challenges cache to the service  to check that the client solved the real challenge.
    - in this implementation used in-memory local cache, but in production can be used, for example, a Redis cache (Also implemented). Redis can improve the scalability of the service, especially with the async client

## Project structure
According to the [Go-Layout](https://github.com/golang-standards/project-layout):
- `cmd/client` – client entry point
- `cmd/service` – service entry point
- `config/config.yaml` – config file
- `data/quotes.txt` – data file, contains quotes from the Book of Wisdom
- `docker` – Docker files for service and client
- `internal/client` – client protocol implementation
- `internal/pkg/cache` – cache for requested challenges
- `internal/pkg/clock` – interface for testing purposes
- `internal/pkg/config` – project configuration. It reads from the config file and env variables. Environment variables have priority and can override settings.
- `internal/pkg/pow` – hashcash implementation
- `internal/pkg/storage` – in-memory storage for quotes
- `internal/pkg/transport` – client-service protocol definition
- `internal/service` – service protocol implementation

## Prerequisites
+ [Go 1.20+](https://go.dev/dl/)
+ [Docker](https://docs.docker.com/engine/install/)

## How to run?

```bash
make install        # to get dependencies
make test           # to run unit tests
make run-service    # to start service
make run-client     # to start the client
make run-prod       # to start both client and server in
                    # docker containers
```
