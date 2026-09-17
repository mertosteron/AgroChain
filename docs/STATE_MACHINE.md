# AgroChain — state machines

Status: Stage 1 normative plan. Lifecycle names replace the illustrative sequence in `AGENTS.md`. [Data contracts](DATA_CONTRACTS.md) define command and record schemas; [security](SECURITY_AND_PRIVACY.md) defines endorsements separately from callers.

## Product lifecycle

```text
ABSENT --CreateBatch--> CREATED
CREATED --OfferPickup--> PICKUP_PENDING
PICKUP_PENDING --AcceptPickup--> IN_TRANSPORT
IN_TRANSPORT --OfferDelivery--> DELIVERY_PENDING
DELIVERY_PENDING --AcceptDelivery--> RECEIVED
RECEIVED --ReportRetailPrice--> RETAIL_REPORTED
```

`ABSENT` is not a stored state. `RETAIL_REPORTED` means a retailer declared an offered price; it is not proof of a completed sale. Anomaly status is not a product lifecycle state.

Every mutation checks schema, actor MSP/role, operation uniqueness, `expectedVersion`, state and all evidence before writing. Batch version starts at 1 and increments for every batch-related successful command, including cost, evaluation and review commands that do not change product state. Transitions, PDC writes, receipts and source-use indices are atomic. Fabric MVCC invalidation is failure, even if the proposal simulation succeeded.

| Command / caller | Preconditions and inputs | Effects / event | Specific failures |
| --- | --- | --- | --- |
| `CreateBatch` / Producer producer | Batch absent; expectedVersion 0; positive bounded whole-batch quantity; tomatoes/grade/region; unused signed CKS eligibility bound to this batch/operation and matching producer/quantity | State CREATED; owner=custodian=Producer; fixed route to Logistics/Retailer; version 1; `BatchCreated` | BATCH_ALREADY_EXISTS, INVALID_QUANTITY, SOURCE_BINDING_MISMATCH, INVALID_SOURCE_SIGNATURE |
| `OfferPickup` / current Producer owner | CREATED; no pending transfer; whole quantity; selected Logistics recipient; matching signed purchase invoice, HKS notification and UETDS manifest; all unused and bound to this command | Write trade invoice/opening; set PICKUP_PENDING; create pickup transfer OFFERED; owner/custodian remain Producer; `PickupOffered` | INVALID_STATE_TRANSITION, UNAUTHORIZED_ORGANIZATION, DOCUMENT_ALREADY_USED, DOCUMENT_HASH_MISMATCH |
| `AcceptPickup` / named Logistics carrier | PICKUP_PENDING; transfer ID matches pending offer; accepts exact quantity and sender | Set IN_TRANSPORT; pickup transfer ACCEPTED; custodian=Logistics; owner remains Producer; `PickupAccepted` | TRANSFER_NOT_FOUND, WRONG_TRANSFER_RECIPIENT, INVALID_QUANTITY |
| `RecordFreightCost` / current Logistics carrier | IN_TRANSPORT or DELIVERY_PENDING; one cost per batch; unused signed freight invoice, payer=Retailer, payee=Logistics, quantity matches | Create freight PDC cost/opening; state unchanged; `FreightCostRecorded` | COST_ALREADY_EXISTS, INVALID_MONEY, SOURCE_BINDING_MISMATCH |
| `OfferDelivery` / current Logistics carrier | IN_TRANSPORT; no active delivery offer; recipient fixed Retailer; full quantity; same route manifest already accepted | Set DELIVERY_PENDING; create delivery transfer OFFERED; owner Producer, custodian Logistics; `DeliveryOffered` | INVALID_STATE_TRANSITION, UNAUTHORIZED_ORGANIZATION, INVALID_QUANTITY |
| `AcceptDelivery` / named Retailer retailer | DELIVERY_PENDING; matching transfer; full quantity acknowledged; freight cost already recorded | Set RECEIVED; delivery transfer ACCEPTED; owner=custodian=Retailer atomically; `DeliveryAccepted` | WRONG_TRANSFER_RECIPIENT, INVALID_QUANTITY, MISSING_EVIDENCE, VERSION_CONFLICT |
| `ReportRetailPrice` / current Retailer owner/custodian | RECEIVED; purchase and freight records exist; positive purchase and nonnegative retail rate; matching currency/tax basis/quantity; unique retail lot ID and report; active policy ID | Set RETAIL_REPORTED; create 1:1 whole-batch lot mapping and private report; mark evaluation pending; `RetailPriceReported` | MISSING_EVIDENCE, LOT_ALREADY_EXISTS, INVALID_MONEY, POLICY_MISMATCH |

Freight must be recorded before acceptance if it was not already recorded in transport. An early delivery acceptance fails without changing state, allowing Logistics to record the missing cost and Retailer to retry. The demo script records freight before offering delivery. No backdated freight mutation exists after RECEIVED in v1.

