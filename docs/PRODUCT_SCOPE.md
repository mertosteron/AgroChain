# AgroChain — product scope

Status: Stage 1 technical contract, version 1.0, 2026-09-16. All runtime capabilities below are **planned**, not implemented. Normative defaults apply until explicitly revised in these documents.

## Evidence and proposal review

A recursive inspection, including hidden entries, found only the root `AGENTS.md`. No README, previous architecture, presentation, source code, tests, deployment configuration, credentials, government access evidence, benchmark results, or Git metadata was supplied. There is no existing code to accommodate and no presentation to evaluate. Absence of evidence here is not evidence about systems outside this directory.

| Proposal issue | Stage 1 resolution |
| --- | --- |
| `AGENTS.md` requires a root architecture; requested deliverables use `docs/` | Root `ARCHITECTURE.md` links to the authoritative document in `docs/`; no duplicated specification. |
| Example lifecycle conflates transfer, transport and sale | Define custody separately from ownership, explicit two-party handoffs, and retail price reporting rather than claiming a completed consumer sale. |
| `AGENTS.md` names three simulators; current brief also names HKS | Include four signed simulators: ÇKS, e-Fatura, HKS and U-ETDS. |
| Fabric transaction time could be mistaken for authoritative physical-event time | Use Fabric proposal time as a consistent ledger attribute; it is client-originated, not an independently trusted clock. |
| Possible alternate name “Gıda-TL” in the brief | It does not appear in supplied repository material. Reserve it as an unverified legacy proposal label, not a component or second product name. Use AgroChain throughout. |
| National-scale, zero-cost, ministry access, legal price-cap or measured performance claims | None is supported. This contract makes no such claim. No legal rule or model-training dataset was supplied. |

## Problem and differentiating claim

Supply-chain records can be disconnected across organizations, commercial amounts need controlled disclosure, and unusual price changes need an explainable review trail. AgroChain will demonstrate that four identified organizations can agree on the history of one tomato batch, retain integrity evidence for signed simulated source records, restrict commercial data, and record a reproducible price-review signal.

Blockchain establishes agreement over accepted records and their ordering. It cannot establish that a harvested quantity, declared cost or physical delivery is truthful. A normal score also does not establish an absence of fraud.

## Users and precise route

| User | Pilot task |
| --- | --- |
| Producer/cooperative operator | Register a tomato batch and offer the whole batch for transport using signed institutional simulator evidence. |
| Logistics operator | Accept custody, record the freight charge and offer delivery. Never becomes the commercial owner. |
| Retailer operator | Accept delivery and ownership, map the entire batch to one retail lot, and report its offered unit selling price. |
| Regulator reviewer | Read all pilot commercial collections, reproduce the score, review a signal and record an explanation. This is a fictional pilot organization, not an actual ministry endpoint. |
| Consumer | Open a QR URL and read an allowlisted non-confidential provenance summary. Has no Fabric identity. |

The pilot has one producer, one logistics provider, one retailer and one regulator organization. Tomatoes move producer → logistics custody → retailer ownership/custody → regulator review. A 100 kg bulk batch maps 1:1 to a 100 kg retail lot; the QR identifies that lot, not each tomato or consumer purchase.

## Scope

In scope: one channel; four MSPs; reproducible local Raft-based Fabric network; explicit ownership/custody transitions; organization authorization; immutable operation receipts; duplicate protection; signed, clearly labeled synthetic institutional records; salted commitments; commercial PDCs; a Spring Boot backend and Fabric Gateway; one deterministic price-change rule; actor, reviewer and consumer views; three seeded scenarios; automated positive and negative tests; a documented offline demonstration.

Out of scope: real government APIs or credentials; a national registry; legal rulings or fines; payments, wallets or tokens; automated tax calculations; ERP replacement; IoT proof of physical truth; nationwide throughput claims; trained ML; multiple currencies; real personal data; batch merging/splitting; partial deliveries/sales; weight loss adjustments; returns, cancellation and invoice correction flows; multi-retailer channel design; Kubernetes, Kafka and Redis dependencies. Physical discrepancies stop the batch instead of silently changing quantity.

## Three demo scenarios

All amounts are synthetic TRY, excluding tax and discounts, with the same tomato grade and kg basis. Purchase and retail amounts are confidential. Freight is displayed as context, not subtracted in the scoring formula. The rule measures price increase, not net profit.

