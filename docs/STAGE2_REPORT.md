# AgroChain — Stage 2 completion report

## Stage

Stage 2 — Reproducible Multi-Organization Fabric Network. Verified on 2026-09-17.

## Completed

- Four TLS peers and four anchors: ProducerMSP, LogisticsMSP, RetailerMSP, RegulatorMSP.
- Three Raft orderers under a separate OrdererMSP, one `agrochannel`, persistent named
  Docker volumes, dedicated bridge networking, loopback-only host port mappings.
- Channel participation API with mutual-TLS administration, cryptogen development
  identities, pinned CLI/images, real `/healthz` checks and bounded readiness polling.
- Safe repeatable generation/bootstrap/stop/reset; four generated Gateway configuration
  profiles from one version-controlled template; stable but undeployed PDC definitions.
- Automated live verification and 20 configuration/safety regression tests; Linux CI
  workflow added. No application chaincode or later-stage application work implemented.
- Architecture and implementation plan reconciled with the explicit Stage 2 request.

## Repository inspection and resolved decisions

The initial Git working tree was clean. The repository contained Stage 1 Markdown
contracts, the root architecture pointer, AGENTS.md and `.gitignore`; no README,
Docker/Fabric implementation, tests, environment files or CI existed. The MSP names,
separate OrdererMSP, LevelDB choice, endorsement rule and collection memberships were
reused without duplicating a network implementation.

The explicit Stage 2 request superseded three Stage 1 decisions: `agrochain` became
`agrochannel`; one Raft orderer became three; disposable smoke-chaincode deployment
was removed from this gate. Stage 2 also no longer chooses future Go/Java/Gateway
dependencies. The existing domain, evidence, lifecycle and business-role contracts
were preserved. Historical Stage 1 reports remain historical evidence.

Fabric CA was not explicitly mandated. Cryptogen is the local development default;
it does not supply domain `agrochain.role` attributes. The fixed Stage 1 collection
names/memberships justify preparing `collections.json`, but no collection is deployed.

## Versions and environment

| Component | Executed version |
| --- | --- |
| OS / architecture | Arch Linux rolling, Linux 7.2.4-arch1-2, x86-64 |
| CPU / memory | Intel Core i7-12650H; MemTotal 32,553,272 KiB |
| Docker Engine | 29.8.0 |
| Docker Compose plugin | 5.5.1 |
| Fabric peer/orderer/CLI | 2.5.15, commit 83c7930 |
| Bash / Python | 5.3.15 / 3.14.7 |
| jq / OpenSSL / GNU Make | 1.8.2 / 3.6.4 / 4.4.1 |

Exact image tags: `hyperledger/fabric-peer:2.5.15` and
`hyperledger/fabric-orderer:2.5.15`. Canonical version and binary checksum pins live
in [network/.env.example](../network/.env.example). Resolved pull digests observed:

```text
fabric-peer:    sha256:c48323777139fcabe77a98cf4eeaf4466827dec225f754064f52f22395a8bd89
fabric-orderer: sha256:72697d099f45b8e57f86d9a69fbedb683cd5888bc7b28ec01695773d0433b791
Linux amd64 CLI archive SHA256:
c0f58626ff73b15bfc9ab0cf2b78a7e60dda3c53051a1032f32ffb3111df677f
```

