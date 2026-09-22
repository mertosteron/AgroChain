# AgroChain — security and privacy contract

## Actual Stage 4 enforcement (2026-09-21)

Mandatory certificate roles replace the historical absent-role fallback. Immutable
Regulator-admin bootstrap registers four simulated source keys. Real chaincode
checks signatures, salted commitments, body bindings and replay indices before
atomically writing public records and three PDCs. Member/nonmember clients and
nonmember peers, original/modified documents, shared block/event/log leakage and
populated restart are tested. See [report](STAGE4_REPORT.md) and
[runtime contract and limitations](STAGE4_PRIVACY.md). No government API is connected.

## Historical Stage 3 enforcement (2026-09-18; superseded above)

The user explicitly selected the Stage 1 fail-closed deployment boundary. Official
Fabric client identity APIs supply the MSP and any `agrochain.role` attribute;
caller-supplied MSP/role/transaction/time fields are rejected. All seven business
entry points are present, but authorized well-formed calls stop with
EVIDENCE_VERIFICATION_UNAVAILABLE before state reads/writes/events. Successful domain
tests use a provider compiled only into test binaries. No runtime switch or admin
method enables it. No evidence assertion is accepted on the network.

Cryptogen identities still lack business attributes. Per the user's explicitly
allowed Stage 3 fallback, missing attributes receive MSP-only checks; a present
attribute must match the organization's business role. This is a temporary
authorization limitation, including for attribute-less Admin identities, not proof
of application-user role enforcement. Stage 4 must provision and require roles
before enabling writes. Regulator has no custody mutation authority.

The chaincode definition and unused collection definitions retain Retailer +
Regulator endorsements. All four admins approved it. Real integration tests verify
that a Regulator-only endorsed health transaction is invalid at commit. Two-party
custody consent in domain tests is separate from peer endorsement; no SBE is added.

Only public DTOs are serializable as state/events. Retail price remains in the
planned retailAuditPrivate record; Stage 3 stores no monetary values anywhere.
No PDC handling, salt/signature verification, source registry or document integrity
implementation exists yet. Public commitment/reference interfaces are internal test
seams, not verified institutional evidence. See [Stage 3 report](STAGE3_REPORT.md)
and [chaincode guide](../chaincode/agrochain/README.md) for exact tests and handoff.

Four external chaincode services use server TLS on the private Docker network, no
host ports, no Docker socket, read-only filesystems and dropped capabilities. Their
development private keys remain in ignored runtime files and never enter packages.
Mutual peer-to-chaincode TLS is not configured; the shared host/network operator
remains trusted. These controls do not establish independent-host isolation.

The following is the Stage 1 normative contract; Stage 5/6 backend and review controls remain planned. The Stage 4 runbook identifies implemented controls. [Data contracts](DATA_CONTRACTS.md) define exact fields; [state machines](STATE_MACHINE.md) define operation preconditions.

## Identities, submission and endorsement

Each organization has its own MSP and test enrollment/signing credentials. A certificate carries `agrochain.role`: `producer`, `carrier`, `retailer`, `reviewer`, `auditor`, `oracle`, `public-reader` or `admin`, assigned by that organization's enrollment authority. Chaincode verifies the allowed MSP/role pair; an arbitrary client-supplied role string is not authority. Regulator `oracle` can evaluate but cannot resolve human reviews. Admin credentials do not automatically get business-operation permissions.

The backend keeps separate signers and selects one from the authenticated session, never an MSP header provided by a browser. One process holding several signers is a demo convenience and a high-trust component. Tests must also call chaincode directly with wrong identities to demonstrate that backend bypass does not bypass authorization.

| Operation group | Submitting identity | Required peer endorsements | Read/audit |
| --- | --- | --- | --- |
| Initialize trust/config | RegulatorMSP/admin, once | Retailer + Regulator | Shared configuration readable by channel members |
| Create / offer pickup | ProducerMSP/producer | Retailer + Regulator | Shared route readable by all channel organizations; trade evidence only trade members |
| Accept pickup / offer delivery / record freight | LogisticsMSP/carrier | Retailer + Regulator | Shared custody readable by all; freight values only freight members |
| Accept delivery / report retail | RetailerMSP/retailer | Retailer + Regulator | Shared route readable by all; offered price only retail audit members |
| Evaluate price | RegulatorMSP/oracle | Retailer + Regulator | Detailed result only Retailer and Regulator |
| Open / resolve review | RegulatorMSP/reviewer | Retailer + Regulator | Review detail only Retailer and Regulator |
| Shared query | Any registered business role or Regulator public-reader | No transaction endorsement/ordering needed for evaluate | Backend limits consumer output separately |
| Private query / original verification | Relevant member MSP and its business role; Regulator reviewer/auditor/oracle | Evaluate on authorized member peer | Regulator audits every collection; no consumer access |

