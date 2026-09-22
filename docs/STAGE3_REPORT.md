# AgroChain — Stage 3 completion report

## Stage

**Stage 3 — Core chaincode and fail-closed deployment.**
Closure verified on **2026-09-21**, using the scope decision already recorded in
[OPEN_QUESTIONS.md](OPEN_QUESTIONS.md) on 2026-09-18.

**PASS for the approved core-domain + fail-closed deployment scope.**
This is explicitly **not** a PASS for a verified live product lifecycle. Business
writes still return `EVIDENCE_VERIFICATION_UNAVAILABLE`; that mandatory final-product
capability depends on Stage 4 evidence verification and PDC storage.

The former inspection-only report was stale: source, deployment and tests existed,
but the report still said implementation had not begun and a scope answer was
pending. This report replaces that obsolete current-state description. The
[2026-09-20 review](PROJECT_REVIEW_ROADMAP_TR.md) preserves the historical finding.

## Completed

- Re-inspected AGENTS.md, the architecture and companion contracts, source, tests,
  deployment scripts and existing local changes before editing.
- Preserved the existing Go domain and public schemas. Seven commands cover batch
  creation, two-step pickup/delivery, freight reference and private-price report
  reference. Ownership remains Producer during transport; Logistics has custody
  only; Retailer acceptance changes both ownership and custody.
- Retained integer grams/kuruş, bounded quantities, explicit state/version checks,
  MSP/role authorization, unique IDs, duplicate rejection, append-only receipts,
  shared queries and deterministic ordered writes.
- Added a **56-case command/checkpoint matrix** covering all seven commands at all
  eight checkpoints, including transport before and after freight. Invalid cases
  assert no writes/events; successful cases check version, event and receipt.
- Added late-freight recovery: delivery may be offered first, acceptance fails
  without freight, the failed operation ID remains unused, and the same acceptance
  succeeds after freight is recorded. Populated shared queries then check all four
  organizations, owner/state index cleanup and ordered seven-step history.
- Added **28 X.509 role/command cases** through the real Fabric client-identity
  parser in the test harness, plus malformed certificate attribute rejection.
  Correct roles still cannot enable writes; wrong roles cannot acquire authority.
  Test certificate generation does not provision live business identities.
- Added `make stage3-check`: configuration/package tests, Go checks/build,
  live network/definition verification, six integration tests and a full
  data-preserving restart. No destructive reset is part of this target.
- Added a Stage 3 GitHub workflow triggered by chaincode, network, Makefile and
  workflow changes. It provisions pinned tools, bootstraps/deploys, runs the same
  gate and checks repeat deployment. Hosted execution is not claimed.
- Reconciled the Stage 3/4 dependency in the implementation plan: the approved
  core/deployment gate is Stage 4's prerequisite; populated verified lifecycle,
  privacy, real business MVCC/events and populated restart evidence remain
  mandatory Stage 4 exit checks.
- Re-deployed the unchanged production binary against the existing definition,
  verified all four organizations and preserved the existing ledger and TLS pairs.

## Changed Files

Changes made in this closure:

- `chaincode/agrochain/internal/agrochain/stage3_test.go` — command/checkpoint
  matrix, recovery, populated queries and index/history assertions.
- `chaincode/agrochain/internal/agrochain/contract_test.go` — certificate
  attribute fixtures and role enforcement through the Fabric identity API.
- `Makefile` — the combined Stage 3 verification target.
- `.github/workflows/stage3-chaincode.yml` — fresh-runner deployment and gate.
- `README.md` — gate instructions, side effects and Python prerequisite.
- `chaincode/agrochain/README.md` — test layers, gate and CI scope.
- `docs/IMPLEMENTATION_PLAN.md` — explicit acceptance-layer/dependency mapping.
- `docs/OPEN_QUESTIONS.md` — closure of the stale report/decision ambiguity.
- `docs/PROJECT_REVIEW_ROADMAP_TR.md` — dated update above the historical review.
- `docs/STAGE3_REPORT.md` — this current completion report.

The implementation already present at the start was retained: domain/contract/
ledger/query/validation/model sources, prior tests, pinned module files, Dockerfile,
four CCAAS services, collection configuration, packaging/client/toolchain/deployment
scripts and Stage 2 adaptations. No production-domain behavior, dependency version,
release version, MSP policy or private-data classification changed in this closure.

