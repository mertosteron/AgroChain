# Stage 4 — privacy runtime and reproduction

Implemented on 2026-09-21. Release 0.2.1; existing local channel sequence 5,
fresh channels sequence 1. See [acceptance report](STAGE4_REPORT.md).

## Reproduce

Use the [network prerequisites](../network/README.md), Python >=3.10 with
`cryptography` (local run: Python 3.14.7/cryptography 50.0.1), and Java >=17 with
Ed25519 support (local run: OpenJDK 26.0.2.1). No Java dependency download is needed.
CI installs pinned cryptography 46.0.3; its hosted execution is not claimed here.

```bash
make bootstrap
make chaincode-prerequisites chaincode-deps
make chaincode-deploy
make privacy-prepare
make stage4-check
```

On this workstation select the supported Python/JDK explicitly:

```bash
PATH=/usr/bin:/bin:$PATH make privacy-prepare
PATH=/usr/bin:/bin:$PATH make stage4-check JAVA=/usr/lib/jvm/java-26-openjdk/bin/java
```

For an existing older deployment use the next explicit sequence and a new semantic
version via `chaincode-upgrade`; do not use `chaincode-deploy` to silently upgrade.
The executed final upgrade was version 0.2.1/sequence 5. Do not repeat it on a channel
already at that version. Routine restart is `make network-down network-up`.
Never delete runtime keys while preserving a bootstrapped ledger.

The suite uses synthetic documents and random IDs/salts to permit repeat runs
without destructive resets. It adds a seven-command batch, a concurrent creation
pair, health probes and an intentionally invalid endorsement transaction. It then
restarts the network and verifies existing public/private data. Run it with no
other writers: block-height assertions and the concurrent race assume an isolated
local acceptance network. Business amounts are repeatable; cryptographic bytes are
intentionally random. It is not the final Stage 8 demonstration seed command.

## Identity and bootstrap

`network/scripts/privacy-prepare.py` issues development role certificates signed
by the existing generated organization CAs. It creates ignored, access-restricted
`network/runtime/identities` and `network/runtime/source-keys`. It preserves existing
identities/keys on subsequent runs. No real institutional keys are used.

All business/shared-query calls require a valid MSP/role combination from the
certificate's `agrochain.role` attribute. Attribute-less User1 can call Health only.
Regulator `admin` may Bootstrap once; admin is not a business role. Bootstrap takes
one exact `agrochain.bootstrap.v1` JSON object containing policyId, thresholdBps and
four source-key records sorted by sourceSystem. Duplicate system or issuer/key
registry slots fail. Source systems are CKS, EFATURA, HKS and UETDS. Key trust and
the policy are immutable in this pilot. The acceptance suite bootstraps only when
configuration is absent; otherwise local public keys must match committed trust.
Bootstrap requires the normal Retailer + Regulator endorsements.

Health returns `evidenceVerification=REQUIRED_STAGE_4`; `writesEnabled` becomes true
after bootstrap. This signals availability of the verified path, not permission to
skip evidence. The contract has no runtime verifier selector or external HTTP calls.

## Confidential data and query interface

| Collection | Intended readers | Stored records |
| --- | --- | --- |
| tradePrivate | Producer/producer, Retailer/retailer, Regulator/auditor, reviewer or oracle | Purchase normalized body + 32-byte secret salt |
| freightPrivate | Logistics/carrier, Retailer/retailer, Regulator/auditor, reviewer or oracle | Freight invoice opening and salted cost record |
| retailAuditPrivate | Retailer/retailer, Regulator/auditor, reviewer or oracle | Salted retail price report |

Regulator/public-reader has shared queries only. Membership is enforced by Fabric
collection policies and chaincode identity checks. The endorsing peers are Retailer
and Regulator, both members of every collection. Nonmember peers are also tested
with an otherwise authorized Regulator identity: the private bytes remain unavailable.