Concrete initial chaincode policy: `AND('RetailerMSP.peer','RegulatorMSP.peer')`. All three collections use the same write endorsement policy. This deliberately favors one feasible endorsement path over availability: both organizations belong to every accessed collection, including evaluation spanning collections. Use explicit endorser targeting for private proposals. Do not use a channel-wide majority requiring a nonmember to see private inputs.

This policy does not mean a human retailer approved each action. Peer endorsement certifies execution of rules, while signed actor submissions and the two-step handoffs establish application consent. The retailer can stall regulator writes by refusing endorsement, and either required peer being unavailable stops writes. Those are explicit pilot limitations. Production governance could revise the policy after testing, but the pilot must not silently fall back to weaker endorsements.

## Private collections

| Collection | Distribution/membership policy | Values |
| --- | --- | --- |
| `tradePrivate` | `OR('ProducerMSP.member','RetailerMSP.member','RegulatorMSP.member')` | Purchase invoice body, commitment salt/opening, purchase rate and total |
| `freightPrivate` | `OR('LogisticsMSP.member','RetailerMSP.member','RegulatorMSP.member')` | Freight invoice, salt/opening and total cost |
| `retailAuditPrivate` | `OR('RetailerMSP.member','RegulatorMSP.member')` | Retail price reports, anomaly calculation, review explanation and request digests for these operations only |

