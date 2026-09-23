# AgroChain

## Yarışma paketini açın

[Aşama 8 teslim paketi](docs/competition/README.md): kurulum rehberi, sabit iki
demo senaryosu, mimari, 12 slaytlık PPTX, üç dakikalık akış ve bağımsız çevrimdışı
HTML. [Kabul raporu](docs/STAGE8_REPORT.md) temiz kurulum provasını ve sınırlarını
belgeler. Kurum verileri **SIMULATED**, yerel blockchain işlemleri gerçektir.

Desteklenen Linux ortamında Fabric araçları hazırlandıktan sonra yeni kurulum:

```bash
make demo-setup
make backend-run
# İkinci terminalde:
make demo-seed demo-check
```

Mevcut ağ için önce `make network-up chaincode-check demo-ready` kullanın.
Arayüz `http://localhost:8080`; normal parti `BAT-DEMONORMAL01`, şüpheli parti
`BAT-DEMOSUSPICIOUS01`. Ayrıntılar ve kimlik bilgileri
[kurulum rehberinde](docs/competition/INSTALLATION.md).
`make competition-package` kaynak ve sunumları `dist/AgroChain-competition.tar.gz`
içine, checksum ve dosya manifestiyle paketler; yerel anahtarlar, tokenlar ve
ledger diskleri dahil edilmez. İlk bağımlılık kurulumu internet gerektirir.

TEKNOFEST agricultural traceability pilot using Hyperledger Fabric. Stage 1 contracts
live in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) and its companion documents.
Stage 2 provisions four organization peers, three Raft orderers and TLS-enabled
`agrochannel`. Stage 3 adds a tested Go lifecycle domain and deployed public queries.
Stage 4 enables verified business writes with signed source evidence, mandatory
certificate roles and three Private Data Collections.
Stage 5 adds the Spring Boot API, Java Fabric Gateway, signed institutional
simulators, durable operation recovery and a consumer-safe public read model.
Stage 6 adds deterministic price analysis attested by chaincode, private review
actions, Turkish actor/inspector screens and a consumer lot QR page. Institutional
data remains simulated; no real government integration is claimed.

Stage 7 provides `make stage7-check`: one sequential test gate covering real PDC
access, backend/browser workflows, pilot latency measurements and two fresh
network reproductions with the existing ledger preserved. See the
[testing guide](docs/STAGE7_TESTING.md) for prerequisites, temporary network
interruption, measurement boundaries and evidence locations.
The [Stage 7 completion report](docs/STAGE7_REPORT.md) records passing component
tests, two successful clean reproductions and the installation defects fixed
during validation. Stage 8 adds the competition package linked above without
changing the domain or deploying the pilot publicly.

Stage 6 operation and acceptance commands are in [the UI guide](docs/STAGE6_UI.md).
Start `make backend-run` and open `http://localhost:8080`. Existing 0.2.1 networks
need an explicit chaincode upgrade; keep the ledger and source keys:

```bash
make network-up
make chaincode-upgrade CHAINCODE_VERSION=0.3.0 CHAINCODE_SEQUENCE=6
make stage6-check
```

Sequence 6 is this development network's next sequence; on another existing
network select its actual next sequence. A fresh channel uses sequence 1.

Backend setup, API payloads, credentials and limits are in the
[backend guide](backend/README.md); executed acceptance is in the
[Stage 5 report](docs/STAGE5_REPORT.md). On an already bootstrapped Stage 4 network:

```bash
make backend-prerequisites
make network-up backend-prepare
make stage5-check
make backend-run
```

The API listens on loopback port 8080. Its source responses are explicitly
SIMULATED; its blockchain operations use the actual local Fabric network.

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

After Stage 2 setup, build and deploy the verified contract (Python >=3.10 with
`cryptography` and Java >=17 are required for the privacy acceptance tools):

```bash
make chaincode-prerequisites
make chaincode-deps
make chaincode-test
make chaincode-deploy
make privacy-prepare
make stage4-check
```

The [chaincode guide](chaincode/agrochain/README.md) documents schemas, authorization,
exact pins and lifecycle/upgrade commands. The [Stage 3 report](docs/STAGE3_REPORT.md)
records the historical core-domain gate. The [Stage 4 report](docs/STAGE4_REPORT.md)
records real populated lifecycle, privacy, integrity and persistence acceptance.
Test evidence providers are never available in deployment. Source fixtures are
explicitly SIMULATED; purchase, freight and retail values stay private.

Run `make stage4-check` on an already deployed network after `make privacy-prepare`.
It creates synthetic batches, inspects real blocks/PDCs and restarts the local
network without deleting data. It bootstraps immutable simulator trust once using
the Regulator admin; later runs verify the existing keys. Keep ignored runtime keys
with the ledger. This remains a local test tool; the Stage 5 backend is documented above.

Use Python >=3.10 for network commands. If a shell environment selects an older
Python, activate a supported interpreter first; on this Arch host the explicit
working invocation is `PATH=/usr/bin:/bin:$PATH make stage4-check JAVA=/usr/lib/jvm/java-26-openjdk/bin/java`.
