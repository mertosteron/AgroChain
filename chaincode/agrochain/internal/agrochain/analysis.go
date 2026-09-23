package agrochain

import (
	"encoding/json"
)

type analysisInput struct {
	AnomalyID string `json:"anomalyId"`
	ReportID  string `json:"reportId"`
}
type score struct {
	Increase       int64  `json:"increaseBps"`
	Classification string `json:"classification"`
	Reason         string `json:"reasonCode"`
}
type anomalyRecord struct {
	SchemaVersion string `json:"schemaVersion"`
	AnomalyID     string `json:"anomalyId"`
	BatchID       string `json:"batchId"`
	ReportID      string `json:"reportId"`
	PolicyID      string `json:"policyId"`
	Rule          string `json:"ruleVersion"`
	Purchase      int64  `json:"purchasePriceKurusPerKg"`
	Retail        int64  `json:"retailPriceKurusPerKg"`
	Freight       int64  `json:"freightTotalKurus"`
	Quantity      int64  `json:"quantityGrams"`
	score
	Threshold   int64  `json:"thresholdBps"`
	ReviewState string `json:"reviewState"`
	Salt        string `json:"recordSaltHex"`
	OperationID string `json:"operationId"`
	TxID        string `json:"txId"`
}
type evaluationReference struct {
	SchemaVersion string `json:"schemaVersion"`
	BatchID       string `json:"batchId"`
	ReportID      string `json:"reportId"`
	AnomalyID     string `json:"anomalyId"`
	TxID          string `json:"txId"`
}
type reviewInput struct {
	Reviewer    string `json:"reviewerRef"`
	Outcome     string `json:"outcome,omitempty"`
	Explanation string `json:"explanation,omitempty"`
	Salt        string `json:"recordSaltHex"`
}
type reviewAction struct {
	SchemaVersion string `json:"schemaVersion"`
	OperationID   string `json:"operationId"`
	AnomalyID     string `json:"anomalyId"`
	Action        string `json:"action"`
	reviewInput
	TxID string `json:"txId"`
}