Proposed values for each: `memberOnlyRead=true`, `memberOnlyWrite=true`, `blockToLive=0`, `requiredPeerCount=1`, `maxPeerCount=2` for three-member collections and `1` for the two-member collection. One peer per organization means the required counterpart must be reachable. Test gossip/anchor configuration and retrieval after a member reconnects before claiming durability. `blockToLive=0` intentionally retains private values for the demo network lifetime; it does not mean zero retention. Collection controls and the distinction between membership and endorsement are described in [Fabric private data configuration](https://hyperledger-fabric.readthedocs.io/en/latest/private-data-arch.html).

Client role and per-operation checks remain necessary in chaincode. Members' peer administrators can inspect their private stores despite application restrictions. PDC membership is at organization level: a second retailer admitted to `RetailerMSP` would share its organization's data. The pilot uses exactly one retailer; extending to competing retailers requires revisiting MSP and collection governance, not one channel per retailer by default.

No real PII is permitted. Keep synthetic private data until an explicit, documented demo reset. Protect local stores/backups through host access control and encryption at rest where available. Reset/purge cannot recall copies held by members or erase public ledger hashes. Legal retention, deletion obligations and regulatory access basis require separate review; PDCs do not establish compliance.

## Field-level storage classification

“Shared” means readable to channel members, including their node administrators. It does **not** mean internet-public. Every new field must be explicitly classified before use; an unclassified field is rejected from public serializers.

| Contract fields | Authoritative location | Consumer exposure |
| --- | --- | --- |
| Batch `schemaVersion`, `batchId`, `productCode`, `gradeCode`, `quantityGrams`, `originRegionCode`, `harvestDate` | Shared | Yes, except schema version may be rendered as interface version |
| Batch `producerMsp`, `ownerMsp`, `custodianMsp`, `intendedRetailerMsp`, `logisticsMsp` | Shared organization references, no people | Publish predefined organization display labels only |
| Batch `state`, `version`, `retailLotId` when assigned | Shared | Yes |
| `createdTxId`, `updatedTxId`, `txTime.seconds`, `txTime.nanos`; receipts/history | Shared | Selected transaction IDs and formatted recorded time; no claim of verified physical-event time |
| Document `header.schemaVersion`, `documentId`, `sourceSystem`, `sourceMode`, `issuerId`, `keyId`, `sourceDocumentId`, `documentType`, `batchId`, `boundOperationId`, `issuedAt`, `nonce`, `commitmentAlgorithm`, `commitment`, `signatureAlgorithm`; `signature` | Shared; `sourceDocumentId` must be opaque, not a natural invoice/identity number | Only source system, mode and verification label; not raw envelope/IDs/hash |
| CKS body `producerMsp`, `productCode`, `gradeCode`, `quantityGrams`, `originRegionCode`, `harvestDate`, `eligible`; HKS body `producerMsp`, `retailerMsp`, `productCode`, `quantityGrams`, `notified`; UETDS body `carrierMsp`, `consignorMsp`, `consigneeMsp`, `productCode`, `quantityGrams`, `declaredDepartureAt`; each opening's `saltHex` | Controlled off-chain store; transmitted transiently for verification; selected already-classified batch/transport attributes copied to shared state | Only approved provenance projection |
| Purchase invoice `sellerMsp`, `buyerMsp`, `quantityGrams`, `priceKurusPerKg`, `totalKurus`, `currency`, `taxBasis`, `attachmentDigest`, `attachmentMediaType`, `saltHex` | `tradePrivate` plus authorized off-chain original | No |
| Freight body `carrierMsp`, `payerMsp`, `quantityGrams`, `totalKurus`, `currency`, `taxBasis`, `attachmentDigest`, `attachmentMediaType`, `saltHex` | `freightPrivate` plus authorized off-chain original | No |
| Custody `schemaVersion`, `transferId`, `batchId`, `kind`, `status`, `fromMsp`, `toMsp`, `quantityGrams`, `offerOperationId`, `acceptOperationId`, `offeredTxId`, `acceptedTxId` | Shared; acceptance fields absent until accepted | Sanitized organization route and status only |
| Cost `schemaVersion`, `costId`, `batchId`, `kind`, `payerMsp`, `payeeMsp`, `quantityGrams`, `totalKurus`, `currency`, `taxBasis`, `sourceDocumentId`, `recordSaltHex`, `operationId`, `txId` | `freightPrivate`; opaque `costId` and presence also shared | No |
| Retail report `schemaVersion`, `reportId`, `batchId`, `retailLotId`, `quantityGrams`, `retailerMsp`, `offeredPriceKurusPerKg`, `currency`, `taxBasis`, `purchaseDocumentId`, `freightCostId`, `policyId`, `reportedAt`, `recordSaltHex`, `operationId`, `txId` | `retailAuditPrivate`; IDs/lot mapping/quantity and evaluation pending marker also shared | Only lot mapping/quantity, not commercial record references |
| Anomaly `schemaVersion`, `anomalyId`, `batchId`, `reportId`, `policyId`, `ruleVersion`, `purchasePriceKurusPerKg`, `retailPriceKurusPerKg`, `freightTotalKurus`, `quantityGrams`, `increaseBps`, `thresholdBps`, `classification`, `reasonCode`, `reviewState`, `recordSaltHex`, `operationId`, `txId` | `retailAuditPrivate`; only anomaly ID/report ID/existence shared | No classification or review signal |
| Review action `schemaVersion`, `operationId`, `anomalyId`, `action`, `outcome`, `explanation`, `reviewerRef`, `recordSaltHex`, `txId` | `retailAuditPrivate` append-only actions | No |
| Command `schemaVersion`, `operationId`, `batchId`, `expectedVersion`, `command`, `payload` | Shared transaction input only if all nested payload fields are approved shared fields | No raw commands |
| Receipt `schemaVersion`, `operationId`, `batchId`, `command`, `actorMsp`, `txId`, `txTime`, `resultVersion`, `objectIds`; public event fields | Shared | Only selected provenance events |
| `privateRequestDigest`, transient evidence openings, PDC random record salts | Affected authorized PDC; never unsalted public request digests of prices | No |
| Original XML/PDF/JSON, local storage path, real-source locator if later allowed | Controlled off-chain file/metadata store; source-specific ACL mirrors relevant collection (CKS/HKS/UETDS: all four business organizations) | No |
| Source registry `schemaVersion`, `issuerId`, `sourceSystem`, `documentTypes`, `keyId`, `publicKeyRawBase64url`, `enabled`, `createdTxId`; policy `schemaVersion`, `policyId`, `ruleVersion`, `thresholdBps`, `currency`, `taxBasis`, `comparison`, `createdTxId` | Shared | Optional generic explanation of pilot, no batch commercial outcome |
| Certificates/private keys, tokens, passwords, signing seeds | Local protected secret store; public certificates used by Fabric as required | Never secrets |
| Logs, authenticated user mapping, replay/recovery staging, access history | Controlled off-chain application storage | No |

Raw natural identifiers (tax IDs, national IDs, addresses, license plates) are excluded, including from log messages and signed public headers. Application IDs are opaque and non-semantic outside test fixtures. Public commitments/hashes can still reveal linkage, timing and equality; do not claim metadata secrecy.

## Input authenticity, replay and the oracle problem

The source signs a domain-separated envelope binding its commitment to issuer, system, document type, batch, intended operation, nonce and schema. Both backend and chaincode verify with ledger-allowlisted source keys. A signature proves control of a configured simulator key and integrity of its asserted payload, not a real government attestation or physical truth. Backend parsing of original files into normalized claims remains an oracle trust boundary; hash checks establish byte consistency, not semantic correctness.

Every command atomically writes an immutable operation receipt keyed by submitting MSP + operation ID. Reuse returns `DUPLICATE_TRANSACTION`; backend recovery can retrieve the existing receipt. It never executes the command a second time or treats changed input as a new effect. Source consumption indices are unique by `(issuerId, sourceDocumentId)` and `(issuerId, nonce)`, and must match the bound batch and operation. Cross-batch or cross-operation reuse fails. Fabric transaction IDs alone do not prevent a freshly signed proposal repeating a business operation.

Random 32-byte commitment salts and per-record PDC salts are generated before submission and remain private. Predictable commercial PDC values require random `recordSaltHex` inside their serialized value because Fabric also publishes hashes of private writes. Do not publish an unsalted price digest as a convenience index. [Fabric's private-data model](https://hyperledger-fabric.readthedocs.io/en/release-2.4/private-data/private-data.html) explains which hashes are retained on the channel; the application's salted commitments are a separate protocol.

## Threat model and required evidence

| Threat | Control and planned test | Residual risk |
| --- | --- | --- |
| Wrong organization claims custody/ownership | MSP/role + current owner + pending recipient checks; direct-chaincode denial tests | Stolen authorized keys can submit permitted claims |
| One party asserts an unaccepted delivery | Two signed actor operations, immutable quantity and version; deny self-acceptance/wrong receiver | Colluding sender/receiver can lie about physical movement |
| Changed/replayed invoice | Signature, salted opening, attachment digest, operation and source-use indices; tamper/replay/race tests | Authorized source can sign false data; key compromise requires operational recovery |
| Backend forges a simulator verification flag | Chaincode verifies signed normalized payload and binding; call directly with forged evidence | Same-host compromise can steal both backend and simulator credentials |
| Price leaks | PDC policies, role checks, transient routing, record salt, no sensitive event fields, consumer allowlist; query/block/log inspection | Collection member/admin misuse, host compromise, metadata leakage |
| Replay/concurrent state updates | Atomic receipt/evidence keys, expected version, Fabric MVCC validation; concurrent requests test | Retry handling must correctly reconcile unknown commit outcomes |
| False or misleading anomaly | Frozen input/policy snapshot, exact integer calculation, human review; boundary/missing-input tests | Seasonal/quality differences can create false positives; missing real-world data creates false negatives |
| Clock manipulation | Proposal time is informational; order/version drives transitions; reject client timestamp authority | No independently certified event time |
| Peer/outage or stale UI | Required endorsement fail closed; pending status; valid-event checkpoint/rebuild tests | No HA on one host; PDC loss if every authorized copy is lost |
| Admin collusion or malicious chaincode upgrade | Separate identities, explicit approvals and pinned artifacts; inspect deployment evidence | Pilot does not resist colluding infrastructure operators |

No sensitive payloads in logs, exceptions, events or URLs. Include safe batch/operation/transaction identifiers, submitting MSP and error code. Rate-limit consumer queries and cap payloads; default maximum normalized document 64 KiB and attachment 5 MiB. Default deny unknown fields, media types and document schemas. Arbitrary remote attachment URLs are not fetched.
