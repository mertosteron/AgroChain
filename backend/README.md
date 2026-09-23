# AgroChain Stage 5 backend

Spring Boot API → Java Fabric Gateway → existing Stage 4 chaincode. Four in-process
institutional adapters produce **SIMULATED** signed documents. Ledger submissions,
endorsements and PDC reads are real Fabric operations. No government API, anomaly
classification or UI is implemented here.

## Setup and run

Linux x86-64; follow the [network guide](../network/README.md) first. The backend
uses repository-local Temurin **21.0.12.1+1** and Maven **3.9.11**. Their archives
and checksums are pinned in `scripts/toolchain.py`; no host Java replacement occurs.
Network scripts require supported Python >=3.10 with PyYAML/cryptography. On the
development Arch host use `PATH=/usr/bin:/bin:$PATH` before Make commands.

For a new environment:

```bash
make backend-prerequisites
make chaincode-prerequisites chaincode-deps
make bootstrap chaincode-deploy
make backend-prepare
make privacy-integration
make stage5-check
make backend-run
```

`privacy-integration` is the existing Stage 4 acceptance initializer: it registers
the four local public keys once, checks privacy, creates synthetic records and
restarts the network without deleting state. An already bootstrapped channel does
not need it on each backend start. Do not replace its persisted source keys.
For an existing compatible Stage 4 network: `make network-up backend-prepare
stage5-check`. Run the acceptance suite without another backend/writer on its
test port. Existing records are preserved; every test run uses new operation IDs.

`backend-prepare` creates development role certificates and random bearer tokens
without printing them. The default API is **http://127.0.0.1:8080**. Use the relevant
token from the protected `network/runtime/backend/tokens.json` as `Authorization:
Bearer <token>`. Tokens map to fixed organization/role signers; X-MSP/X-Role request
headers grant no authority. Treat this token file as a credential, never a fixture
to commit or include in logs. `make backend-run` runs in the foreground; Ctrl-C
stops it while preserving SQLite, attachments and Fabric data.

First dependency download requires internet. Subsequent builds can use the local
Maven cache (`bash backend/mvnw -o package`). The application and its simulated
sources require no external internet during the demonstration.

## Supported API

All paths start with `/api/v1`. Every response carries `X-Source-Mode: SIMULATED`,
`X-Blockchain-Mode: FABRIC` and `Cache-Control: no-store`. Authentication is required
except for the health and consumer-safe public lot queries.

| Method/path | Permission and behavior |
| --- | --- |
| GET `/health` | Process liveness and projection checkpoint; not Fabric readiness |
| POST `/batches` | Producer — CreateBatch with ÇKS eligibility |
| POST `/batches/{id}/pickup-offers` | Producer — purchase e-Fatura, HKS and U-ETDS |
| POST `/batches/{id}/pickup-acceptances` | Carrier — accept pickup |
| POST `/batches/{id}/freight-costs` | Carrier — freight e-Fatura |
| POST `/batches/{id}/delivery-offers` | Carrier — offer delivery |
| POST `/batches/{id}/delivery-acceptances` | Retailer — accept ownership/custody |
| POST `/batches/{id}/retail-reports` | Retailer — private offered price |
| GET `/batches/{id}` | Authenticated shared batch query |
| GET `/batches/{id}/history?bookmark=...` | Authenticated receipt history, pages of 100 |
| GET `/batches/{id}/commercial?kind=purchase\|freight\|retail` | Collection/role-filtered private query |
| GET `/documents/{id}` | Authorized committed original invoice bytes |
| POST `/documents/{id}/verify` | Authorized exact-byte integrity check |
| GET `/operations/{id}` | Status scoped to the authenticated actor |
| GET `/public/lots/{id}` | Explicit public projection; no price, salt, document ID or anomaly |

Private reads are evaluate-only. No API accepts an arbitrary chaincode function,
collection/key, signer/MSP or Gateway destination. All transient-bearing proposals
go through the Retailer Gateway with explicit Retailer + Regulator endorsers,
while retaining the business actor's signature. The backend does not submit
GetPurchase/GetFreightCost/GetRetailReport responses to the ordering service.

Mutation body has exactly `command`, `scenario` and `privateInput`. Scenario is
`NORMAL` or `SUSPICIOUS`; these are synthetic input labels, not an anomaly result.
The `Idempotency-Key` header must equal the command's valid opaque operation ID.
The client chooses that ID once and reuses it after timeouts. Example:

```json
{
  "command": {
    "schemaVersion": "agrochain.command.v1",
    "operationId": "OP-EXAMPLE0001",
    "batchId": "BAT-EXAMPLE0001",
    "expectedVersion": 0,
    "command": "CreateBatch",
    "payload": {
      "productCode": "TOMATO", "gradeCode": "STANDARD", "quantityGrams": 100000,
      "originRegionCode": "07", "harvestDate": "2026-09-15",
      "logisticsMsp": "LogisticsMSP", "intendedRetailerMsp": "RetailerMSP"
    }
  },
  "scenario": "NORMAL",
  "privateInput": {}
}
```

Subsequent public payloads are the unchanged [chaincode schemas](../chaincode/agrochain/README.md).
`expectedVersion` increments on each command, including freight. `privateInput`
is empty except for ReportRetailPrice, where it contains exactly
`offeredPriceKurusPerKg`, `currency: TRY`, `taxBasis: EXCLUDING_TAX` and
`reportedAt` as UTC milliseconds. The backend adds a private random salt. Do not put
price fields in `command.payload`. The integrity body is
`{"originalBase64":"<base64 of exact invoice bytes>"}`. Maximum attachment 5 MiB,
maximum HTTP body 8 MiB. The original is never put on the public ledger.

## Commit, retry and recovery

