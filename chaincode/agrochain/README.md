# AgroChain chaincode — Stage 6

Release 0.3.0 adds EvaluatePrice (Regulator/oracle), OpenReview and ResolveReview
(Regulator/reviewer), GetAnomaly and GetReviewHistory (batch ID, private retail audit
readers). Each accepted command increments batch version; evaluation/review leave
RETAIL_REPORTED unchanged. Scores, classifications and explanations remain in PDC;
public references/events contain opaque identifiers only. Integer boundary and
private review tests are in `internal/agrochain/analysis_test.go`. See the
[Stage 6 guide](../../docs/STAGE6_UI.md) for complete runtime behavior.

Implemented in Go using the official Fabric shim. Business commands require trusted
signed evidence and mandatory certificate roles. Before immutable trust bootstrap,
they fail with CONFIGURATION_REQUIRED. There is no runtime test-verifier bypass.
See the [Stage 4 report](../../docs/STAGE4_REPORT.md) for actual live acceptance and
the [privacy runbook](../../docs/STAGE4_PRIVACY.md) for setup and APIs.

## Contract functions and schemas

Canonical schemas and transitions remain in
[DATA_CONTRACTS.md](../../docs/DATA_CONTRACTS.md) and
[STATE_MACHINE.md](../../docs/STATE_MACHINE.md). All monetary values, including
retail prices, are private. Public arguments must never contain commercial values.

Each mutation takes exactly one JSON string containing
`schemaVersion=agrochain.command.v1`, `operationId`, `batchId`, `expectedVersion`,
`command` (matching the function), and `payload`. Unknown/missing fields are rejected.

| Function | Public payload | Caller | Verified domain behavior |
| --- | --- | --- | --- |
| CreateBatch | productCode, gradeCode, quantityGrams, originRegionCode, harvestDate, logisticsMsp, intendedRetailerMsp | Producer | Create whole tomato batch after verified eligibility |
| OfferPickup | transferId, quantityGrams | Producer owner/custodian | CREATED → PICKUP_PENDING; unique transfer; require purchase/HKS/manifest |
| AcceptPickup | transferId, quantityGrams | Named Logistics | PICKUP_PENDING → IN_TRANSPORT; change custody only |
| RecordFreightCost | costId | Logistics custodian | IN_TRANSPORT or DELIVERY_PENDING; unique verified freight reference |
| OfferDelivery | transferId, quantityGrams | Logistics custodian | IN_TRANSPORT → DELIVERY_PENDING; owner remains Producer |
| AcceptDelivery | transferId, quantityGrams | Named Retailer | DELIVERY_PENDING → RECEIVED; require freight; atomically change owner and custodian |
| ReportRetailPrice | reportId, retailLotId, policyId | Retailer owner/custodian | RECEIVED → RETAIL_REPORTED; unique lot/report; private monetary validation via internal evidence seam |

No cancellation, expiry, split, merge, shrinkage, conversion, partial transfer,
evaluation or review function exists. Stage 1 integrates production evidence into
creation/pickup and transport evidence into pickup/freight; incompatible
`RecordProductionCommitment`/`RecordTransportCommitment` aliases are not invented.
`RETAIL_REPORTED` denotes an offered price declaration, not a completed sale.

Successful domain execution returns `agrochain.receipt.v1`: operationId, batchId,
command, actorMsp, Fabric txId/txTime, resultVersion, sorted unique objectIds.
Public Batch, Transfer, EvidenceRefs and Lot structs in `internal/agrochain/models.go`
are DTO allowlists matching Stage 1; none has a price, opening, salt or personal field.

## Shared queries

Recognized channel-member MSPs may query shared data. This is not a consumer QR API.

