# AgroChain — Stage 1 completion report

## Stage

Stage 1 — Scope and Technical Contract. Specification version 1.0, 2026-09-16. PASS means the documentation gate passed, not that the pilot is implemented or its future runtime tests have passed.

## Completed

- Inspected all files, including hidden entries: only `AGENTS.md` existed. No supplied presentation, README, prior architecture, code, tests, network configuration or institutional-access evidence was available.
- Defined one producer → logistics custody → retailer ownership → regulator review route and consumer trace, with explicit whole-batch conservation and two-party handoffs.
- Specified four signed institutional simulators, exact monetary units, replay protection, versioned records, a reproducible commitment/signature vector, PDC membership and feasible private endorsement routing.
- Separated shared ledger fields, private commercial data and original off-chain documents; distinguished record authenticity from physical truth and review signals from legal findings.
- Defined three demo scenarios, stage gates, minimum shippable scope and team questions with recommended defaults. Root architecture is a pointer to satisfy the repository's required path.
- Used the context-mode skill to inspect and validate documentation with focused outputs; it did not change product scope or authorize implementation.

## Changed Files

All nine files are newly created; no existing file was changed.

| File | Purpose |
| --- | --- |
| [PRODUCT_SCOPE.md](PRODUCT_SCOPE.md) | Evidence inventory, scope, scenarios, scoring rule, measurable targets and three-minute flow |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Components, adapters, data flow, Gateway/commit recovery, deterministic boundaries and topology |
| [DATA_CONTRACTS.md](DATA_CONTRACTS.md) | Versioned field definitions, examples, canonical signing/commitments, storage keys, errors and consumer allowlist |
| [SECURITY_AND_PRIVACY.md](SECURITY_AND_PRIVACY.md) | MSP/operation matrix, PDC/endorsement policies, field classification and threat model |
| [STATE_MACHINE.md](STATE_MACHINE.md) | Product, custody and review transitions, invariants, events and failure cases |
| [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) | Stages 2–8 dependencies, verification gates and minimum delivery path |
| [OPEN_QUESTIONS.md](OPEN_QUESTIONS.md) | Verified facts, explicit assumptions, team decisions and risk register |
| [../ARCHITECTURE.md](../ARCHITECTURE.md) | Root-level entry point to the authoritative docs architecture |
| [STAGE1_REPORT.md](STAGE1_REPORT.md) | This completion report and reproducible documentation validation |

`AGENTS.md` remains unchanged. No chaincode, network, UI, backend, dependency manifest or runtime test suite was created. Git diff/status is unavailable because the supplied directory has no Git repository; this is an environment observation, not a passing test.

## Tests Executed

- Recursive repository inventory — completed; only the original `AGENTS.md` was present before writing.
- Initial JSON/link/vector validation through Node.js in context-mode — PASS: seven requested documents, nine parseable JSON examples, 21 local links at that revision, 288-byte canonical vector, matching commitment, attachment digest and valid Ed25519 signature.
- Final `node` documentation-check program below, executed through context-mode — PASS: nine new documents, nine JSON examples, 32 local links; example structure/IDs/units, cross-example quantities and references, signature/commitment positives and tamper negatives, seven formula/boundary cases, collection/endorser compatibility, command coverage, Markdown fence/whitespace checks.
- Manual contract review — PASS for Stage 1: caller versus endorser distinction; owner versus custodian; private transients and hashes; source reuse versus later references; quantity conservation; retail offer versus sale; review state versus lifecycle; pending commit versus success; required freight before receipt.
- Fabric, chaincode, API, PDC access and E2E runtime tests — **NOT RUN**: no runtime exists. Cross-language Go/Java JCS conformance is deferred to Stage 4. Formula checks here validate documented examples, not an implemented anomaly engine.

### Reproduce the documentation checks

Run from the repository root with Node.js available. This inline documentation check is not application implementation. Its small canonicalizer is restricted to the ASCII-key/integer test vector and is not a production RFC 8785 library.

