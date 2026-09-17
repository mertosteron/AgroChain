# AgroChain

TEKNOFEST agricultural traceability pilot using Hyperledger Fabric. Stage 1 contracts
live in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and its companion documents.
Stage 2 provisions four organization peers, three Raft orderers and TLS-enabled
`agrochannel`. No business chaincode, backend, government integration or UI exists yet.

Install the pinned prerequisites in the [network guide](network/README.md), then:

```bash
make check
make bootstrap
make verify
make smoke
make network-down
```

`make clean-generated` explicitly deletes only generated demo credentials/artifacts
and the seven named demo ledger volumes after shutdown. `make test` checks configuration
and cleanup safety. See [Stage 2 results](docs/STAGE2_REPORT.md) for evidence, acceptance
status, assumptions and the Stage 3 handoff. This is a local competition test network,
not a production or nationwide deployment.
