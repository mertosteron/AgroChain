# AgroChain — staged implementation plan

Status: Stages 1–8 are implemented and locally verified. The [Stage 7 report](STAGE7_REPORT.md) records tests and measurements; the [Stage 8 report](STAGE8_REPORT.md) records the final package, fresh-workspace installation and fixed demo replay. Physical validation on a second computer remains a documented limitation. Earlier dependency discussions are historical. One stage at a time.

## Stage 1 gate

This document set fixes the route, product/state semantics, organization roles, storage classifications, payload and cryptographic formats, simulator boundaries, score and demo outcomes. Documentation validation checks file/link presence, JSON examples, the commitment/signature vector, scenario arithmetic and a consistency trace. It does not prove a working Fabric deployment. The root [ARCHITECTURE.md](../ARCHITECTURE.md) is a pointer for repository instructions, not a second architecture.

## Remaining seven stages

### Stage 3 / Stage 4 acceptance boundary

The scope decision recorded on 2026-09-18 in [open questions](OPEN_QUESTIONS.md)
preserves the Stage 1 security contract: test the complete domain, deploy queries
and fail-closed mutation entry points, and defer verified live writes. The seven
core Stage 3 criteria in AGENTS.md are exercised at the domain/harness layer;
deployment, authorization rejection, actual endorsement and persistence are also
tested on Fabric. No test-only provider is deployed. A Stage 3 PASS must always be
qualified by this scope; it is not a PASS for the original expanded live-write set.

For the sequential dependency below, **Stage 3 PASS means that approved core and
deployment gate**, not the future verified product workflow. Stage 4 must implement
and pass the real populated lifecycle, replay/MVCC, business-event and populated
public/private restart checks in addition to its privacy/integrity criteria, before
Stage 5 may begin. This makes the existing test-layer accommodation explicit without
enabling an unverified workflow or discarding any final-product acceptance test.

`make stage3-check` runs the approved gate on an already deployed network. Its
positive product lifecycle evidence is explicitly unit/harness evidence; its real
commits are health probes. The completion report records both layers separately.

| Stage / dependencies | Concrete deliverables | Verification and mandatory acceptance gate |
| --- | --- | --- |
| **2 — Fabric network — PASS locally**, requires Stage 1 PASS | Fabric 2.5.15 pins/checksum; Compose; four application MSPs/peers; separate OrdererMSP with three Raft orderers; agrochannel; TLS/cryptogen development identities; stable collection definitions prepared but undeployed; profile template; Make lifecycle/verification and Arch guide | Explicit Stage 2 request supersedes the single-orderer/smoke-contract plan: no chaincode installed or committed. Bootstrap from empty generated state; verify all nodes, TLS including negative checks, four memberships/MSPs/anchors, three active consenters and exact TLS certificates, admin/client channel queries, peer restarts, shutdown/re-bootstrap, cleanup and rerun. Actual business deployment, endorsement execution and PDC behavior move to their implementation stages. See report for completed local evidence. |
| **3 — Core chaincode**, requires Stage 2 PASS | Go package; public batch/transfer/receipt state; command schemas and storage seams for evidence; owner/custody rules; fixed quantities; unique IDs/operations; errors/events; public queries | Unit/contract-harness tests for create, pickup, freight reference presence, delivery/ownership, query; wrong MSP/role/recipient; quantity mismatch; duplicate replay; invalid state/version; simulated concurrent acceptance. Real-network checks at this gate cover deployment and read/smoke behavior only. Test helpers may supply explicit test-only evidence, but no enabled network path may accept unverified source assertions. Product writes awaiting Stage 4 verification remain disabled on deploy; Stage 3 domain evidence uses verifier/PDC test doubles and is labeled accordingly. |
| **4 — Privacy and evidence**, requires Stage 3 core/deployment PASS as defined above | PDC writes/reads; exact member/endorsement policies; source-key bootstrap registry; signature and commitment verification; private record salts; original-byte integrity verifier; signed test fixtures; eliminate/disable all runtime test-verifier paths | Actual member client retrieves intended trade/freight/audit data; nonmember client and nonmember peer cannot retrieve it; shared provenance remains readable; direct chaincode invalid-signature/binding/replay rejects; altered opening/attachment fails; valid original succeeds; Go/Java reproduce vector; inspect blocks/events/logs for leakage. Verified populated product lifecycle, real duplicate/MVCC handling, business events and populated public/private restart persistence must pass here. Missing private data yields explicit unavailable, not default zero. Test fixture generators are not institutional integrations. |
| **5 — Spring Boot and four institutional simulators**, requires Stage 4 PASS | Gateway integration; authenticated per-MSP signers; CKS, EFATURA (purchase + freight), HKS, UETDS adapters and signed scenario sources; API; operation recovery; controlled attachments; public projection/checkpoint; API documentation | Whole batch follows producer → carrier → retailer through backend using signed simulator data; all four source labels remain SIMULATED; attachment/evidence and privacy queries work. Meaningful 4xx/5xx/pending errors; forged role, dependency outage, timeout-after-commit, crash/replay and duplicates tested. Price report can be stored; anomaly/UI completion remains Stage 6. |
| **6 — Explainable anomaly and UI**, requires Stage 5 PASS | Deterministic backend score and chaincode attestation check; immutable policy snapshot; pending/retry workflow; review state/actions; actor/reviewer/consumer pages and lot QR; no ML training | A jury member compares normal 40% and suspicious 85% cases against 50%, explains inputs/basis, and performs a permitted review transition. Test exactly 50%, above boundary, falling price, zero purchase and missing evidence. Consumer endpoint/UI omit every PDC-only value and review classification; logistics denied trade data. No UI-only mock of required backend capabilities. |
| **7 — Tests and measurement**, requires Stage 6 PASS | Unified test runner; clean-seed integration/E2E cases for all three scenarios; negative access tests against real peers; concurrency/recovery tests; environment and results report; basic latency observations | All competition-critical tests pass. Two clean reset/seed runs reproduce outcomes. Record hardware, OS, versions, topology, warm/cold state, payload sizes, sample count, concurrency and failures. Measure submit-to-VALID-commit, shared/private query and full route duration (e.g. 30 sequential local samples); report median/p95 with methodology. No national-capacity inference or SLA claim. |
| **8 — Competition package**, requires Stage 7 PASS | Final README/install guide; architecture figure; deterministic seed commands; 10–12 slides; three-minute demo runbook; preflight; local recorded fallback and screenshot/evidence pack; clear simulator/privacy limitations | A teammate on another supported machine follows only documented steps, starts environment, loads scenarios, shows route/normal/suspicious/tamper/replay and private/public boundaries. No live external services required during presentation once dependencies are installed. Rehearse live and labeled fallback; record unresolved environmental constraints. |