Test/build/runtime outputs were refreshed under `chaincode/agrochain/build` and
`network/runtime`; existing generated credentials were retained. Existing user
changes were not discarded or committed. The chaincode source was untracked when
this task began; no Git commit/push or remote-checkout completeness is claimed.

## Tests Executed

| Command / check | Actual result | Meaning |
| --- | --- | --- |
| `make chaincode-format` | PASS | New Go tests formatted; production behavior unchanged. |
| `make chaincode-test` | PASS | Formatting, module verification, vet, race-enabled uncached tests, coverage and static Linux build. |
| `PATH=/usr/bin:/bin:$PATH make network-up chaincode-deploy stage3-check` | PASS, exit 0 | Current network started, matching release re-deployed, and the complete combined gate executed. |
| `make test`, invoked by the gate | PASS, 24 tests | Network/configuration, cleanup safety, packaging and disabled-production-provider checks. |
| `make verify`, invoked by the gate | PASS | Seven nodes, TLS/mTLS and negative TLS checks, four MSPs/anchors, three Raft consenters, organization queries and peer restarts. |
| `make chaincode-check`, invoked by the gate/restart | PASS | Committed definition/approvals and fail-closed health for all four organizations. |
| `make chaincode-integration`, invoked by the gate | PASS, 6 tests | Real queries, denied writes, wrong organization/input rejection, valid two-endorser health commits and invalid one-endorser commit. |
| `make chaincode-restart-check`, invoked by the gate | PASS | Full shutdown/start preserves four ledger tips, contract health and absent test-batch state. |
| Workflow YAML and embedded shell checks | PASS | System PyYAML parsed the workflow; trigger/permission checks and `bash -n` passed. Does not execute GitHub Actions. |
| Stage 1 documented Node contract checker | PASS | 9 documents, 9 JSON examples, 44 local links, crypto/tamper vectors and 7 scoring examples. |
| Changed-document link/format checks and `git diff --check` | PASS | Local references resolve and whitespace checks pass. |

Go statement coverage: **89.7%** for `internal/agrochain`, **87.8%** overall.
The server bootstrap remains 0% unit coverage; real CCAAS health is exercised by
live tests. Coverage is not a substitute for PDC or end-to-end security evidence.

An initial sandboxed live command failed with Docker access denied, before running
deployment. Retrying with approved Docker access succeeded. Network commands used
the system Python 3.14.7 because the shell's default Python 3.9.23 does not satisfy
the documented >=3.10 prerequisite. No checks were weakened.

### Observed deployment and persistence

- Fabric **2.5.15**, Docker **29.8.0**, Compose **5.5.1**, Go **1.27.1**.
- Four application peers, three Raft orderers, four TLS CCAAS services,
  channel `agrochannel`.
- Release **0.1.2**; existing matching definition retained, no new upgrade requested.
- Production binary SHA-256:
  `fdb796230eafbf04e04e32152abde0c81d4b63bd63038cef428908414f07e58a`.
- Four valid health transactions recorded in blocks **32–35**.
- One intentionally invalid single-endorser transaction is recorded but rejected
  by peer validation with `ENDORSEMENT_POLICY_FAILURE`.
- Final ledger height **37** on each peer; before/after restart snapshots equal.
- Health reports `writesEnabled=false` and
  `evidenceVerification=UNAVAILABLE_STAGE_3`.
- The test batch remains absent. Health commits do not create business records,
  private writes or business events.

Generated local evidence: `network/runtime/chaincode/integration-commits.json`,
`network/runtime/chaincode/restart-evidence.json`,
`network/runtime/chaincode/manifest.json` and
`chaincode/agrochain/build/coverage.out`. These are ignored machine-specific
artifacts; this report records their meaning rather than committing private material.

### Reproduce

Install the pinned prerequisites described in the network/chaincode guides,
select Python >=3.10, then run:

```bash
make chaincode-prerequisites
make chaincode-deps
make bootstrap
make chaincode-deploy
make stage3-check
```

For an already deployed network only `make stage3-check` is needed.
It restarts nodes and appends health/negative-endorsement test transactions.
`make network-down` stops without deleting identities/ledgers.
`make clean-generated` is a separate explicit destructive reset and was not run.

Fresh-runner provisioning is configured in CI, but fresh empty-state bootstrap and
hosted CI were **not executed in this closure**. Local repeat deployment and
data-preserving restarts were executed.

## Acceptance Criteria

The layer column is part of each assertion; it must not be omitted when presenting
these results. This is the previously approved Stage 3 test-layer accommodation,
not permission to deploy unauthenticated evidence.

