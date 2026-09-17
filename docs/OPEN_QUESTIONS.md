# AgroChain — evidence, assumptions and open questions

Status: Stage 1 assumptions, 2026-09-16, with Stage 2 topology revision on 2026-09-17. The repository facts below describe the historical Stage 1 inspection; current network evidence is in [STAGE2_REPORT.md](STAGE2_REPORT.md). Team answers may require a documented contract change.

## Historical Stage 1 repository facts

- A complete recursive inspection of `/home/mert/Projects/AgroChain`, including hidden entries, found only `AGENTS.md` before this work.
- No supplied presentation, README, previous architecture, code, test suite, API-access proof, measured performance results or Git metadata existed in that directory.
- `AGENTS.md` requires eight sequential stages, Spring Boot, deterministic Fabric chaincode, commercial privacy, institutional simulators and explainable anomaly detection.
- The Stage 1 task explicitly added HKS, a consumer trace and three scenarios, and prohibited beginning Stage 2 at that time. The subsequent Stage 2 request authorizes only network provisioning and verification.
- Technical references support Fabric PDCs, Raft ordering and client-originated proposal timestamps; they do not prove this project implements them.

## Adopted assumptions, not verified facts

| ID | Stage 1 default | Consequence |
| --- | --- | --- |
| A1 | No live institutional access; all four sources signed by local simulator identities | Proves provenance of synthetic assertions only; no ministry integration claim. |
| A2 | One organization per role and one retailer; all identities/records synthetic | Three fixed PDCs on one channel are sufficient; does not prove isolation among competing retailers within one MSP. |
| A3 | Logistics has custody only; retailer acquires ownership on acknowledged receipt | Purchase invoice represents agreed terms before delivery, not proof of payment or legal transfer. |
| A4 | One whole-batch bulk lot → one retail lot; no quantity loss or splitting | Demo is tractable; real trade discrepancies block workflow and need a future model. |
| A5 | TRY/kg, tax-exclusive comparable rates, no discounts; freight contextual | Score is a price-increase review signal, not net margin or legally excessive profit. |
| A6 | One-host demo, four peers, three-node Raft in separate OrdererMSP, agrochannel; retailer and regulator endorse future business writes | Stage 2 request supersedes the single-orderer default. Crash fault tolerance for one ordering node, no host HA, Byzantine protection or independent ordering operators. |
| A7 | 50% strict-greater threshold; immutable `price-increase-v1` policy | 50% itself is no signal; changing meaning requires a new configuration version. |
| A8 | Source keys configured at bootstrap, no live rotation; synthetic data retained until explicit reset | Local key compromise/reset and real retention policies remain operational design work. |
| A9 | Baseline Linux x86-64, 16 GB RAM, 20 GB free disk; Go chaincode and Java backend in future stages | Stage 2 pins Fabric 2.5.15 and verifies the network on local Arch Linux; hardware baseline is not a measured minimum. |
| A10 | “Gıda-TL” is only a possible historical proposal label | No evidence supports treating it as a component; all deliverables use AgroChain. |

## Questions genuinely requiring team input

Technical defaults such as canonical JSON, integer money and excluding Kafka are already decided here; they are not disguised requests for approval.

| Priority / question / decision owner | Recommended default | Consequence or alternative; needed by |
| --- | --- | --- |
| **1 — Who independently controls the four organizations, especially regulator keys and endorsement availability?** Team lead + intended demo participants | Four distinct demo identities on one laptop, explicit shared-host trust notice; Retailer + Regulator endorse writes | Without independent operators this proves logical role separation only. A retailer can stall review commits. If independent governance is a judging requirement, assign operators and revise infrastructure/policy; before Stage 2. |
| **2 — Can the team defend the 1:1 lot, zero-loss and ownership-on-receipt model to domain reviewers?** Product/domain lead | Keep the one-whole-batch model and reject any mismatch | Real shipment shrinkage, partial delivery, returned goods or many-to-many invoices need new conservation and dispute flows, not rounding. Confirm before Stage 3. |
| **3 — What data and roles may be shown publicly or to the regulator?** Data owner + competition contact | Synthetic data only; consumer gets provenance, not price/score; regulator reads all three collections | Real company names, invoice data or PII require consent/access/retention review. Publishing anomaly labels could imply accusations; before Stage 4 and any external demonstration. |
| **4 — What source material or institutional relationship exists outside this directory?** Team lead + institutional contact | No access presumed; local signed CKS/EFATURA/HKS/UETDS simulators; obtain original presentation only if available | A presentation may reveal conflicting promises. Real access needs verifiable authorization, interface specification, sandbox and provenance review; must not silently replace simulators. Before Stage 5, nonblocking if unavailable. |
| **5 — Is gross unit-price increase the intended differentiator, and are the selected scenarios defensible?** Product/domain lead | 20→28 and 20→37 TRY/kg, 50% pilot parameter, freight shown separately, human review | No market calibration, legal cap or net-margin meaning exists. A cost-inclusive rule would require new inputs/formula/policy version and revised examples; before Stage 6. |
| 6 — Which machine, deadline and offline constraints apply at the venue? Team lead + demo operator | Linux x86-64 baseline, predownloaded pinned dependencies, one host, rehearsed local video fallback | Unknown hardware may prevent four-peer demo. Validate resources early; reduce optional polish rather than core security; before Stage 2 procurement/setup and Stage 8 rehearsal. |
| 7 — Should confidential synthetic records survive beyond the competition, and who handles compromised simulator/operator keys? Data owner + demo operator | Retain to explicit reset; no production data; destroy only named demo artifacts with documented backup decision | `blockToLive=0` is indefinite on that ledger, and reset does not erase third-party copies. Real retention/rotation needs separate governance; before Stage 4 operations guide. |

## Five highest-risk unresolved issues

1. **Governance and compromise concentration:** separate MSPs do not isolate keys from the shared host/backend; retailer-required endorsement can block regulator review. Demonstrate logical controls and state the limitation, or revise topology with team input.
2. **Physical/oracle truth:** signed simulator values and matching hashes cannot prove the actual tomatoes, quantities or invoice economics. Missing original presentation and real integration evidence limit what can be asserted to a jury.
3. **Domain representativeness:** whole-lot/no-loss/no-return assumptions may be too narrow for the intended story. Expand only after explicit domain confirmation and conservation design.
4. **Commercial disclosure and access governance:** collection members receive plaintext; source metadata and public transaction history still leak relationships. Confirm all demonstration data are synthetic and who is entitled to audit.
5. **Anomaly interpretation and calibration:** the 50% parameter is unvalidated and the rule ignores many market factors. Show exact arithmetic and human review, and prohibit legal or fraud-detection guarantees.

Hardware/version reproducibility is an additional delivery risk, addressed by the Stage 2 gate rather than invented benchmark numbers.

## Deferred / Future Stage

Real institutional connectors and native signature validation; inter-company identity operations; production privacy/legal assessment; multiple retailers and dynamic collections; partial quantities, losses, returns and disputes; updated prices and invoice corrections; statistical training with labeled data; fault tolerance across hosts; measured scale testing; governed key/policy rotation. These are not part of Stage 1 implementation and are not silently promised as Stage 8 capabilities.