| Function | Arguments | Result |
| --- | --- | --- |
| Health | None | `agrochain.health.v1`: build version, writesEnabled reflects bootstrap, evidenceVerification=REQUIRED_STAGE_4 |
| GetBatch | BAT ID | Batch or BATCH_NOT_FOUND |
| BatchExists | BAT ID | JSON boolean |
| GetTransfer | TRF ID | Transfer or TRANSFER_NOT_FOUND |
| GetOperation | OP ID | Receipt within caller MSP namespace, or OPERATION_NOT_FOUND |
| GetBatchHistory | Page JSON; filter=batchId | Append-only receipts in version order |
| QueryBatchesByOwner | Page JSON; filter=owner MSP | Batches ordered by batch ID |
| QueryBatchesByState | Page JSON; filter=state | Batches ordered by batch ID |

Page request: `{"schemaVersion":"agrochain.page-request.v1","filter":"ProducerMSP","pageSize":10,"bookmark":""}`.
Response: `{"schemaVersion":"agrochain.page.v1","records":[],"bookmark":""}`.
Size is 1–100; reuse the returned opaque bookmark until empty. Bookmarks bind the
object type/filter; LevelDB start keys outside that prefix fail. An extra empty final
page is possible. Pagination is not a frozen snapshot during concurrent commits.
No CouchDB syntax or full-world-state scan is exposed. Requests are at most 64 KiB,
depth 8, 32 members per object and 512 characters per string/bookmark; individual
field grammars impose tighter limits. All permitted public input strings are ASCII.

## Authorization and endorsement

Caller MSP and mandatory `agrochain.role` come from official `cid` APIs. Roles must
match producer/carrier/retailer and their MSP; Regulator readers may have
reviewer/auditor/oracle/public-reader. Admin grants only Regulator Bootstrap, not
domain rights. Attribute-less cryptogen User1 is limited to Health. `privacy-prepare`
issues development role certificates using the existing development organization
CAs. These are not production enrollment/revocation infrastructure.

The definition uses `AND('RetailerMSP.peer','RegulatorMSP.peer')`, preserving Stage 1
and its future common-PDC-member path. All four admins approve the same definition;
lifecycle commit targets all peers. Business clients retain their own identities.
A live test orders a Regulator-only endorsed health transaction and asserts
`ENDORSEMENT_POLICY_FAILURE` at commit.

Client authorization controls submissions; peer endorsement certifies execution;
Raft orders transactions; peer validation checks endorsements/MVCC before atomic
commit. Separate sender/recipient submissions express handoff consent. This fixed
policy does **not** require both handoff organizations' peers. No state-based
endorsement is introduced; a later governance change would need explicit policy/PDC
compatibility tests. MSP checks do not replace endorsements. Either required peer
can stall writes; the shared host/operator remains a trust boundary.

## State, replay and events

Composite keys preserve Stage 1 `batch`, `transfer`, `operation`, `history`
(ten-digit version), `batchEvidence`, `lot`. Additional bounded-query/reference
indexes are `owner`, `state`, `pendingTransfer`, `documentReference`, `costReference`,
`reportReference`; their values contain only opaque IDs. No history DB is needed.
All primary domain records are versioned; internal index value formats are documented
in the data-contract Stage 3 addendum.

Every mutation increments batch version, including freight without a state change.
Operations are unique by MSP + operationId. Identical and changed-payload replays
both return `DUPLICATE_TRANSACTION`; `GetOperation` recovers the original receipt.
No unsalted private-request digest is published. Document/transfer/cost/report/lot
IDs cannot be reused. A sorted write plan validates and serializes everything before
stub writes. Fabric supplies final atomicity/MVCC; failed simulations must not commit.

One `agrochain.event.v1` event accompanies each committed valid business command:
`BatchCreated`, `PickupOffered`, `PickupAccepted`, `FreightCostRecorded`,
`DeliveryOffered`, `DeliveryAccepted`, `RetailPriceReported`. Its fields are eventType,
operationId, batchId, actorMsp, txId, txTime, resultVersion, sorted objectIds,
previousState and newState (empty previousState on creation). Freight repeats the
same state. Transition metadata avoids a second event: Fabric supports one chaincode
event per transaction. No commercial values, signatures, salts or personal data occur.
Integration tests inspect business event allowlists and public/private hashed writes;
health blocks still contain no domain writes/events. Tests verify endorsements and
validation flags, including actual MVCC conflict and endorsement-policy rejection.