Each evidence record is consumed once during its original operation. Later transitions reference the already accepted manifest/invoice by ID without re-consuming or replacing it. Source document ID deduplication applies to ingestion, not to read references. A future real CKS registration could support multiple batches, but the pilot consumes a unique batch-bound eligibility attestation per creation, not an entire reusable registration document.

## Transfer agreement and invariants

There is one active transfer per batch, with `kind=PICKUP` or `DELIVERY` and `status=OFFERED|ACCEPTED`. The sender creates an offer; only its nominated recipient can accept it. Submission on behalf of both parties from one identity is forbidden. No cancellation/expiry in v1: a refused or disputed handoff remains pending for operator review; do not manually edit the ledger to advance it.

1. `quantityGrams` is a positive immutable integer. Every invoice, cost basis, manifest, transfer and retail lot agrees with it. Quantities cannot be negative, zero, silently rounded or exceed available stock.
2. Exactly one owner and one custodian exist. Initially both are Producer; during transit only custody is Logistics; Retailer acceptance changes both. Logistics never becomes owner.
3. No split, merge, shrinkage, reclassification, reweigh adjustment or partial sale exists. Whole-batch quantity at creation = transport quantity = received quantity = lot quantity. A mismatch is rejected and leaves the prior state intact.
4. One producer purchase invoice, one freight cost, one retail report and one evaluation per batch; no overwrite. Corrections require a future explicit contract, not resubmission under another ID.
5. Batch ID and retail lot ID are unique. A lot cannot map to multiple batches and a batch cannot map to multiple lots.
6. Every accepted mutation has exactly one immutable operation receipt and event; replay does not produce another transition. A query has no mutation effects.
7. Source-use and nonce keys are written atomically with their business operation. Concurrent simulations may both succeed but at most one conflicting transaction can commit validly.
8. Declared dates cannot authorize transitions. Ledger version and committed sequence establish ordering.

## Separate anomaly/review lifecycle

```text
Retail report committed -> EVALUATION_PENDING (projection marker)
EvaluatePrice -> NO_SIGNAL / reviewState=NOT_REQUIRED
             -> REVIEW_REQUIRED / reviewState=OPEN
OPEN --OpenReview--> IN_REVIEW
IN_REVIEW --ResolveReview--> RESOLVED
```

`EVALUATION_PENDING` means no accepted anomaly record yet. It must survive worker restart through reconciliation of unevaluated reports. The deterministic classification remains immutable after review, including when a reviewer finds a reasonable explanation.

| Command / caller | Preconditions | Atomic effect / event |
| --- | --- | --- |
| `EvaluatePrice` / Regulator oracle | RETAIL_REPORTED; report/purchase/freight PDC records and policy available; no existing evaluation; command carries private proposed result | Chaincode rederives exact score from committed records; rejects mismatching proposed result; creates private anomaly record and public opaque evaluation reference; `PriceEvaluated` contains no classification or prices |
| `OpenReview` / Regulator reviewer | Classification REVIEW_REQUIRED; reviewState OPEN | Append private action; reviewState IN_REVIEW; `ReviewUpdated` |
| `ResolveReview` / Regulator reviewer | IN_REVIEW; nonempty explanation 1–2,000 characters; allowed outcome | Append private action; reviewState RESOLVED; `ReviewUpdated` |

Allowed outcomes: `EXPLAINED`, `FOLLOW_UP_RECOMMENDED`, `INSUFFICIENT_EVIDENCE`. None denotes a legal violation, fine or exoneration. No review commands for NOT_REQUIRED, no direct OPEN → RESOLVED, no reopening or editing a prior resolution in v1. Failed scoring returns `MISSING_EVIDENCE`, `INVALID_MONEY`, `POLICY_MISMATCH` or `ANOMALY_RESULT_MISMATCH`; it never defaults to NO_SIGNAL. Invalid review transition returns `INVALID_REVIEW_TRANSITION`.

## Event and query contract

One event per accepted command, with `schemaVersion=agrochain.event.v1`, `eventType`, `operationId`, `batchId`, `actorMsp`, `txId`, `resultVersion`, and relevant opaque `objectIds`. Consumers consult commit validity and never treat an endorsement event as final. Retail/evaluation/review events reveal operation occurrence to channel members but exclude rates, classification, text, private hashes and salts. The consumer projection includes only creation and accepted custody/provenance history.

Keep append-only receipts keyed by batch/version plus MSP/operation lookup, so history does not depend on an optional peer history database. Current transfer and batch state are queryable separately. Conflicts require re-read and user-visible retry decision; they are not permission to replace `expectedVersion` and silently repeat a meaningful operation.

## Required transition tests in later stages

Cover every allowed edge and at least one forbidden edge from each stored state, wrong-MSP and wrong-role calls, wrong recipient, missing offer, repeated acceptance, duplicate operation, changed operation replay, reused invoice/nonce, negative/zero/mismatched quantity, wrong lot mapping, transfer race and MVCC failure. Verify no partial public/private writes after rejection. Check private scoring at 40%, exactly 50%, just above 50%, 85%, a falling price, zero purchase and missing cost. Verify a resolved anomaly still retains its original classification and append-only review history.
