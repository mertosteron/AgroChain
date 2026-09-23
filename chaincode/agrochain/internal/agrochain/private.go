package agrochain

import (
	"encoding/json"
)

type purchaseBody struct {
	Seller    string `json:"sellerMsp"`
	Buyer     string `json:"buyerMsp"`
	Quantity  int64  `json:"quantityGrams"`
	Price     int64  `json:"priceKurusPerKg"`
	Total     int64  `json:"totalKurus"`
	Currency  string `json:"currency"`
	Tax       string `json:"taxBasis"`
	Digest    string `json:"attachmentDigest"`
	MediaType string `json:"attachmentMediaType"`
}
type freightBody struct {
	Carrier   string `json:"carrierMsp"`
	Payer     string `json:"payerMsp"`
	Quantity  int64  `json:"quantityGrams"`
	Total     int64  `json:"totalKurus"`
	Currency  string `json:"currency"`
	Tax       string `json:"taxBasis"`
	Digest    string `json:"attachmentDigest"`
	MediaType string `json:"attachmentMediaType"`
}
type costRecord struct {
	SchemaVersion    string `json:"schemaVersion"`
	CostID           string `json:"costId"`
	BatchID          string `json:"batchId"`
	Kind             string `json:"kind"`
	Payer            string `json:"payerMsp"`
	Payee            string `json:"payeeMsp"`
	Quantity         int64  `json:"quantityGrams"`
	Total            int64  `json:"totalKurus"`
	Currency         string `json:"currency"`
	Tax              string `json:"taxBasis"`
	SourceDocumentID string `json:"sourceDocumentId"`
	Salt             string `json:"recordSaltHex"`
	OperationID      string `json:"operationId"`
	TxID             string `json:"txId"`
}
type retailInput struct {
	Price      int64  `json:"offeredPriceKurusPerKg"`
	Currency   string `json:"currency"`
	Tax        string `json:"taxBasis"`
	ReportedAt string `json:"reportedAt"`
	Salt       string `json:"recordSaltHex"`
}
type retailRecord struct {
	SchemaVersion string `json:"schemaVersion"`
	ReportID      string `json:"reportId"`
	BatchID       string `json:"batchId"`
	LotID         string `json:"retailLotId"`
	Quantity      int64  `json:"quantityGrams"`
	Retailer      string `json:"retailerMsp"`
	retailInput
	PurchaseID  string `json:"purchaseDocumentId"`
	CostID      string `json:"freightCostId"`
	PolicyID    string `json:"policyId"`
	OperationID string `json:"operationId"`
	TxID        string `json:"txId"`
}
type realEvidence struct {
	s         privateLedger
	transient map[string][]byte
	x         execution
}

