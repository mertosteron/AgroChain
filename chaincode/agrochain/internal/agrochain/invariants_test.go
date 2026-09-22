package agrochain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestUniqueReferenceAndMissingEvidenceGuards(t *testing.T) {
	for _, tc := range []struct {
		name           string
		n              int
		kind, id, want string
	}{
		{"transfer", 1, "transfer", "TRF-PICKUP01", "TRANSFER_ALREADY_EXISTS"},
		{"pending", 1, "pendingTransfer", "BAT-NORMAL01", "INVALID_STATE_TRANSITION"},
		{"cost", 3, "costReference", "CST-FREIGHT1", "COST_ALREADY_EXISTS"},
		{"lot", 6, "lot", "LOT-RETAIL01", "LOT_ALREADY_EXISTS"},
		{"report", 6, "reportReference", "RPT-RETAIL01", "REPORT_ALREADY_EXISTS"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, e := setup(t, tc.n)
			m.values[key(tc.kind, tc.id)] = []byte(`"reserved"`)
			reject(t, m, e, command(tc.n), context(20, steps[tc.n].actor), tc.want)
		})
	}
	for _, n := range []int{1, 4, 6} {
		m, e := setup(t, n)
		refs := EvidenceRefs{SchemaVersion: "agrochain.batch-evidence.v1", BatchID: "BAT-NORMAL01"}
		m.values[key("batchEvidence", refs.BatchID)], _ = canonical(refs)
		reject(t, m, e, command(n), context(20, steps[n].actor), "MISSING_EVIDENCE")
	}
}
func TestTamperedOwnerAndTransferBindings(t *testing.T) {
	for _, n := range []int{1, 3, 4, 6} {
		m, e := setup(t, n)
		b := batch(t, m)
		b.CustodianMSP = Regulator
		m.values[key("batch", b.BatchID)], _ = canonical(b)
		reject(t, m, e, command(n), context(20, steps[n].actor), "UNAUTHORIZED_ORGANIZATION")
	}
	for _, tc := range []struct {
		name, want string
		change     func(*Transfer)
	}{
		{"recipient", "WRONG_TRANSFER_RECIPIENT", func(tr *Transfer) { tr.ToMSP = Retailer }},
		{"batch", "TRANSFER_NOT_FOUND", func(tr *Transfer) { tr.BatchID = "BAT-OTHER001" }},
		{"quantity", "INVALID_QUANTITY", func(tr *Transfer) { tr.QuantityGrams-- }},
		{"sender", "UNAUTHORIZED_ORGANIZATION", func(tr *Transfer) { tr.FromMSP = Regulator }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, e := setup(t, 2)
			var tr Transfer
			_, _ = load(m, key("transfer", "TRF-PICKUP01"), &tr)
			tc.change(&tr)
			m.values[key("transfer", tr.TransferID)], _ = canonical(tr)
			reject(t, m, e, command(2), context(20, Logistics), tc.want)
		})
	}
}
func TestVersionOverflowAndContextValidation(t *testing.T) {
	m, e := setup(t, 1)
	b := batch(t, m)
	b.Version = 2147483647
	m.values[key("batch", b.BatchID)], _ = canonical(b)
	c := command(1)
	c.ExpectedVersion = b.Version
	reject(t, m, e, c, context(20, Producer), "VERSION_CONFLICT")
	for _, x := range []execution{{Actor: Producer, TxID: "client-tx"}, {Actor: Producer, TxID: context(0, Producer).TxID, Time: TxTime{"01", 0}}, {Actor: Producer, TxID: context(0, Producer).TxID, Time: TxTime{"1", 1000000000}}} {
		m, e := setup(t, 0)
		x.Role = "producer"
		reject(t, m, e, command(0), x, "INTERNAL_ERROR")
	}
}
func TestQueryOrderingAndBounds(t *testing.T) {
	m, _ := setup(t, 1)
	original := batch(t, m)
	for _, suffix := range []string{"00000004", "00000002", "00000003", "00000001"} {
		b := original
		b.BatchID = "BAT-" + suffix
		m.values[key("batch", b.BatchID)], _ = canonical(b)
		m.values[key("owner", Producer, b.BatchID)], _ = canonical(b.BatchID)
	}
	ids := []string{}
	bookmark := ""
	for {
		p, _ := json.Marshal(pageInput{"agrochain.page-request.v1", Producer, 2, bookmark})
		result, err := query(m, Producer, "QueryBatchesByOwner", []string{string(p)})
		if err != nil {
			t.Fatal(err)
		}
		page := result.(Page)
		for _, raw := range page.Records {
			var b Batch
			_ = json.Unmarshal(raw, &b)
			ids = append(ids, b.BatchID)
		}
		bookmark = page.Bookmark
		if bookmark == "" {
			break
		}
	}
	want := []string{"BAT-00000001", "BAT-00000002", "BAT-00000003", "BAT-00000004", "BAT-NORMAL01"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatal(ids)
	}
	for _, tc := range []struct{ fn, filter, want string }{{"QueryBatchesByState", "SOLD", "INVALID_STATE_TRANSITION"}, {"QueryBatchesByOwner", "UnknownMSP", "UNAUTHORIZED_ORGANIZATION"}, {"GetBatchHistory", "invalid", "INVALID_IDENTIFIER"}} {
		p, _ := json.Marshal(pageInput{"agrochain.page-request.v1", tc.filter, 2, ""})
		_, err := query(m, Producer, tc.fn, []string{string(p)})
		code(t, err, tc.want)
	}
}
func TestSimulationFailuresDiscardAllStagedWrites(t *testing.T) {
	committed, _ := setup(t, 1)
	before := committed.clone()
	simulation := committed.clone()
	simulation.failEvent = true
	_, err := run(engine{simulation, fixtureEvidence{proof(), nil}}, command(1), context(20, Producer))
	code(t, err, "INTERNAL_ERROR")
	// Model Fabric's transaction boundary: a failed simulation is never committed,
	// even if PutState had already staged writes before SetEvent failed.
	if !reflect.DeepEqual(committed.values, before.values) || !reflect.DeepEqual(committed.events, before.events) {
		t.Fatal("partial transaction committed")
	}
}
func TestPayloadIdentifiersAndBounds(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4, 5, 6} {
		c := command(n)
		var p map[string]any
		_ = json.Unmarshal(c.Payload, &p)
		for k, v := range p {
			if _, ok := v.(string); ok {
				p[k] = "INVALID"
				break
			}
		}
		c.Payload, _ = json.Marshal(p)
		b, _ := json.Marshal(c)
		_, err := parseCommand(c.Command, string(b))
		code(t, err, "INVALID_IDENTIFIER")
	}
	for _, n := range []int{1, 2, 4, 5} {
		c := command(n)
		c.Payload = []byte(fmt.Sprintf(`{"transferId":"TRF-PICKUP01","quantityGrams":%d}`, 0))
		b, _ := json.Marshal(c)
		_, err := parseCommand(c.Command, string(b))
		code(t, err, "INVALID_QUANTITY")
	}
}