## Determinism and validation

- `GetTxID()`/`GetTxTimestamp()` supply execution metadata. Seconds stay a decimal
  string, nanos are bounded. Signed proposal time is not an independent wall clock
  or proof of a physical event; declared harvest dates never authorize transitions.
- Transaction code has no RNG, external calls, filesystem/environment access,
  host clock, locale sorting or floating-point money. Only server startup reads TLS
  files/package ID outside endorsement execution.
- Object/write keys and returned index order are deterministic; ID arrays are
  sorted/deduplicated. Restricted ASCII/integer public JSON is canonical; full
  cryptographic vectors reproduce the schema-restricted JCS encoding in Go and Java.
- Raw integer syntax rejects exponent/fraction/negative-zero/coercion/overflow.
  Monetary bounds keep all exact products within int64. Quantity is immutable
  integer grams, 1–100,000,000; unsupported unit fields fail instead of converting.
- Money is integer kuruş, TRY, EXCLUDING_TAX. Private purchase rate must be positive,
  totals exactly conserve quantity, retail/freight may be zero but not missing.
- Allowlists reject duplicate/unknown/prototype-sensitive keys, invalid Unicode,
  nulls and excessive size/depth. Arbitrary user text is never stored or logged.

## Error codes

Errors contain `schemaVersion=agrochain.error.v1`, `code`, fixed safe `message`,
`retryable=false`. No private input, certificate or internal exception is interpolated.
Backend-only operation/correlation fields are not fabricated by chaincode.

Implemented: INVALID_SCHEMA, INVALID_IDENTIFIER, UNSUPPORTED_SCHEMA_VERSION,
INVALID_QUANTITY, INVALID_MONEY, UNSUPPORTED_PILOT_OPERATION,
UNAUTHORIZED_ORGANIZATION, UNAUTHORIZED_ROLE, WRONG_TRANSFER_RECIPIENT,
BATCH_NOT_FOUND, TRANSFER_NOT_FOUND, OPERATION_NOT_FOUND, BATCH_ALREADY_EXISTS,
TRANSFER_ALREADY_EXISTS, LOT_ALREADY_EXISTS, COST_ALREADY_EXISTS,
REPORT_ALREADY_EXISTS, DUPLICATE_TRANSACTION, DOCUMENT_ALREADY_USED,
INVALID_STATE_TRANSITION, VERSION_CONFLICT, MISSING_EVIDENCE,
SOURCE_BINDING_MISMATCH, POLICY_MISMATCH, INVALID_PAGE,
EVIDENCE_VERIFICATION_UNAVAILABLE, INTERNAL_ERROR.

## Build, test and lifecycle

After [Stage 2 prerequisites](../../network/README.md), from repository root:

```bash
make chaincode-prerequisites  # pinned Go, repository-local; no host package install
make chaincode-deps           # pinned modules, go.sum verification
make chaincode-format
make test                    # network/config/package safety
make chaincode-test          # format check, vet, race tests, coverage, static build
make bootstrap
make verify
make chaincode-deploy        # test/build/package/install/approve/readiness/commit/check
make privacy-prepare         # development certificate roles and signing keys
make stage4-check            # includes populated private/public restart check
```

On an already deployed network, `make stage3-check` combines configuration/package
tests, Go format/vet/race/build checks, live network and definition verification,
the six integration tests, and full data-preserving restart verification. Run it
with Python >=3.10 selected in PATH. It commits health probes and one deliberately
invalid single-endorser transaction, not business records.

The chaincode CI workflow runs the Stage 4 gate on a fresh Ubuntu runner after bootstrap and
deployment, then checks repeat deployment. A configured workflow is not evidence
that a hosted run has passed; actual local results are in the completion report.

Domain tests include all 56 command/checkpoint combinations, late-freight recovery
after rejected receipt, populated shared queries/history/index cleanup, and X.509
role extraction through Fabric's identity API. Positive domain tests use test-only
evidence; additional Stage 4 tests use the production verifier and live populated
Fabric/PDC state. No hosted Fabric CA service is claimed.