func (v *verified) private(collection, k string, value any) error {
	raw, e := canonical(value)
	if e != nil {
		return e
	}
	v.Private = append(v.Private, privateWrite{collection, k, raw})
	return nil
}
func media(digest, kind string) bool {
	return hexID.MatchString(digest) && (kind == "application/xml" || kind == "application/json" || kind == "application/pdf")
}
func privateLoad(s privateLedger, collection, k string, dst any) error {
	b, e := s.readPrivate(collection, k)
	if e != nil || len(b) == 0 {
		return fail("PRIVATE_DATA_UNAVAILABLE")
	}
	if json.Unmarshal(b, dst) != nil {
		return fail("INTERNAL_ERROR")
	}
	return nil
}
func (e realEvidence) prepare(c Command, b Batch, refs EvidenceRefs) (verified, error) {
	if c.Command == "EvaluatePrice" || c.Command == "OpenReview" || c.Command == "ResolveReview" {
		return e.analysis(c, b, refs)
	}
	v := verified{Public: writePlan{}}
	config, err := getConfig(e.s)
	if err != nil {
		return v, err
	}
	required := map[string]string{}
	extra := ""
	switch c.Command {
	case "CreateBatch":
		required["PRODUCER_ELIGIBILITY"] = "CKS"
	case "OfferPickup":
		required = map[string]string{"PURCHASE_INVOICE": "EFATURA", "TRADE_NOTIFICATION": "HKS", "TRANSPORT_MANIFEST": "UETDS"}
	case "RecordFreightCost":
		required["FREIGHT_INVOICE"] = "EFATURA"
		extra = "recordSaltHex"
	case "ReportRetailPrice":
		extra = "retailReportInput"
	}
	expected := 0
	if len(required) > 0 {
		expected++
	}
	if extra != "" {
		expected++
	}
	if len(e.transient) != expected {
		return v, fail("MISSING_EVIDENCE")
	}
	for k := range e.transient {
		if k != extra && !(k == "evidence" && len(required) > 0) {
			return v, fail("INVALID_SCHEMA")
		}
	}
	var documents []bundle
	if len(required) > 0 {
		documents, err = decodeBundles(e.transient["evidence"])
		if err != nil {
			return v, err
		}
		if len(documents) != len(required) {
			return v, fail("MISSING_EVIDENCE")
		}
	}
	for _, doc := range documents {
		h := doc.Envelope.Header
		if required[h.DocumentType] != h.SourceSystem || required[h.DocumentType] == "" {
			return v, fail("SOURCE_BINDING_MISMATCH")
		}
		delete(required, h.DocumentType)
		if err := verifyBundle(e.s, doc); err != nil {
			return v, err
		}
		if h.BatchID != c.BatchID || h.OperationID != c.OperationID {
			return v, fail("SOURCE_BINDING_MISMATCH")
		}
		for _, check := range []struct{ k, code string }{{key("document", h.DocumentID), "DOCUMENT_ALREADY_USED"}, {key("sourceUse", h.IssuerID, h.SourceDocumentID), "DOCUMENT_ALREADY_USED"}, {key("nonceUse", h.IssuerID, h.Nonce), "SOURCE_NONCE_ALREADY_USED"}} {
			if _, ok := v.Public[check.k]; ok {
				return v, fail(check.code)
			}
			if err := absent(e.s, check.k, check.code); err != nil {
				return v, err
			}
			if err := v.Public.put(check.k, map[string]string{"batchId": c.BatchID, "operationId": c.OperationID, "txId": e.x.TxID}); err != nil {
				return v, err
			}
		}
		if err := v.Public.put(key("document", h.DocumentID), doc.Envelope); err != nil {
			return v, err
		}
		collection := ""
		switch h.DocumentType {
		case "PRODUCER_ELIGIBILITY":
			var body struct {
				Producer string `json:"producerMsp"`
				Product  string `json:"productCode"`
				Grade    string `json:"gradeCode"`
				Quantity int64  `json:"quantityGrams"`
				Region   string `json:"originRegionCode"`
				Harvest  string `json:"harvestDate"`
				Eligible bool   `json:"eligible"`
			}
			if err := exact(doc.Body, &body); err != nil {
				return v, err
			}
			var p createInput
			_ = json.Unmarshal(c.Payload, &p)
			if !body.Eligible {
				return v, fail("SOURCE_CLAIM_REJECTED")
			}
			if body.Producer != e.x.Actor || body.Product != p.ProductCode || body.Grade != p.GradeCode || body.Quantity != p.QuantityGrams || body.Region != p.OriginRegionCode || body.Harvest != p.HarvestDate {
				return v, fail("SOURCE_BINDING_MISMATCH")
			}
			v.CKS = h.DocumentID
			v.Quantity = body.Quantity
		case "PURCHASE_INVOICE":
			var p purchaseBody
			if err := exact(doc.Body, &p); err != nil {
				return v, err
			}
			if p.Seller != b.ProducerMSP || p.Buyer != b.IntendedRetailerMSP || !media(p.Digest, p.MediaType) {
				return v, fail("SOURCE_BINDING_MISMATCH")
			}
			v.Purchase = h.DocumentID
			v.Quantity = p.Quantity
			v.PurchasePrice = p.Price
			v.PurchaseTotal = p.Total
			v.Currency = p.Currency
			v.TaxBasis = p.Tax
			collection = trade
		case "TRADE_NOTIFICATION":
			var p struct {
				Producer string `json:"producerMsp"`
				Retailer string `json:"retailerMsp"`
				Product  string `json:"productCode"`
				Quantity int64  `json:"quantityGrams"`
				Notified bool   `json:"notified"`
			}
			if err := exact(doc.Body, &p); err != nil {
				return v, err
			}
			if !p.Notified {
				return v, fail("SOURCE_CLAIM_REJECTED")
			}
			if p.Producer != b.ProducerMSP || p.Retailer != b.IntendedRetailerMSP || p.Product != b.ProductCode || p.Quantity != b.QuantityGrams {
				return v, fail("SOURCE_BINDING_MISMATCH")
			}
			v.HKS = h.DocumentID
		case "TRANSPORT_MANIFEST":
			var p struct {
				Carrier   string `json:"carrierMsp"`
				Consignor string `json:"consignorMsp"`
				Consignee string `json:"consigneeMsp"`
				Product   string `json:"productCode"`
				Quantity  int64  `json:"quantityGrams"`
				Departure string `json:"declaredDepartureAt"`
			}
			if err := exact(doc.Body, &p); err != nil {
				return v, err
			}
			if p.Carrier != b.LogisticsMSP || p.Consignor != b.ProducerMSP || p.Consignee != b.IntendedRetailerMSP || p.Product != b.ProductCode || p.Quantity != b.QuantityGrams || !validDateTime(p.Departure) {
				return v, fail("SOURCE_BINDING_MISMATCH")
			}
			v.Manifest = h.DocumentID
		case "FREIGHT_INVOICE":
			var p freightBody
			if err := exact(doc.Body, &p); err != nil {
				return v, err
			}
			if p.Carrier != b.LogisticsMSP || p.Payer != b.IntendedRetailerMSP || !media(p.Digest, p.MediaType) {
				return v, fail("SOURCE_BINDING_MISMATCH")
			}
			v.Quantity = p.Quantity
			v.FreightTotal = p.Total
			v.Currency = p.Currency
			v.TaxBasis = p.Tax
			collection = freight
			var salt string
			raw, err := strictJSON(e.transient["recordSaltHex"])
			if err != nil {
				return v, err
			}
			salt, ok := raw.(string)
			if !ok || !hexID.MatchString(salt) {
				return v, fail("INVALID_SCHEMA")
			}
			var input freightInput
			_ = json.Unmarshal(c.Payload, &input)
			if err := v.private(freight, key("cost", input.CostID), costRecord{"agrochain.cost.v1", input.CostID, c.BatchID, "FREIGHT", p.Payer, p.Carrier, p.Quantity, p.Total, p.Currency, p.Tax, h.DocumentID, salt, c.OperationID, e.x.TxID}); err != nil {
				return v, err
			}
		}
		if collection != "" {
			if err := v.private(collection, key("documentOpening", h.DocumentID), opening{doc.Body, doc.Salt}); err != nil {
				return v, err
			}
		}
	}
	if c.Command == "ReportRetailPrice" {
		var input retailInput
		if err := exact(e.transient["retailReportInput"], &input); err != nil {
			return v, err
		}
		if !validDateTime(input.ReportedAt) || !hexID.MatchString(input.Salt) {
			return v, fail("INVALID_SCHEMA")
		}
		var buy opening
		if err := privateLoad(e.s, trade, key("documentOpening", refs.PurchaseDocumentID), &buy); err != nil {
			return v, err
		}
		var p purchaseBody
		if err := exact(buy.Body, &p); err != nil {
			return v, err
		}
		var cost costRecord
		if err := privateLoad(e.s, freight, key("cost", refs.FreightCostID), &cost); err != nil {
			return v, err
		}
		if cost.BatchID != b.BatchID || cost.Quantity != b.QuantityGrams || cost.Currency != input.Currency || cost.Tax != input.Tax || p.Currency != input.Currency || p.Tax != input.Tax {
			return v, fail("SOURCE_BINDING_MISMATCH")
		}
		v.Quantity = p.Quantity
		v.PurchasePrice = p.Price
		v.PurchaseTotal = p.Total
		v.RetailPrice = input.Price
		v.Currency = input.Currency
		v.TaxBasis = input.Tax
		v.PolicyID = config.PolicyID
		var r reportInput
		_ = json.Unmarshal(c.Payload, &r)
		if err := v.private(audit, key("report", r.ReportID), retailRecord{"agrochain.retail-price-report.v1", r.ReportID, c.BatchID, r.RetailLotID, b.QuantityGrams, e.x.Actor, input, refs.PurchaseDocumentID, refs.FreightCostID, r.PolicyID, c.OperationID, e.x.TxID}); err != nil {
			return v, err
		}
	}
	return v, nil
}

