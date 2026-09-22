# Stage 4 collections

The single authoritative configuration is
[`network/config/collections.json`](../../../network/config/collections.json).
Stage 4 writes and reads all three collections with mandatory MSP/role checks.
Do not duplicate or weaken that configuration here.

tradePrivate holds purchase openings; freightPrivate holds freight openings/costs;
retailAuditPrivate holds retail reports and later analysis/review. All require
Retailer + Regulator endorsement. Membership configuration alone proves no privacy.
Salted records, source verification, role enforcement and real member/nonmember
access tests are implemented. See the [privacy runbook](../../../docs/STAGE4_PRIVACY.md).