Stage 3's test-double allowance is a test-layer dependency accommodation, not permission to claim authenticated documents or deploy an insecure evidence bypass. Stage 4 remains the gate for verified product writes. Cryptogen identities lack business attributes; the user's Stage 3 MSP-only fallback is implemented and explicitly limited. Present role attributes are checked in code/tests, but real role-bearing identities must be provisioned and required before Stage 4 enables writes. The implemented command surface and exact deployed boundary are in the [chaincode guide](../chaincode/agrochain/README.md).

## Planned reproducible command surface

Stage 2 implements `make check`, `generate`, `network-up`, `channel-create`, `verify`, `network-down`, `clean-generated`, `bootstrap`, `smoke`, and `test` using `network/scripts/`; see the [network guide](../network/README.md). The following originally proposed paths are historical targets, not a second network implementation. Future stages should extend the existing Make surface instead of duplicating lifecycle scripts:

```text
scripts/check-prerequisites.sh
scripts/start-network.sh
scripts/deploy-chaincode.sh
scripts/stop-network.sh
scripts/reset-demo.sh
scripts/seed-demo.sh
scripts/test-all.sh
scripts/demo-preflight.sh
```

Reset is explicit/destructive and names only generated demo volumes/credentials. Stop preserves data. Fresh keys and salts may change across reset; business scenario values stay fixed. Record exact prerequisite downloads and offline cache preparation. Never overwrite a real wallet or unscoped filesystem path.

## Requirement-to-evidence trace

| Claim | Specification | Stage evidence |
| --- | --- | --- |
| Multi-organization ledger | Architecture topology; security identity matrix | Stage 2 live peer/channel/TLS/configuration and persistence output; business deployment/commit in Stage 3/4 |
| Valid ownership and custody | State-machine edges and invariants | Stage 3 positive, denial and race tests; Stage 5 API route |
| Confidentiality beyond frontend | PDC memberships, field table, transient routing | Stage 4 authorized/nonmember/peer inspection tests; Stage 6 public allowlist tests |
| Authentic simulator input and tamper detection | Data signature/commitment protocol and source bindings | Stage 4 crypto vectors; Stage 5 all four signed adapters; Stage 7 replay/tamper scenario |
| Explainable review signal | Product formula, policy snapshot, review machine | Stage 6 boundary/results tests and reviewer UI |
| Reproducible demo | Product scenarios and timing; scripts/install plan | Stage 7 repeat runs; Stage 8 teammate rehearsal |

## Minimum shippable path and cut line

Retain all mandatory proof points: four application organizations, three Raft orderers and one shared channel, fixed whole tomato batch, actor-signed handoffs, real PDC exclusion, signed fixtures from all four simulators, exact tamper/replay rejection, Spring Boot route, two explainable price outcomes, regulator review and consumer-safe trace. A lightweight simulator module and small server-rendered UI are sufficient. The minimal path still passes each stage; time pressure does not justify skipping privacy or replacing ledger commits with UI mocks.

Cut optional polish first: elaborate dashboards, charts, animations, streaming infrastructure, multi-host resilience, ML, complex search, cloud deployment and additional commodities/organizations. The three local orderers are now required Stage 2 topology, not optional polish. Use seeded historical route plus one live review transaction for the three-minute talk. If required tests or privacy gates fail, label the system incomplete and use an honestly labeled partial/offline demonstration; do not report a passing stage.

## Stage completion discipline

Each later stage reports implemented requirements, changed files, exact test commands and actual results, acceptance checklist, assumptions, limitations and deferred work. Add tests while implementing each stage, not only in Stage 7. Update this contract whenever a design decision changes; preserve historical input/policy meaning and fixture migration notes. No automatic advance into the next stage after a failed gate.