func mayRead(actor, role, collection string) bool {
	if actor == Regulator {
		return role == "reviewer" || role == "auditor" || role == "oracle"
	}
	if actor == Retailer {
		return role == "retailer"
	}
	return (collection == trade && actor == Producer && role == "producer") || (collection == freight && actor == Logistics && role == "carrier")
}
func privateQuery(s privateLedger, actor, role, fn, id string, transient map[string][]byte) (any, error) {
	if fn == "GetAnomaly" || fn == "GetReviewHistory" {
		return analysisQuery(s, actor, role, fn, id)
	}
	collection := map[string]string{"GetPurchase": trade, "GetFreightCost": freight, "GetRetailReport": audit}[fn]
	if !mayRead(actor, role, collection) {
		return nil, fail("PRIVATE_DATA_ACCESS_DENIED")
	}
	if !validID(id, "BAT") {
		return nil, fail("INVALID_IDENTIFIER")
	}
	var b Batch
	exists, err := load(s, key("batch", id), &b)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fail("BATCH_NOT_FOUND")
	}
	var refs EvidenceRefs
	if _, err := load(s, key("batchEvidence", id), &refs); err != nil {
		return nil, err
	}
	switch fn {
	case "GetPurchase":
		var o opening
		if err := privateLoad(s, trade, key("documentOpening", refs.PurchaseDocumentID), &o); err != nil {
			return nil, err
		}
		var env envelope
		if _, err := load(s, key("document", refs.PurchaseDocumentID), &env); err != nil {
			return nil, err
		}
		return bundle{env, o.Body, o.Salt}, nil
	case "GetFreightCost":
		var v costRecord
		if err := privateLoad(s, freight, key("cost", refs.FreightCostID), &v); err != nil {
			return nil, err
		}
		return v, nil
	case "GetRetailReport":
		var v retailRecord
		if err := privateLoad(s, audit, key("report", refs.RetailReportID), &v); err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, fail("UNSUPPORTED_PILOT_OPERATION")
}
func verifyOpening(s privateLedger, actor, role, id string, t map[string][]byte) (any, error) {
	if !validID(id, "DOC") {
		return nil, fail("INVALID_IDENTIFIER")
	}
	var env envelope
	found, err := load(s, key("document", id), &env)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fail("DOCUMENT_NOT_FOUND")
	}
	collection := trade
	if env.Header.DocumentType == "FREIGHT_INVOICE" {
		collection = freight
	}
	if !mayRead(actor, role, collection) {
		return nil, fail("PRIVATE_DATA_ACCESS_DENIED")
	}
	if len(t) != 1 {
		return nil, fail("INVALID_SCHEMA")
	}
	var o opening
	if err := exact(t["opening"], &o); err != nil {
		return nil, err
	}
	if err := verifyBundle(s, bundle{env, o.Body, o.Salt}); err != nil {
		return nil, err
	}
	return map[string]any{"schemaVersion": "agrochain.opening-result.v1", "documentId": id, "sourceMode": "SIMULATED", "commitmentMatched": true, "signatureVerified": true}, nil
}
