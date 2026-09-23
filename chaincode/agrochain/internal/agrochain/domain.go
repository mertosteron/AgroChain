package agrochain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type execution struct {
	Actor, Role, TxID string
	Time              TxTime
}
type engine struct {
	store    ledger
	evidence evidenceProvider
}

func (e engine) apply(c Command, x execution) (Receipt, error) {
	var receipt Receipt
	if err := authorize(x.Actor, x.Role, c.Command); err != nil {
		return receipt, err
	}
	if err := validatePayload(c); err != nil {
		return receipt, err
	}
	if !hexID.MatchString(x.TxID) {
		return receipt, fail("INTERNAL_ERROR")
	}
	s, err := strconv.ParseInt(x.Time.Seconds, 10, 64)
	if err != nil || s < 0 || strconv.FormatInt(s, 10) != x.Time.Seconds || x.Time.Nanos < 0 || x.Time.Nanos > 999999999 {
		return receipt, fail("INTERNAL_ERROR")
	}
	if err := absent(e.store, key("operation", x.Actor, c.OperationID), "DUPLICATE_TRANSACTION"); err != nil {
		return receipt, err
	}
	var b Batch
	exists, err := load(e.store, key("batch", c.BatchID), &b)
	if err != nil {
		return receipt, err
	}
	if c.Command == "CreateBatch" {
		if exists {
			return receipt, fail("BATCH_ALREADY_EXISTS")
		}
	} else if !exists {
		return receipt, fail("BATCH_NOT_FOUND")
	}
	if b.Version != c.ExpectedVersion || b.Version == 2147483647 {
		return receipt, fail("VERSION_CONFLICT")
	}
	old := b
	refs := EvidenceRefs{SchemaVersion: "agrochain.batch-evidence.v1", BatchID: c.BatchID}
	if exists {
		if _, err := load(e.store, key("batchEvidence", c.BatchID), &refs); err != nil {
			return receipt, err
		}
	}
	plan := writePlan{}
	ids := []string{c.BatchID}
	eventName := ""
	// The production provider always rejects. Unit-only providers furnish typed
	// verified outcomes. Stage 4 must replace this seam with verification AND
	// atomic private writes, not just a boolean or an environment toggle.
	proof, err := e.evidence.prepare(c, b, refs)
	if err != nil {
		return receipt, err
	}
	switch c.Command {
	case "EvaluatePrice", "OpenReview", "ResolveReview":
		var p analysisInput
		_ = json.Unmarshal(c.Payload, &p)
		ids = append(ids, p.AnomalyID)
		eventName = "ReviewUpdated"
		if c.Command == "EvaluatePrice" {
			eventName = "PriceEvaluated"
			ids = append(ids, p.ReportID)
		}
	case "CreateBatch":
		var p createInput
		_ = json.Unmarshal(c.Payload, &p)
		b = Batch{SchemaVersion: "agrochain.batch.v1", BatchID: c.BatchID, ProductCode: p.ProductCode, GradeCode: p.GradeCode, QuantityGrams: p.QuantityGrams, OriginRegionCode: p.OriginRegionCode, HarvestDate: p.HarvestDate, ProducerMSP: x.Actor, LogisticsMSP: p.LogisticsMSP, IntendedRetailerMSP: p.IntendedRetailerMSP, OwnerMSP: x.Actor, CustodianMSP: x.Actor, State: Created, CreatedTxID: x.TxID}
		if proof.Quantity != b.QuantityGrams {
			return receipt, fail("SOURCE_BINDING_MISMATCH")
		}
		if err := consumeDocument(e.store, plan, proof.CKS, c); err != nil {
			return receipt, err
		}
		refs.CKSDocumentID = proof.CKS
		ids = append(ids, proof.CKS)
		eventName = "BatchCreated"
	case "OfferPickup", "OfferDelivery":
		var p handoffInput
		_ = json.Unmarshal(c.Payload, &p)
		state, next, kind, to := Created, PickupPending, "PICKUP", b.LogisticsMSP
		if c.Command == "OfferDelivery" {
			state, next, kind, to = InTransport, DeliveryPending, "DELIVERY", b.IntendedRetailerMSP
		}
		if b.State != state {
			return receipt, fail("INVALID_STATE_TRANSITION")
		}
		if x.Actor != b.CustodianMSP || (kind == "PICKUP" && x.Actor != b.OwnerMSP) {
			return receipt, fail("UNAUTHORIZED_ORGANIZATION")
		}
		if p.QuantityGrams != b.QuantityGrams {
			return receipt, fail("INVALID_QUANTITY")
		}
		if err := absent(e.store, key("pendingTransfer", b.BatchID), "INVALID_STATE_TRANSITION"); err != nil {
			return receipt, err
		}
		if err := absent(e.store, key("transfer", p.TransferID), "TRANSFER_ALREADY_EXISTS"); err != nil {
			return receipt, err
		}
		if kind == "PICKUP" {
			if refs.CKSDocumentID == "" {
				return receipt, fail("MISSING_EVIDENCE")
			}
			if proof.Quantity != b.QuantityGrams {
				return receipt, fail("SOURCE_BINDING_MISMATCH")
			}
			if !validPurchase(proof) {
				return receipt, fail("INVALID_MONEY")
			}
			for _, id := range []string{proof.Purchase, proof.HKS, proof.Manifest} {
				if err := consumeDocument(e.store, plan, id, c); err != nil {
					return receipt, err
				}
				ids = append(ids, id)
			}
			refs.PurchaseDocumentID, refs.HKSDocumentID, refs.TransportDocumentID = proof.Purchase, proof.HKS, proof.Manifest
			eventName = "PickupOffered"
		} else {
			if refs.TransportDocumentID == "" {
				return receipt, fail("MISSING_EVIDENCE")
			}
			eventName = "DeliveryOffered"
		}
		t := Transfer{SchemaVersion: "agrochain.custody-transfer.v1", TransferID: p.TransferID, BatchID: b.BatchID, Kind: kind, Status: "OFFERED", FromMSP: x.Actor, ToMSP: to, QuantityGrams: b.QuantityGrams, OfferOperationID: c.OperationID, OfferedTxID: x.TxID}
		if err := plan.put(key("transfer", t.TransferID), t); err != nil {
			return receipt, err
		}
		if err := plan.put(key("pendingTransfer", b.BatchID), p.TransferID); err != nil {
			return receipt, err
		}
		b.State = next
		ids = append(ids, p.TransferID)
	case "AcceptPickup", "AcceptDelivery":
		var p handoffInput
		_ = json.Unmarshal(c.Payload, &p)
		var t Transfer
		found, err := load(e.store, key("transfer", p.TransferID), &t)
		if err != nil {
			return receipt, err
		}
		if !found || t.BatchID != b.BatchID {
			return receipt, fail("TRANSFER_NOT_FOUND")
		}
		if t.ToMSP != x.Actor {
			return receipt, fail("WRONG_TRANSFER_RECIPIENT")
		}
		state, next, kind := PickupPending, InTransport, "PICKUP"
		if c.Command == "AcceptDelivery" {
			state, next, kind = DeliveryPending, Received, "DELIVERY"
		}
		var pending string
		if _, err := load(e.store, key("pendingTransfer", b.BatchID), &pending); err != nil {
			return receipt, err
		}
		if b.State != state || t.Status != "OFFERED" || t.Kind != kind || pending != t.TransferID {
			return receipt, fail("INVALID_STATE_TRANSITION")
		}
		if p.QuantityGrams != b.QuantityGrams || t.QuantityGrams != b.QuantityGrams {
			return receipt, fail("INVALID_QUANTITY")
		}
		if t.FromMSP != b.CustodianMSP {
			return receipt, fail("UNAUTHORIZED_ORGANIZATION")
		}
		if kind == "DELIVERY" && refs.FreightCostID == "" {
			return receipt, fail("MISSING_EVIDENCE")
		}
		t.Status, t.AcceptOperationID, t.AcceptedTxID = "ACCEPTED", c.OperationID, x.TxID
		if err := plan.put(key("transfer", t.TransferID), t); err != nil {
			return receipt, err
		}
		plan[key("pendingTransfer", b.BatchID)] = nil
		b.State, b.CustodianMSP = next, x.Actor
		if kind == "DELIVERY" {
			b.OwnerMSP = x.Actor
			eventName = "DeliveryAccepted"
		} else {
			eventName = "PickupAccepted"
		}
		ids = append(ids, p.TransferID)
	case "RecordFreightCost":
		var p freightInput
		_ = json.Unmarshal(c.Payload, &p)
		if b.State != InTransport && b.State != DeliveryPending {
			return receipt, fail("INVALID_STATE_TRANSITION")
		}
		if b.CustodianMSP != x.Actor {
			return receipt, fail("UNAUTHORIZED_ORGANIZATION")
		}
		if refs.FreightCostID != "" {
			return receipt, fail("COST_ALREADY_EXISTS")
		}
		if err := absent(e.store, key("costReference", p.CostID), "COST_ALREADY_EXISTS"); err != nil {
			return receipt, err
		}
		if proof.Quantity != b.QuantityGrams {
			return receipt, fail("SOURCE_BINDING_MISMATCH")
		}
		if proof.FreightTotal < 0 || proof.FreightTotal > 1000000000000 || proof.Currency != "TRY" || proof.TaxBasis != "EXCLUDING_TAX" {
			return receipt, fail("INVALID_MONEY")
		}
		refs.FreightCostID = p.CostID
		if err := plan.put(key("costReference", p.CostID), c.BatchID); err != nil {
			return receipt, err
		}
		ids = append(ids, p.CostID)
		eventName = "FreightCostRecorded"
	case "ReportRetailPrice":
		var p reportInput
		_ = json.Unmarshal(c.Payload, &p)
		if b.State != Received {
			return receipt, fail("INVALID_STATE_TRANSITION")
		}
		if b.OwnerMSP != x.Actor || b.CustodianMSP != x.Actor {
			return receipt, fail("UNAUTHORIZED_ORGANIZATION")
		}
		if refs.PurchaseDocumentID == "" || refs.FreightCostID == "" {
			return receipt, fail("MISSING_EVIDENCE")
		}
		if refs.RetailReportID != "" || b.RetailLotID != "" {
			return receipt, fail("REPORT_ALREADY_EXISTS")
		}
		if err := absent(e.store, key("lot", p.RetailLotID), "LOT_ALREADY_EXISTS"); err != nil {
			return receipt, err
		}
		if err := absent(e.store, key("reportReference", p.ReportID), "REPORT_ALREADY_EXISTS"); err != nil {
			return receipt, err
		}
		if proof.Quantity != b.QuantityGrams {
			return receipt, fail("SOURCE_BINDING_MISMATCH")
		}
		if !validPurchase(proof) || proof.RetailPrice < 0 || proof.RetailPrice > 1000000000 {
			return receipt, fail("INVALID_MONEY")
		}
		if p.PolicyID != proof.PolicyID {
			return receipt, fail("POLICY_MISMATCH")
		}
		refs.RetailReportID, b.RetailLotID, b.State = p.ReportID, p.RetailLotID, RetailReported
		lot := Lot{"agrochain.retail-lot.v1", p.RetailLotID, b.BatchID, b.QuantityGrams, x.Actor, p.ReportID, x.TxID}
		if err := plan.put(key("lot", p.RetailLotID), lot); err != nil {
			return receipt, err
		}
		if err := plan.put(key("reportReference", p.ReportID), c.BatchID); err != nil {
			return receipt, err
		}
		ids = append(ids, p.ReportID, p.RetailLotID, p.PolicyID)
		eventName = "RetailPriceReported"
	default:
		return receipt, fail("UNSUPPORTED_PILOT_OPERATION")
	}
	b.Version = old.Version + 1
	b.UpdatedTxID, b.TxTime = x.TxID, x.Time
	ids = sortedUnique(ids)
	receipt = Receipt{"agrochain.receipt.v1", c.OperationID, c.BatchID, c.Command, x.Actor, x.TxID, x.Time, b.Version, ids}
	for _, item := range []struct {
		k string
		v any
	}{
		{key("batch", c.BatchID), b}, {key("batchEvidence", c.BatchID), refs},
		{key("operation", x.Actor, c.OperationID), receipt}, {key("history", c.BatchID, fmt.Sprintf("%010d", b.Version)), receipt},
		{key("owner", b.OwnerMSP, b.BatchID), b.BatchID}, {key("state", b.State, b.BatchID), b.BatchID},
	} {
		if err := plan.put(item.k, item.v); err != nil {
			return Receipt{}, err
		}
	}
	if exists && old.OwnerMSP != b.OwnerMSP {
		plan[key("owner", old.OwnerMSP, b.BatchID)] = nil
	}
	if exists && old.State != b.State {
		plan[key("state", old.State, b.BatchID)] = nil
	}
	event := Event{"agrochain.event.v1", eventName, c.OperationID, c.BatchID, x.Actor, x.TxID, x.Time, b.Version, ids, old.State, b.State}
	for k, v := range proof.Public {
		plan[k] = v
	}
	if len(proof.Private) > 0 {
		ps, ok := e.store.(privateLedger)
		if !ok {
			return Receipt{}, fail("INTERNAL_ERROR")
		}
		for _, w := range proof.Private {
			if err := ps.writePrivate(w.Collection, w.Key, w.Value); err != nil {
				return Receipt{}, fail("INTERNAL_ERROR")
			}
		}
	}
	if err := plan.flush(e.store, event); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validPurchase(p verified) bool {
	// Bounds make both products fit signed int64 (at most 10^17 / 10^15).
	return p.Quantity > 0 && p.Quantity <= 100000000 && p.PurchasePrice > 0 && p.PurchasePrice <= 1000000000 && p.PurchaseTotal > 0 && p.PurchaseTotal <= 1000000000000 && p.Currency == "TRY" && p.TaxBasis == "EXCLUDING_TAX" && p.PurchasePrice*p.Quantity == p.PurchaseTotal*1000
}
func consumeDocument(s ledger, p writePlan, id string, c Command) error {
	if !validID(id, "DOC") {
		return fail("MISSING_EVIDENCE")
	}
	k := key("documentReference", id)
	if _, ok := p[k]; ok {
		return fail("DOCUMENT_ALREADY_USED")
	}
	if err := absent(s, k, "DOCUMENT_ALREADY_USED"); err != nil {
		return err
	}
	// Only a public reference. No fabricated signature/commitment is stored.
	return p.put(k, struct {
		BatchID     string `json:"batchId"`
		OperationID string `json:"operationId"`
	}{c.BatchID, c.OperationID})
}