// Bounded integer products fit int64. Floor is important for negative differences;
// classification uses the exact ratio, never the rounded display basis points.
func priceScore(purchase, retail, threshold int64) (score, error) {
	if purchase <= 0 || purchase > 1000000000 || retail < 0 || retail > 1000000000 {
		return score{}, fail("INVALID_MONEY")
	}
	if threshold < 0 || threshold > 100000 {
		return score{}, fail("POLICY_MISMATCH")
	}
	numerator := (retail - purchase) * 10000
	increase := numerator / purchase
	if numerator < 0 && numerator%purchase != 0 {
		increase--
	}
	s := score{increase, "NO_SIGNAL", "WITHIN_PILOT_THRESHOLD"}
	if numerator > purchase*threshold {
		s.Classification = "REVIEW_REQUIRED"
		s.Reason = "ABOVE_PILOT_THRESHOLD"
	}
	return s, nil
}
func readSalt(t map[string][]byte, name string) (string, error) {
	raw, err := strictJSON(t[name])
	if err != nil {
		return "", err
	}
	s, ok := raw.(string)
	if !ok || !hexID.MatchString(s) {
		return "", fail("INVALID_SCHEMA")
	}
	return s, nil
}
func (e realEvidence) analysis(c Command, b Batch, refs EvidenceRefs) (verified, error) {
	v := verified{Public: writePlan{}}
	if b.State != RetailReported {
		return v, fail("INVALID_STATE_TRANSITION")
	}
	var p analysisInput
	_ = json.Unmarshal(c.Payload, &p)
	var a anomalyRecord
	if c.Command == "EvaluatePrice" {
		if len(e.transient) != 2 {
			return v, fail("INVALID_SCHEMA")
		}
		if refs.RetailReportID == "" || refs.PurchaseDocumentID == "" || refs.FreightCostID == "" {
			return v, fail("MISSING_EVIDENCE")
		}
		if p.ReportID != refs.RetailReportID {
			return v, fail("SOURCE_BINDING_MISMATCH")
		}
		if err := absent(e.s, key("evaluation", b.BatchID), "ANOMALY_ALREADY_EXISTS"); err != nil {
			return v, err
		}
		if err := absent(e.s, key("anomalyReference", p.AnomalyID), "ANOMALY_ALREADY_EXISTS"); err != nil {
			return v, err
		}
		cfg, err := getConfig(e.s)
		if err != nil {
			return v, err
		}
		var buy opening
		var purchase purchaseBody
		var report retailRecord
		var cost costRecord
		if err := privateLoad(e.s, trade, key("documentOpening", refs.PurchaseDocumentID), &buy); err != nil {
			return v, err
		}
		if err := exact(buy.Body, &purchase); err != nil {
			return v, err
		}
		if err := privateLoad(e.s, audit, key("report", refs.RetailReportID), &report); err != nil {
			return v, err
		}
		if err := privateLoad(e.s, freight, key("cost", refs.FreightCostID), &cost); err != nil {
			return v, err
		}
		if report.BatchID != b.BatchID || cost.BatchID != b.BatchID || purchase.Quantity != b.QuantityGrams || report.Quantity != b.QuantityGrams || cost.Quantity != b.QuantityGrams || report.PurchaseID != refs.PurchaseDocumentID || report.CostID != refs.FreightCostID {
			return v, fail("SOURCE_BINDING_MISMATCH")
		}
		if report.PolicyID != cfg.PolicyID {
			return v, fail("POLICY_MISMATCH")
		}
		if purchase.Currency != "TRY" || report.Currency != "TRY" || cost.Currency != "TRY" || purchase.Tax != "EXCLUDING_TAX" || report.Tax != "EXCLUDING_TAX" || cost.Tax != "EXCLUDING_TAX" || cost.Total < 0 || cost.Total > 1000000000000 {
			return v, fail("INVALID_MONEY")
		}
		result, err := priceScore(purchase.Price, report.Price, cfg.ThresholdBps)
		if err != nil {
			return v, err
		}
		var proposed score
		if err := exact(e.transient["proposedResult"], &proposed); err != nil {
			return v, err
		}
		if proposed != result {
			return v, fail("ANOMALY_RESULT_MISMATCH")
		}
		salt, err := readSalt(e.transient, "recordSaltHex")
		if err != nil {
			return v, err
		}
		state := "NOT_REQUIRED"
		if result.Classification == "REVIEW_REQUIRED" {
			state = "OPEN"
		}
		a = anomalyRecord{"agrochain.anomaly-result.v1", p.AnomalyID, b.BatchID, p.ReportID, cfg.PolicyID, "price-increase-v1", purchase.Price, report.Price, cost.Total, b.QuantityGrams, result, cfg.ThresholdBps, state, salt, c.OperationID, e.x.TxID}
		ref := evaluationReference{"agrochain.evaluation-reference.v1", b.BatchID, p.ReportID, p.AnomalyID, e.x.TxID}
		if err := v.Public.put(key("evaluation", b.BatchID), ref); err != nil {
			return v, err
		}
		if err := v.Public.put(key("anomalyReference", p.AnomalyID), ref); err != nil {
			return v, err
		}
	} else {
		if len(e.transient) != 2 {
			return v, fail("INVALID_SCHEMA")
		}
		var ref evaluationReference
		found, err := load(e.s, key("evaluation", b.BatchID), &ref)
		if err != nil {
			return v, err
		}
		if !found || ref.AnomalyID != p.AnomalyID {
			return v, fail("INVALID_REVIEW_TRANSITION")
		}
		if err := privateLoad(e.s, audit, key("anomaly", p.AnomalyID), &a); err != nil {
			return v, err
		}
		if a.Classification != "REVIEW_REQUIRED" {
			return v, fail("INVALID_REVIEW_TRANSITION")
		}
		var input reviewInput
		fields := []string{"reviewerRef", "recordSaltHex"}
		action := "OPEN_REVIEW"
		if c.Command == "ResolveReview" {
			fields = append(fields, "outcome", "explanation")
			action = "RESOLVE_REVIEW"
		}
		if err := decode(e.transient["reviewActionInput"], &input, fields...); err != nil {
			return v, err
		}
		if !validID(input.Reviewer, "REV") || !hexID.MatchString(input.Salt) {
			return v, fail("INVALID_SCHEMA")
		}
		salt, err := readSalt(e.transient, "reviewStateSaltHex")
		if err != nil {
			return v, err
		}
		if salt == input.Salt || salt == a.Salt {
			return v, fail("INVALID_SCHEMA")
		}
		if c.Command == "OpenReview" {
			if a.ReviewState != "OPEN" {
				return v, fail("INVALID_REVIEW_TRANSITION")
			}
			a.ReviewState = "IN_REVIEW"
		} else {
			if a.ReviewState != "IN_REVIEW" {
				return v, fail("INVALID_REVIEW_TRANSITION")
			}
			if (input.Outcome != "EXPLAINED" && input.Outcome != "FOLLOW_UP_RECOMMENDED" && input.Outcome != "INSUFFICIENT_EVIDENCE") || !validExplanation(input.Explanation) {
				return v, fail("INVALID_SCHEMA")
			}
			a.ReviewState = "RESOLVED"
		}
		a.Salt = salt
		if err := v.private(audit, key("reviewAction", p.AnomalyID, c.OperationID), reviewAction{"agrochain.review-action.v1", c.OperationID, p.AnomalyID, action, input, e.x.TxID}); err != nil {
			return v, err
		}
		if err := v.Public.put(key("reviewActionReference", p.AnomalyID, c.OperationID), c.OperationID); err != nil {
			return v, err
		}
	}
	if err := v.private(audit, key("anomaly", p.AnomalyID), a); err != nil {
		return v, err
	}
	return v, nil
}
func analysisQuery(s privateLedger, actor, role, fn, id string) (any, error) {
	if !mayRead(actor, role, audit) {
		return nil, fail("PRIVATE_DATA_ACCESS_DENIED")
	}
	if !validID(id, "BAT") {
		return nil, fail("INVALID_IDENTIFIER")
	}
	var b Batch
	found, err := load(s, key("batch", id), &b)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fail("BATCH_NOT_FOUND")
	}
	var ref evaluationReference
	found, err = load(s, key("evaluation", id), &ref)
	if err != nil {
		return nil, err
	}
	if !found {
		return map[string]string{"status": "EVALUATION_PENDING", "batchId": id}, nil
	}
	if fn == "GetAnomaly" {
		var a anomalyRecord
		err := privateLoad(s, audit, key("anomaly", ref.AnomalyID), &a)
		return a, err
	}
	entries, _, err := s.page("reviewActionReference", []string{ref.AnomalyID}, 10, "")
	if err != nil {
		return nil, err
	}
	actions := []reviewAction{}
	for _, entry := range entries {
		var op string
		if json.Unmarshal(entry.Value, &op) != nil {
			return nil, fail("INTERNAL_ERROR")
		}
		var a reviewAction
		if err := privateLoad(s, audit, key("reviewAction", ref.AnomalyID, op), &a); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}
