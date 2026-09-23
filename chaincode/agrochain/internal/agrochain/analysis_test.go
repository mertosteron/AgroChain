package agrochain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPriceScoreBoundaries(t *testing.T) {
	for _, v := range []struct {
		p, r, b int64
		signal  bool
	}{{2000, 2800, 4000, false}, {2000, 3000, 5000, false}, {2000, 3001, 5005, true}, {2000, 3700, 8500, true}, {30000, 45001, 5000, true}, {3000, 2999, -4, false}, {2000, 0, -10000, false}} {
		s, e := priceScore(v.p, v.r, 5000)
		if e != nil || s.Increase != v.b || (s.Classification == "REVIEW_REQUIRED") != v.signal {
			t.Fatalf("%+v: %+v %v", v, s, e)
		}
	}
	for _, p := range []int64{0, -1, 1000000001} {
		_, e := priceScore(p, 2800, 5000)
		code(t, e, "INVALID_MONEY")
	}
}
func analysisFixture(t *testing.T, price int64) *privateMemory {
	m, keys := privacySetup(t)
	for i, step := range steps {
		c := command(i)
		x := context(i, step.actor)
		tr := transientFor(i, c, keys)
		if i == 6 {
			tr["retailReportInput"] = bytesOf(retailInput{price, "TRY", "EXCLUDING_TAX", "2026-09-21T09:30:00.000Z", salt()})
		}
		if _, e := dispatch(m, x.Actor, x.Role, c.Command, []string{string(bytesOf(c))}, proposal{x, tr}); e != nil {
			t.Fatal(e)
		}
	}
	return m
}
func analysisCommand(fn string, version int64) Command {
	c := command(6)
	c.Command = fn
	c.ExpectedVersion = version
	c.OperationID = "OP-" + strings.ToUpper(fn) + "0001"
	p := map[string]string{"anomalyId": "ANM-ANALYSIS1"}
	if fn == "EvaluatePrice" {
		var r reportInput
		_ = json.Unmarshal(c.Payload, &r)
		p["reportId"] = r.ReportID
	}
	c.Payload = bytesOf(p)
	return c
}
func analysisDispatch(m *privateMemory, c Command, role string, tr map[string][]byte) (any, error) {
	x := context(int(c.ExpectedVersion)+30, Regulator)
	x.Role = role
	return dispatch(m, Regulator, role, c.Command, []string{string(bytesOf(c))}, proposal{x, tr})
}
func TestAttestationReviewAndPrivacy(t *testing.T) {
	m := analysisFixture(t, 3700)
	c := analysisCommand("EvaluatePrice", 7)
	s, _ := priceScore(2000, 3700, 5000)
	tr := map[string][]byte{"proposedResult": bytesOf(s), "recordSaltHex": bytesOf(salt())}
	_, err := analysisDispatch(m, c, "auditor", tr)
	code(t, err, "UNAUTHORIZED_ROLE")
	wrong := s
	wrong.Increase = 4000
	tr["proposedResult"] = bytesOf(wrong)
	before := len(m.private)
	_, err = analysisDispatch(m, c, "oracle", tr)
	code(t, err, "ANOMALY_RESULT_MISMATCH")
	if len(m.private) != before {
		t.Fatal("partial private write")
	}
	tr["proposedResult"] = bytesOf(s)
	if _, err = analysisDispatch(m, c, "oracle", tr); err != nil {
		t.Fatal(err)
	}
	_, err = analysisDispatch(m, c, "oracle", tr)
	code(t, err, "DUPLICATE_TRANSACTION")
	for _, actor := range []string{Producer, Logistics} {
		_, err = analysisQuery(m, actor, context(0, actor).Role, "GetAnomaly", c.BatchID)
		code(t, err, "PRIVATE_DATA_ACCESS_DENIED")
	}
	_, err = analysisQuery(m, Regulator, "public-reader", "GetAnomaly", c.BatchID)
	code(t, err, "PRIVATE_DATA_ACCESS_DENIED")
	original, _ := analysisQuery(m, Retailer, "retailer", "GetAnomaly", c.BatchID)
	resolve := analysisCommand("ResolveReview", 8)
	review := map[string][]byte{"reviewActionInput": bytesOf(reviewInput{"REV-REVIEWER1", "EXPLAINED", "Taşıma ve kalite farkı incelendi.", salt()}), "reviewStateSaltHex": bytesOf(salt())}
	_, err = analysisDispatch(m, resolve, "reviewer", review)
	code(t, err, "INVALID_REVIEW_TRANSITION")
	open := analysisCommand("OpenReview", 8)
	opentr := map[string][]byte{"reviewActionInput": bytesOf(reviewInput{Reviewer: "REV-REVIEWER1", Salt: salt()}), "reviewStateSaltHex": bytesOf(salt())}
	if _, err = analysisDispatch(m, open, "reviewer", opentr); err != nil {
		t.Fatal(err)
	}
	resolve.ExpectedVersion = 9
	if _, err = analysisDispatch(m, resolve, "reviewer", review); err != nil {
		t.Fatal(err)
	}
	final, _ := analysisQuery(m, Retailer, "retailer", "GetAnomaly", c.BatchID)
	a := final.(anomalyRecord)
	a.ReviewState = original.(anomalyRecord).ReviewState
	a.Salt = original.(anomalyRecord).Salt
	if !reflect.DeepEqual(a, original) {
		t.Fatal("classification or evidence changed")
	}
	history, err := analysisQuery(m, Regulator, "reviewer", "GetReviewHistory", c.BatchID)
	if err != nil || len(history.([]reviewAction)) != 2 {
		t.Fatalf("history %v %v", history, err)
	}
	for k, v := range m.values {
		_ = k
		for _, secret := range []string{"purchasePriceKurusPerKg", "increaseBps", "REVIEW_REQUIRED", "EXPLAINED", "Taşıma"} {
			if strings.Contains(string(v), secret) {
				t.Fatalf("shared secret %s", secret)
			}
		}
	}
}
func TestMissingEvidenceNeverNormal(t *testing.T) {
	m := analysisFixture(t, 2800)
	c := analysisCommand("EvaluatePrice", 7)
	var refs EvidenceRefs
	_, _ = load(m, key("batchEvidence", c.BatchID), &refs)
	delete(m.private, freight+key("cost", refs.FreightCostID))
	s, _ := priceScore(2000, 2800, 5000)
	_, e := analysisDispatch(m, c, "oracle", map[string][]byte{"proposedResult": bytesOf(s), "recordSaltHex": bytesOf(salt())})
	code(t, e, "PRIVATE_DATA_UNAVAILABLE")
	result, e := analysisQuery(m, Retailer, "retailer", "GetAnomaly", c.BatchID)
	if e != nil || result.(map[string]string)["status"] != "EVALUATION_PENDING" {
		t.Fatal(result, e)
	}
}