Independent commands: `make chaincode-package`, `chaincode-install`,
`chaincode-approve`, `chaincode-commit`, `chaincode-check`. They consume the same
definition from `network/config/chaincode.env`. Combined deploy tests first;
individual lifecycle commands assume previous steps succeeded.

Toolchain: checksum-pinned **Go 1.27.1**, Linux amd64. Direct modules:
**fabric-chaincode-go/v2 v2.3.0**, **fabric-protos-go-apiv2 v0.3.6**,
**protobuf v1.36.5**. Indirect modules: grpc **1.70.0**, x/net **0.34.0**,
x/sys **0.29.0**, x/text **0.21.0**, genproto/googleapis/rpc
**v0.0.0-20250115164207-1a7da9e5054f**. Exact versions/checksums are in go.mod/go.sum.

A static binary runs in four independent scratch CCAAS containers on the existing
private Docker network, without host ports or Docker socket mounts. Each has a
development TLS server certificate; lifecycle packages contain only its public root,
never a key. Mutual peer-to-chaincode TLS/client authentication is not configured.
Server TLS, private network and shared-host trust boundaries must not be overstated.

Archive entry order/modes/ownership/timestamps are fixed; labels include the binary
digest and the Fabric CLI calculates package IDs. Packages reproduce byte-for-byte
for the same binary and existing TLS material. Fresh random development certificates
intentionally change IDs. Generated binaries/packages/caches/TLS/blocks are ignored.
Preserve `network/runtime/chaincode` for restart; deleting its TLS files is not a
routine rebuild. Images use scratch and need no base-image download.

Fresh channels start at sequence 1. Repeated deployment reuses a matching committed
version/sequence only if package bytes match. Upgrades require a new semantic version
and explicit next sequence, never a clock-based increment. Example after deliberately
updating code and the configured release:

```bash
make chaincode-upgrade CHAINCODE_VERSION=0.2.1 CHAINCODE_SEQUENCE=5
```

The current test network is release 0.2.1/sequence 5 after explicit data-preserving
upgrades. A fresh channel uses the same 0.2.1 source at sequence 1. Upgrades recreate
services and interrupt queries; rolling availability is not promised. Retain packages
and TLS after a failure, fix the cause and resume lifecycle steps without resetting
ledgers. Candidate packaging does not replace active service configuration until
installation checks pass.

`make network-down` stops chaincode before peers and preserves state;
`make network-up` restores both. `make clean-generated` explicitly destroys demo
ledgers/identities/runtime after shutdown and was not used here. Tool caches and
local Docker images may remain after cleanup.

## Stage 5 handoff and known limitations

`evidenceProvider.prepare(Command, Batch, EvidenceRefs)` is an internal **test seam**,
not a public API accepting claimed verification. Its typed `verified` result supplies
normalized reference/quantity/money checks. The deployed dispatcher exclusively
constructs `realEvidence`, which implements trusted source signatures, body binding,
source/nonce deduplication, commitment opening checks and atomic PDC persistence.
No verifier may be selectable by public input or deployment environment.

Purchase openings → tradePrivate; freight openings/costs → freightPrivate; retail
reports → retailAuditPrivate. These collections are exercised on real member and
nonmember peers. Original-byte verification exists in the off-chain acceptance
helper; its service/API integration belongs to Stage 5. Source fixtures are labeled
SIMULATED. Institutional adapters and anomaly logic remain unimplemented.

Full live lifecycle, duplicate submissions, concurrent creation MVCC rejection,
business events and populated public/private persistence are tested. Handoff-specific
race modeling remains in unit tests; the live race submits duplicate creation.
General Unicode/floating-point JCS inputs are deliberately rejected. No production scalability,
legal interpretation or government integration is claimed.

References: [official Go shim](https://github.com/hyperledger/fabric-chaincode-go/tree/v2.3.0)
and [Fabric 2.5 external service protocol](https://hyperledger-fabric.readthedocs.io/en/release-2.5/cc_service.html).