| Core criterion | Layer / verification | Status |
| --- | --- | --- |
| Valid batch creation succeeds | Test-only evidence, domain state/receipt/event checked | PASS |
| Valid ownership transfer succeeds | Whole lifecycle and late-freight route; Retailer acceptance changes owner/custodian | PASS |
| Unauthorized transfer fails | MSP/recipient/role tests with no mutation; live wrong-organization proposals rejected | PASS |
| Duplicate operation fails without duplicate records | Domain replay tests and caller-scoped receipt recovery | PASS |
| Invalid quantity fails | Schema/bound/whole-batch domain tests; live malformed/negative creation rejected | PASS |
| Invalid lifecycle transition fails | 56 command/checkpoint cases, no writes/events on rejection | PASS |
| Resulting state is queryable | Populated shared domain queries/history/index tests; deployed real queries over current empty business state | PASS |

Additional approved deployment/closure criteria:

- [x] Build, formatting, module verification, static analysis and race tests pass.
- [x] All seven command entry points and bounded shared query APIs exist.
- [x] Test evidence cannot be enabled in the deployed dispatcher.
- [x] All four identities query the deployed chaincode; well-formed writes fail closed.
- [x] Retailer + Regulator endorsements commit; Regulator-only endorsement fails.
- [x] Health blocks are inspected for validation flags, signer identities and no
  domain/private writes or business events.
- [x] Definition and current ledger state survive full shutdown/start.
- [x] Automated gate and Stage 3 CI configuration exist.
- [x] Current documentation describes implementation, evidence layers and limitations.
- [x] No Stage 4–8 feature was implemented or claimed.

## Assumptions

- The scope decision recorded on 2026-09-18 remains in effect: keep the Stage 1
  trust/privacy contract, verify the core in tests, and deploy fail-closed writes.
- Synthetic whole batches, fixed Producer → Logistics → Retailer route, one MSP per
  role and the existing two-peer endorsement policy remain authoritative.
- Existing local identities, TLS material and ledger volumes must be preserved.
- A local working tree is the delivery here; Git publication is separate.

## Known Limitations

- **Live business writes remain disabled.** No product batch was successfully
  created/transferred on Fabric. The expanded live-workflow acceptance is not met.
- Positive domain tests use synthetic verified outcomes; they do not verify source
  signatures, original attachments, nonce consumption or actual PDC operations.
- Cryptogen live identities lack business-role attributes. Stage 3's documented
  missing-role MSP fallback remains. Test certificates prove parser/enforcement
  behavior, not real CA issuance or mandatory live role provisioning.
- Concurrent business acceptance and transaction atomicity have model-level tests;
  real populated Fabric MVCC/commit behavior remains to be exercised.
- Current restart evidence covers the definition, ledger tip and empty business
  state, not persisted populated business/PDC records.
- One host controls all organizations. Server TLS is enabled for CCAAS; mutual
  peer-to-chaincode client authentication is not configured.
- No hosted CI, second-machine rehearsal, fresh empty-state deployment in this
  closure, backend, UI, performance or anomaly claim.
- Existing sources remain local/uncommitted. A remote clone is not asserted to
  contain these changes.

## Deferred to Future Stages

**Stage 4 — mandatory before enabling writes and before Stage 5:**

- Provision/require business roles; bootstrap the trusted source-key registry.
- Implement signature, canonical commitment, source/body/batch/operation/nonce
  verification and atomic private/public writes.
- Route private transients only to eligible peers and prove member/nonmember access.
- Reproduce crypto vectors in Go/Java; original documents pass, tampered ones fail.
- Replace the closed evidence seam only with the real verifier/storage path.
- Run the populated real lifecycle, duplicate/source replay, MVCC, business-event
  leakage and populated public/private restart tests.
- Keep prices/salts/openings private; never use a test-verifier runtime switch.

Stages 5–8 retain backend/simulators, anomaly/UI, full E2E measurement and the
competition package respectively. No automatic implementation of Stage 4 follows
this report.

## Result

**PASS — Stage 3 approved core-domain and fail-closed deployment scope.**

**FAIL / NOT COMPLETE — the original expanded verified live-business workflow.**
Those mandatory product checks remain explicitly assigned to Stage 4; neither
this result nor the passing health transactions certify them.

The next implementable stage is Stage 4 under the dependency mapping in
[IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md). The final competition product
remains incomplete.