- New confirmed VALID commit: HTTP 201 and `agrochain.operation-status.v1` with
  status COMMITTED and its receipt. An identical recovered request returns 200
  with the same receipt. Same operation ID with different input returns 409.
- Unknown outcome: HTTP 202, status SUBMITTED_UNKNOWN, known txId and statusUrl;
  no success receipt. Retain the operation ID and query its status or retry the
  exact same request. Never generate a new ID merely because the request timed out.
- Rejections use `agrochain.error.v1` with a safe code/message, correlationId and
  retryable flag: 400 malformed values, 401 unauthenticated, 403 unauthorized,
  404 inaccessible/missing record, 409 state/version/idempotency conflict,
  422 evidence failure, 503 dependency/private-data unavailability, 502 other
  explicitly invalid Fabric commit, 500 internal error. No exception message,
  certificate, token or PDC body appears in an error response.

SQLite stages the request before submission, caches verified evidence and stores
the endorsed transaction bytes/ID before sending. A two-second recovery worker
checks unresolved operations. It first recovers immutable receipts, then checks
the persisted transaction's VALID/invalid commit status. Only an unresolved
transaction may be resent with its original ID/bytes. An explicit invalid commit
is terminal. Unavailable dependencies before endorsement return 503; the caller
can retry the same request after the dependency returns.

Evidence metadata becomes committed in the same SQLite transaction as its receipt.
A crash after the Fabric commit but before off-chain promotion is recovered by
receipt reconciliation. Unresolved files are retained; automatic file expiry is
deliberately not implemented. A committed attachment cannot be read through the
API until that promotion is complete.

## Simulator and projection boundaries

`InstitutionalAdapter` is the replaceable boundary. `Simulators` implements CKS,
EFATURA (purchase and freight), HKS and UETDS inside the same process. Cached
requests return the same document ID, nonce, body, signature and original bytes
across retry/restart. The frozen fixture is 100 kg STANDARD tomatoes from region
07: purchase 2,000 kuruş/kg, purchase total 200,000 kuruş and freight 20,000 kuruş.
The acceptance workflows declare retail 2,800 and 3,700 kuruş/kg respectively.
These values are not prices fetched from a government system.

`EvidenceService` verifies schema, committed public-key trust, Ed25519 signature,
salted commitment, batch/operation/party/quantity bindings and exact original-byte
digest before endorsement. The chaincode independently repeats consensus-critical
validation. Simulator signing keys are kept separate from the ledger's verifier
trust. Both are controlled by the same local pilot operator; signatures do not
establish independent institutional or physical truth.

The event consumer reads only valid committed chaincode events, atomically stores
transaction deduplication/checkpoint and an explicit public DTO, and resumes from
the last event block inclusively. Replay repairs committed-but-unprojected batches.
Projection fields follow the [public allowlist](../docs/DATA_CONTRACTS.md); no
private query serializer is used. `X-Projection-Consistency: EVENTUAL`, the response
checkpoint header and `asOfBlock` disclose the view's observation point. An absent
lot returns 404; there is no fallback to private state. This supplies Stage 5's
read-model foundation, not a completed Stage 6 QR interface.

## Configuration and validation

| Variable | Default/purpose |
| --- | --- |
| AGROCHAIN_ROOT | Repository root selected by Make |
| AGROCHAIN_BACKEND_DATA | `network/runtime/backend`; SQLite and attachments |
| AGROCHAIN_TOKEN_FILE | `<data>/tokens.json`; fixed development token-to-actor map |
| AGROCHAIN_SOURCE_KEYS | `network/runtime/source-keys`; local simulator private keys |
| AGROCHAIN_GATEWAY_ENDPOINT | `localhost:9051`; Retailer Gateway, configured Retailer TLS authority/root |
| AGROCHAIN_CHANNEL / AGROCHAIN_CHAINCODE | `agrochannel` / `agrochain` |
| AGROCHAIN_API_PORT | 8080; loopback only |
| AGROCHAIN_DISABLED_SIMULATORS | Comma-separated CKS, EFATURA, HKS, UETDS for explicit outage tests |
| AGROCHAIN_TEST_API_PORT | 18080; acceptance server |

Use `make backend-test` for isolated tests, `make backend-integration` for actual
HTTP/network workflows and `make backend-gateway-test` for actual valid/MVCC-invalid
transaction recovery using a reconstructed Gateway. `make stage5-check` builds,
checks the deployed chaincode, and runs both live suites. The Gateway live test
is intentionally skipped in ordinary unit tests and explicitly enabled in its
separate target. Acceptance artifacts remain under ignored runtime directories.

## Known limitations

This is one process, one SQLite database and one local Fabric environment. Writes
are serialized for predictable recovery. Files/database are access-restricted,
not encrypted against the host administrator. Pilot bearer credentials have no
interactive login/rotation flow; HTTP is loopback-only, not a remote deployment.
The Retailer Gateway is a single availability dependency. Event-loop retries retain
their checkpoint; no distributed queue or high availability is claimed. Retained
failed/pending staged files need an operator-reviewed cleanup policy before larger
deployments. Stage 6 anomaly/review endpoints and UI are documented in
[the UI guide](../docs/STAGE6_UI.md). The UI's manual token entry is not an identity
provider or production authentication service.

Dependency choices follow [Spring Boot 3.5 requirements](https://docs.spring.io/spring-boot/3.5/system-requirements.html)
and the [Fabric Gateway evaluate/endorse/submit/commit model](https://hyperledger-fabric.readthedocs.io/en/latest/gateway.html).
The pinned Gateway POM requires the aligned protobuf 4.33.4 and gRPC 1.78.0 runtimes;
the build declares their BOMs to prevent transitive downgrade.