The official [2.5.15 release](https://github.com/hyperledger/fabric/releases/tag/v2.5.15)
includes the Docker Engine 29 compatibility fix. The downloaded CLI archive checksum
matched official release metadata before extraction into ignored `network/tools/`.
No system package installation, daemon reconfiguration or sudo command was performed.
Docker/network access needed the execution environment's sandbox approval.

## Changed Files

Modified:

- `.gitignore`
- `docs/ARCHITECTURE.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/OPEN_QUESTIONS.md`

Created:

- `.github/workflows/stage2-network.yml`
- `Makefile`
- `README.md`
- `docs/STAGE2_REPORT.md`
- `network/.env.example`
- `network/README.md`
- `network/channel-artifacts/.gitkeep`
- `network/compose/compose-agrochain.yaml`
- `network/config/collections.json`
- `network/config/configtx.yaml`
- `network/config/crypto-config.yaml`
- `network/connection-profiles/organization.json.template`
- `network/organizations/.gitkeep`
- `network/scripts/clean-generated.sh`
- `network/scripts/common.sh`
- `network/scripts/create-channel.sh`
- `network/scripts/down.sh`
- `network/scripts/generate.sh`
- `network/scripts/network.py`
- `network/scripts/prerequisites.sh`
- `network/scripts/smoke.sh`
- `network/scripts/up.sh`
- `network/scripts/verify.sh`
- `network/tests/test_network.py`

Local generated credentials, certificates, blocks, profiles, CLI tools and runtime
evidence are intentionally ignored and are not source deliverables. Final inspection
checked 192 generated/runtime/tool files: all were ignored. No files were staged.

## Commands

After the explicit prerequisite installation in [network/README.md](../network/README.md):

```bash
make check
make bootstrap       # Start/create/join; existing valid state is preserved
make verify          # Includes four sequential peer restarts
make smoke           # Stop, bootstrap/verify, stop, bootstrap/verify
make network-down    # Preserve ledger volumes and credentials
make clean-generated # Explicitly erase only named demo volumes/generated artifacts
make test            # Configuration and safety tests; no live network needed
```

Individual `make generate`, `make network-up` and `make channel-create` are available.
`make -C /path/to/AgroChain <target>` works from another directory.

## Tests Executed

- `sha256sum /tmp/agrochain-fabric-2.5.15.tar.gz` — PASS; matched pinned official digest.
- `make check && make generate && make network-up && make channel-create` — initial
  network-up FAILED because the peer's default Docker builder registered a health
  dependency on an unmounted Docker socket. Fixed by deriving a container-local
  core.yaml with its Docker endpoint disabled, retaining the image's built-in
  external builder. An empty environment override was tested and did not fix it.
- `make network-up && make channel-create && make verify` — PASS after that fix.
- `make network-down && make clean-generated && make clean-generated` — PASS;
  seven exact volumes removed, source/tools preserved, repeated cleanup safe.
- Fresh-state `make smoke` exposed two preflight failures: absent Docker networks
  use `network agrochain-fabric not found`, and a recently closed TLS endpoint left
  port 7053 in TIME_WAIT. Both fixed; regression tests distinguish these from daemon
  permission failures and active port listeners. No TLS or health checks were weakened.
- Final `make test` — PASS, **20 tests**; includes shell syntax and rendered Compose,
  cleanup symlink/foreign-volume/preservation checks, absent-network/error handling,
  active-port/TIME_WAIT checks, malformed channel MSP/anchor/consenter/certificate
  rejection and profile/collection configuration tests.
- Final `make smoke > /tmp/agrochain-stage2-final-smoke.log 2>&1` — PASS, exit 0.
  Both full bootstrap/verification cycles passed, with a shutdown between them.
  Each cycle restarted all four peers and verified identical ledger state afterward.
- From `/tmp`: `make -C /home/mert/Projects/AgroChain check` and
  `make -C /home/mert/Projects/AgroChain generate` — PASS; existing manifest validated,
  identities unchanged, live network's own occupied ports accepted.
- `git diff --check`, `git status --short --untracked-files=all`,
  `git diff --cached --name-only`, generated-file `git check-ignore --stdin`, and
  source private-key scan — PASS; no generated secrets exposed or staged changes.
- GitHub-hosted CI — **NOT RUN** in this session. The workflow performs tests, two
  bootstraps and a fresh cleanup/bootstrap on Ubuntu 24.04; local results do not
  assert hosted-run success or a second-machine rehearsal.

The final ignored log is `network/runtime/stage2-smoke.log`. Both cycles verified
all nodes, four TLS peers, three TLS orderers, three mutual-TLS admin endpoints,
wrong-CA rejection and missing-client-certificate rejection. Each organization's
Admin and User1 successfully queried its channel. All installed/committed application
chaincode lists were empty. Four decoded live peer configurations had the required
MSPs, roots, NodeOUs, anchors and exact Raft TLS certificates.

All peers reported height **1** and the same genesis block hash across both final cycles:
`N1FVfy6+bs95EhTa93t08xdRI09OvpLWm+zX++BWHbE=`. A destructive regeneration changes this
hash because new random development certificates are embedded. This is infrastructure
evidence; no business transaction or chaincode execution was claimed.

## Acceptance Criteria

- [x] PASS — Clean source checkout can discover prerequisites through the guide and `make check`.
- [x] PASS — Required tools, supported versions, pins/checksum, Arch setup and ports documented.
- [x] PASS — Four-organization, three-orderer network starts from empty generated state.
- [x] PASS — TLS verified on all peer/orderer endpoints; admin mutual TLS and negative checks pass.
- [x] PASS — `agrochannel` created; all three ordering nodes report active consenter status.
- [x] PASS — Every peer joins and queries the channel using its own organization identities.
- [x] PASS — Four application MSPs/roots/NodeOUs and anchors verified from live channel configuration.
- [x] PASS — Three Raft consenters and their TLS certificates verified from live configuration.
- [x] PASS — Peer restarts, shutdown, second bootstrap, clean reset and repeated cleanup tested.
- [x] PASS — Generated credentials/blocks/runtime files ignored; no private material staged.
- [x] PASS — Versioned profile template generates four organization profiles with relative CA references.
- [x] PASS — Automated verification exits successfully; 20 local automated tests pass.
- [x] PASS — No business chaincode installed/committed; no later-stage application implementation.
- [x] PASS — Stage 1 topology assumptions and stage gate revised consistently with this request.

## Assumptions

- A single trusted local operator controls the ordering MSP and demo host; separate
  MSPs demonstrate logical organization separation only.
- Linux x86-64 with standard Docker Engine and an accessible Docker socket for the
  **host lifecycle commands**; no Docker socket is mounted into a peer.
- Cryptogen development identity generation is sufficient for this network-only
  stage. No real institutional source, user data or production credential is used.
- Lifecycle commands execute sequentially; concurrent administrative invocations
  against the same named network are unsupported.

## Known Limitations

- No independent-host resilience, Byzantine fault tolerance, performance benchmark
  or production PKI/revocation operations. A three-node quorum was configured and
  verified; orderer fault-injection testing was not performed.
- No actual Gateway Java client, business endorsement transaction, private-data
  read/write or private-data recovery test. Collection configuration alone proves no privacy.
- Cryptogen User1/Admin identities lack `agrochain.role`; real domain authorization
  tests require correctly issued attributes in the next relevant stage.
- Images use explicit release tags; observed resolved digests are recorded, but
  Compose does not enforce digest immutability. Cache verified artifacts for offline use.
- CLI checks restrict the supported installation to Linux amd64. Broader platforms,
  lower-bound Docker versions and Ubuntu-hosted CI were not executed locally.
- No unresolved local Stage 2 acceptance blocker. Hardware baseline remains an
  assumption, not a measured minimum; all nodes still share this one machine.

## Deferred to Future Stages

Stage 3 may rely on repeatable TLS networking, the four application MSP names,
`agrochannel`, channel governance/default endorsement configuration, peer endpoints,
LevelDB persistence, three Raft consenters, anchors, development admin/client identities,
and the connection-profile/collection configuration sources.

Stage 3 must implement and verify Go chaincode packaging, execution/deployment,
role-bearing test identities, domain authorization, lifecycle/quantity/idempotency
rules and transaction commit evidence. Do not infer these capabilities from Stage 2.
The bundled external builder is present; chaincode execution is not yet configured
or demonstrated. Extend the existing network scripts rather than adding a second network.

Stage 4 must prove PDC access and document integrity. Stage 5 supplies application
signers and Gateway integration/simulators. Anomaly/UI, measurements and competition
materials remain Stages 6–8. Work stops at Stage 2 here.

## Result

**PASS — Stage 2 acceptance criteria verified locally.** Seven nodes remain running
after the final smoke test. No source files are staged or committed. This is a local
competition network result, not a production, privacy or business-workflow certification.