| Function | Argument | Result |
| --- | --- | --- |
| GetConfiguration | none | Public immutable bootstrap configuration |
| GetDocument | DOC ID | Committed public signed envelope |
| GetPurchase | BAT ID | Purchase envelope, body and salt for authorized readers |
| GetFreightCost | BAT ID | Salted freight cost record for authorized readers |
| GetRetailReport | BAT ID | Salted price report for authorized readers |
| VerifyDocument | DOC ID + transient opening | Committed signature and salted normalized-body match |

Use private queries through evaluate/query only. Do not order a transaction whose
response contains private data: Fabric can retain submitted transaction response
payloads. The Stage 5 Gateway must expose separate evaluate and submit allowlists.
An authorized recipient can copy/disclose data it legitimately receives; PDC does
not protect against that recipient or the shared machine's administrator.

`VerifyDocument` accepts exactly transient `opening={body,saltHex}`. Freight invoice
verification uses freight membership; other supported openings use trade membership.
It returns `agrochain.opening-result.v1` with documentId, sourceMode=SIMULATED,
commitmentMatched=true and signatureVerified=true only after success. It does not
claim to have inspected attachment bytes. There is no arbitrary PDC-key query.
Missing private bytes return PRIVATE_DATA_UNAVAILABLE, never a default zero amount.

CLI helper: set `AGROCHAIN_ROLE` to a provisioned role and supply transient values
as a base64-valued JSON map file via `AGROCHAIN_TRANSIENT_FILE`. The file and CLI
arguments are local test tooling; protect them from other local users. Private
business fields never belong in the public command argument. Query peer override
`AGROCHAIN_QUERY_PEER` exists for acceptance inspection, not authorization bypass.

## Integrity, replay and storage

The domain-separated salted SHA-256 commitment and Ed25519 signature format follow
[data contracts](DATA_CONTRACTS.md). Keys are selected only from committed trust,
never caller-provided keys. Complete header signatures bind the system, type,
document, batch, operation, nonce and commitment. The verifier checks the normalized
body against the intended parties/product/quantity/currency and exact integer money.
Source document IDs, nonces and public document IDs cannot be reused. Commands also
retain MSP-scoped operation IDs and batch versions. Fabric MVCC rejects competing
read versions at commit; public and private changes share that atomic transaction.

Invoice bodies/salts stay in PDC; source envelopes and opaque replay indices stay
shared. CKS/HKS/UETDS openings are transient inputs, not on-chain archives. Their
controlled archive is Stage 5 work. Private cost/report records include independent
random 32-byte salts; confidentiality is not based on an unsalted price hash.
Randomness is generated outside chaincode. Chaincode validates encoding/length,
not the statistical entropy of client-supplied salt.

Canonicalization is the schema-restricted JCS subset: printable ASCII strings,
integers, booleans, exact objects and bounded arrays. No fractional/exponent money,
negative zero, duplicate keys, nulls or arbitrary Unicode/free text is accepted.
The documented commitment and Ed25519 signature reproduce in both Go and Java;
this is not a general-purpose RFC 8785 library. Inputs are capped at 64 KiB per
parsed JSON value, depth 8 and bounded field/array sizes.

`test/integration/evidence_support.py::verify_original` is the executable off-chain
attachment verifier: it uses the committed envelope and trusted configuration,
rechecks signature/commitment, then hashes exact original bytes (maximum 5 MiB).
The original XML passes and an added newline fails DOCUMENT_HASH_MISMATCH.
These are integrity checks of signed assertions, not proof of physical crop truth
or semantic government validation. Full document storage, authentication, HTTP error
mapping and integration with Spring Boot remain Stage 5.

## Limits and next stage

This is one host with development CAs, shared administrative trust and server-only
peer-to-chaincode TLS. No source-key rotation/revocation workflow, nationwide
capacity, actual government integration or production security certification is
claimed. Stage 5 should reuse these schemas/PDCs and mandatory roles, implement
Gateway evaluate/submit allowlists, signed institutional simulators and controlled
off-chain documents. Anomaly calculation, review actions and consumer UI remain
Stage 6. The bootstrap threshold is only a pilot parameter; no anomaly result is
computed by Stage 4.