| Scenario | Fixed inputs | Expected observations |
| --- | --- | --- |
| Normal, `BAT-NORMAL01` / `LOT-NORMAL01` | 100,000 g; purchase 2,000 kuruş/kg; freight 20,000 kuruş total; retail 2,800 kuruş/kg; threshold 5,000 basis points (50%) | Whole route commits; increase 40%; `NO_SIGNAL`; consumer sees provenance but no prices, threshold calculation or commercial documents. |
| Suspicious, `BAT-SUSPECT1` / `LOT-SUSPECT1` | Same quantity, purchase and freight; retail 3,700 kuruş/kg | Increase 85%; `REVIEW_REQUIRED`; regulator opens and resolves a review with an explanation, without a legal judgment. |
| Integrity/replay, `BAT-TAMPER01` | Valid original signed invoice, altered bytes/body, then an already-used invoice/command resubmitted | Original verifies; altered attachment fails `DOCUMENT_HASH_MISMATCH`; changed signed header fails `INVALID_SOURCE_SIGNATURE`; repeated operation creates no second record; reused source document under a new operation fails `DOCUMENT_ALREADY_USED`. Rejected attempts do not advance batch state or consume evidence. |

Use distinct identifiers per scenario. The third scenario contains both tamper and replay checks; duplicate submission is not itself proof of malicious intent. Successful byte verification is performed against the committed original, not a newly generated commitment to the altered document.

## Deterministic first-version rule

Let `P` be the positive purchase price in kuruş/kg, `R` the nonnegative offered retail price on the same basis, and `T` the integer threshold in basis points. A signal is produced exactly when:

```text
(R - P) * 10000 > P * T
```

Default `T = 5000`. Equality at 50% is `NO_SIGNAL`. Calculate with checked integer arithmetic (arbitrary-precision intermediates preferred), never binary floating-point. Display `floor((R-P)*10000/P)` as `increaseBps`, formatted as percent with two decimal places; classification uses the exact comparison, not rounded display. Negative differences are permitted. There is no probabilistic confidence score. Invalid/missing inputs produce an error or an explicitly pending result, never `NO_SIGNAL`.

Policy `price-increase-v1` is versioned and snapshotted by each price report. The backend proposes the result; chaincode recomputes it before accepting the attestation. Freight helps the reviewer contextualize the signal but this first rule does not estimate margins, spoilage, seasonality, taxes or regional market conditions. Parameters are pilot settings, not legal limits.

## Measurable success criteria

These are future acceptance targets, not measurements:

1. A clean supported machine starts four organization peers, one channel and the documented orderer configuration through scripts; it can deploy, submit, query, stop and reset with no undocumented steps.
2. All three scenarios run successfully after a clean seed; every successful operation has a valid committed transaction ID and a queryable receipt.
3. Normal and suspicious results reproduce the exact values above; the 50% boundary, zero purchase rejection and concurrent duplicate cases are tested.
4. All unauthorized mutation, unauthorized PDC query and consumer-field leakage tests pass; original document verification succeeds and changed document verification fails.
5. For every handoff, the immutable 100,000 g quantity and exclusive custody/ownership remain consistent; no unimplemented split/loss path is accepted.
6. After two independent resets, identical business outcomes recur. Cryptographic salts, keys, transaction IDs and timestamps need not recur.
7. A teammate performs the three-minute presentation flow from the Stage 8 guide. Record actual transaction/query durations and environment in Stage 7; no latency or scaling result is promised here.

## Pilot versus ambition

A later national deployment would require institutional agreements, legal and privacy review, independently operated infrastructure, realistic trading models, governance, availability planning and measured capacity. Those are separate projects. This pilot provides evidence about a local multi-organization workflow and its controls only.

## Proposed three-minute live flow

Preflight and seed both completed normal and suspicious histories before the talk; keep the suspicious review open and the third batch ready for integrity checks.

| Time | Action and evidence |
| --- | --- |
| 0:00–0:30 | Show four organizations and the “SIMULATED institutional data / real local Fabric transactions” banner. |
| 0:30–1:10 | Follow the normal batch from producer to retailer, show two-party custody receipts and 40% / no signal in the authorized view. |
| 1:10–1:50 | Compare the suspicious batch: 20 → 37 TRY/kg, 85% against a 50% pilot parameter; regulator submits a review action and shows its committed receipt. |
| 1:50–2:25 | Verify the original invoice, reject a changed copy, retry a used operation and show unchanged state. |
| 2:25–3:00 | Show logistics denied access to trade prices; open consumer QR showing only provenance. State limitations and the fallback recording location. |

The complete route must also be executable through the seed/demo script; pre-seeding presentation history is not a substitute for the end-to-end test. A recording or fixture-only fallback must be labeled as recorded/simulated presentation material, never shown as a live commit.