```javascript
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const root = process.cwd();
const names = ['PRODUCT_SCOPE', 'ARCHITECTURE', 'DATA_CONTRACTS',
  'SECURITY_AND_PRIVACY', 'STATE_MACHINE', 'IMPLEMENTATION_PLAN', 'OPEN_QUESTIONS'];
const files = names.map(n => `docs/${n}.md`).concat('ARCHITECTURE.md', 'docs/STAGE1_REPORT.md');
const texts = Object.fromEntries(files.map(f => [f, fs.readFileSync(path.join(root, f), 'utf8')]));
let links = 0;
let jsonCount = 0;
const examples = [];
for (const [file, text] of Object.entries(texts)) {
  assert.equal((text.match(/^```/gm) || []).length % 2, 0, `unclosed fence: ${file}`);
  assert(!/[\t ]+$/m.test(text), `trailing whitespace: ${file}`);
  assert(text.endsWith('\n'), `missing final newline: ${file}`);
  for (const match of text.matchAll(/\]\(([^)]+)\)/g)) {
    if (/^(https?:|#)/.test(match[1])) continue;
    assert(fs.existsSync(path.resolve(root, path.dirname(file), match[1].split('#')[0])), match[1]);
    links++;
  }
  for (const match of text.matchAll(/```json\n([\s\S]*?)\n```/g)) {
    const value = JSON.parse(match[1]);
    jsonCount++;
    if (file === 'docs/DATA_CONTRACTS.md') examples.push(value);
  }
}
assert.equal(jsonCount, 9);
function visit(x) {
  if (typeof x === 'number') assert(Number.isSafeInteger(x));
  if (typeof x === 'string' && /^(BAT|LOT|DOC|TRF|CST|RPT|ANM|OP|CFG)-/.test(x))
    assert(/^(BAT|LOT|DOC|TRF|CST|RPT|ANM|OP|CFG)-[A-Z0-9]{8,32}$/.test(x), x);
  if (x && typeof x === 'object') for (const [k, v] of Object.entries(x)) {
    if (k === 'quantityGrams') assert(v > 0 && v <= 100000000);
    if (/SaltHex$/.test(k)) assert(/^[0-9a-f]{64}$/.test(v));
    if (/TxId$/.test(k) || k === 'txId') assert(/^[0-9a-f]{64}$/.test(v));
    visit(v);
  }
}
examples.forEach(visit);
const byType = type => examples.find(x => x.schemaVersion === `agrochain.${type}.v1`);
const batch = byType('batch'), transfer = byType('custody-transfer');
const cost = byType('cost'), report = byType('retail-price-report'), anomaly = byType('anomaly-result');
for (const x of [transfer, cost, report, anomaly]) {
  assert.equal(x.batchId, batch.batchId);
  assert.equal(x.quantityGrams, batch.quantityGrams);
}
assert.equal(report.retailLotId, batch.retailLotId);
assert.equal(report.freightCostId, cost.costId);
assert.equal(anomaly.reportId, report.reportId);
assert.equal(anomaly.policyId, report.policyId);
assert.equal(anomaly.retailPriceKurusPerKg, report.offeredPriceKurusPerKg);
assert.equal(anomaly.freightTotalKurus, cost.totalKurus);
const envelope = examples.find(x => x.header), body = examples.find(x => x.sellerMsp);
assert.equal(report.purchaseDocumentId, envelope.header.documentId);
assert.equal(body.priceKurusPerKg * body.quantityGrams, body.totalKurus * 1000);
assert.equal(anomaly.purchasePriceKurusPerKg, body.priceKurusPerKg);
const canonical = x => Array.isArray(x) ? '[' + x.map(canonical).join(',') + ']'
  : x && typeof x === 'object' ? '{' + Object.keys(x).sort().map(k => JSON.stringify(k) + ':' + canonical(x[k])).join(',') + '}'
  : JSON.stringify(x);
const salt = Buffer.from('202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f', 'hex');
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
function commitment(value, secret = salt) {
  const bytes = Buffer.from(canonical(value)), len = Buffer.alloc(8);
  len.writeBigUInt64BE(BigInt(bytes.length));
  return digest(Buffer.concat([Buffer.from('AgroChain/document/v1\0'), secret, len, bytes]));
}
assert.equal(Buffer.byteLength(canonical(body)), 288);
assert.equal(commitment(body), envelope.header.commitment);
assert.notEqual(commitment({...body, totalKurus: 200001}), envelope.header.commitment);
assert.notEqual(commitment(body, Buffer.alloc(32)), envelope.header.commitment);
assert.equal(commitment(Object.fromEntries(Object.entries(body).reverse())), envelope.header.commitment);
const key = crypto.createPublicKey({format: 'der', type: 'spki', key: Buffer.concat([
  Buffer.from('302a300506032b6570032100', 'hex'),
  Buffer.from('njeHGEyz0EF6HP6iEDUHPftD5gwevFdrlCcOnJgTFkY', 'base64url')])});
const signature = Buffer.from(envelope.signature, 'base64url');
assert.equal(signature.length, 64);
const verify = h => crypto.verify(null, Buffer.concat([
  Buffer.from('AgroChain/source-signature/v1\0'), Buffer.from(canonical(h))]), key, signature);
assert(verify(envelope.header));
assert(!verify({...envelope.header, batchId: 'BAT-TAMPER01'}));
const original = '<invoice id="SIM-001">200000</invoice>';
assert.equal(digest(original), body.attachmentDigest);
assert.notEqual(digest(original + '\n'), body.attachmentDigest);
function score(p, r, t = 5000) {
  assert(p > 0 && r >= 0);
  const d = (BigInt(r) - BigInt(p)) * 10000n, P = BigInt(p);
  const floor = d / P - (d < 0n && d % P !== 0n ? 1n : 0n);
  return [Number(floor), d > P * BigInt(t)];
}
assert.deepEqual(score(2000, 2800), [4000, false]);
assert.deepEqual(score(2000, 3700), [8500, true]);
assert.deepEqual(score(2000, 3000), [5000, false]);
assert.deepEqual(score(2000, 3001), [5005, true]);
assert.deepEqual(score(2000, 1500), [-2500, false]);
assert.deepEqual(score(3000, 1999), [-3337, false]);
assert.throws(() => score(0, 3700));
const security = texts['docs/SECURITY_AND_PRIVACY.md'];
const policies = [...security.matchAll(/OR\('([A-Za-z]+MSP)\.member'[^`]+/g)].map(m => m[0]);
assert.equal(policies.length, 3);
for (const p of policies) assert(p.includes('RetailerMSP.member') && p.includes('RegulatorMSP.member'));
assert(security.includes("AND('RetailerMSP.peer','RegulatorMSP.peer')"));
const state = texts['docs/STATE_MACHINE.md'], contracts = texts['docs/DATA_CONTRACTS.md'];
for (const command of ['CreateBatch', 'OfferPickup', 'AcceptPickup', 'RecordFreightCost',
  'OfferDelivery', 'AcceptDelivery', 'ReportRetailPrice', 'EvaluatePrice', 'OpenReview', 'ResolveReview']) {
  assert(state.includes(command) && contracts.includes(command), command);
}
console.log(`PASS: ${files.length} new documents; ${jsonCount} JSON examples; ${links} local links; crypto/tamper checks; 7 scoring cases; PDC/command consistency.`);
```

Execute the block with `node` (for example, paste it into a Node stdin invocation); no package installation or network is required for these checks. Relative-link validation is local only; it does not certify external website availability. Structural checks are intentionally narrower than future executable schemas and security tests.

## Acceptance Criteria

- [x] Organization/MSP roles, submitting identities, endorsements and read/audit permissions are explicit.
- [x] Product lifecycle, ownership/custody transitions, quantities, failure behavior and separate anomaly review lifecycle are explicit.
- [x] Shared ledger, consumer-public, PDC and off-chain boundaries are explicit, with field-level classification.
- [x] Backend/adapter/oracle/Gateway/chaincode responsibilities and deterministic execution restrictions are explicit.
- [x] Four signed simulator assumptions, replay controls, canonical commitments and source authenticity limits are documented.
- [x] Normal, suspicious, tampered/duplicate scenarios have concrete expected outcomes and measurable future acceptance targets.
- [x] Anomaly formula, units, parameter snapshot, boundary behavior, human review and limitations are fixed.
- [x] Source contradictions/missing evidence are recorded; core defaults are consistent; genuine team decisions remain visible.
- [x] All seven requested deliverables exist and documentation checks pass.
- [x] No Stage 2 implementation or unrelated change was made.

## Assumptions

Synthetic-only data; one organization per role; one-host demo; logistics never owns the batch; zero-loss whole-batch route; one purchase invoice/freight record/retail report; tax-exclusive TRY/kg; 50% strict-greater pilot threshold; source keys configured at bootstrap. The complete assumption register and consequences are in [OPEN_QUESTIONS.md](OPEN_QUESTIONS.md).

## Known Limitations — five highest-risk unresolved issues

1. Shared-host/backend key custody and retailer-required endorsement concentrate trust and permit stalled regulator writes.
2. Signed simulator records cannot prove physical truth; no source presentation or real institutional access evidence was supplied.
3. Whole-lot/no-loss assumptions may not match real tomato distribution, requiring domain confirmation before expansion.
4. PDC members can read plaintext and public metadata remains linkable; disclosure and audit governance need team confirmation before real data.
5. The 50% price-increase parameter lacks market calibration and legal meaning; the rule can produce false positives/negatives.

The untested machine/version baseline is an additional delivery risk. Documentation checks do not validate Fabric behavior, cross-language JCS edge cases or production security.

## Proposed live-demo flow

Preflight and seed normal/suspicious histories; retain an open suspicious review and integrity test document. In three minutes: identify the four organizations and simulator banner → inspect normal custody/provenance and 40% result → compare 85% suspicious result and commit a regulator review action → reject tampered/replayed evidence without changing state → demonstrate logistics private-data denial and consumer QR provenance. The full route must also run through the seed/E2E workflow. Detailed timing and a clearly labeled recorded fallback are specified in [PRODUCT_SCOPE.md](PRODUCT_SCOPE.md).

## Deferred to Future Stages

Stages 2–8 implementation, real-peer privacy/authorization tests, signed adapter services, backend/API/UI, cross-language canonicalization tests, E2E measurement, installation guide and final presentation. Real government access, national deployment, trained models, losses/splits/returns and independent-host resilience remain explicitly outside the minimum pilot.

## Result

**PASS — Stage 1 documentation acceptance only.** No next stage was started. Defaults remain assumptions where identified; runtime claims require the later-stage evidence gates.
