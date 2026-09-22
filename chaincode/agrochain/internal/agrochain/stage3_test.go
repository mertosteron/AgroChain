package agrochain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// Exercise every command at every checkpoint, including transport before/after
// freight. Evidence is test-only; these are domain, not live Fabric, assertions.
func TestEveryCommandAtEveryLifecycleCheckpoint(t *testing.T) {
	want := [8][7]string{
		{"", "BATCH_NOT_FOUND", "BATCH_NOT_FOUND", "BATCH_NOT_FOUND", "BATCH_NOT_FOUND", "BATCH_NOT_FOUND", "BATCH_NOT_FOUND"},
		{"BATCH_ALREADY_EXISTS", "", "TRANSFER_NOT_FOUND", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "TRANSFER_NOT_FOUND", "INVALID_STATE_TRANSITION"},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "TRANSFER_NOT_FOUND", "INVALID_STATE_TRANSITION"},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "", "", "TRANSFER_NOT_FOUND", "INVALID_STATE_TRANSITION"},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "COST_ALREADY_EXISTS", "", "TRANSFER_NOT_FOUND", "INVALID_STATE_TRANSITION"},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "COST_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "", "INVALID_STATE_TRANSITION"},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", ""},
		{"BATCH_ALREADY_EXISTS", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION", "INVALID_STATE_TRANSITION"},
	}
	for n, outcomes := range want {
		for i, expected := range outcomes {
			t.Run(fmt.Sprintf("checkpoint%d/%s", n, steps[i].fn), func(t *testing.T) {
				m, e := setup(t, n)
				c := command(i)
				c.OperationID = fmt.Sprintf("OP-MATRIX%02d%02d", n, i)
				c.ExpectedVersion = int64(n)
				x := context(50+i, steps[i].actor)
				if expected != "" {
					reject(t, m, e, c, x, expected)
					return
				}
				r, err := run(e, c, x)
				if err != nil {
					t.Fatal(err)
				}
				if r.ResultVersion != int64(n+1) || len(m.events) != n+1 {
					t.Fatal("successful command must produce one version and one event")
				}
				result, err := query(m, steps[i].actor, "GetOperation", []string{c.OperationID})
				if err != nil || !reflect.DeepEqual(result, r) {
					t.Fatal("committed domain receipt cannot be recovered", result, err)
				}
			})
		}
	}
}

func TestLateFreightRecoveryAndPopulatedQueries(t *testing.T) {
	m, e := setup(t, 3)
	apply := func(index int, version int64) Receipt {
		t.Helper()
		c := command(index)
		c.ExpectedVersion = version
		r, err := run(e, c, context(60+index, steps[index].actor))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	apply(4, 3) // Delivery can be offered before freight is recorded.
	accept := command(5)
	accept.ExpectedVersion = 4
	reject(t, m, e, accept, context(70, Retailer), "MISSING_EVIDENCE")
	if _, err := query(m, Retailer, "GetOperation", []string{accept.OperationID}); err == nil {
		t.Fatal("rejected acceptance consumed its operation ID")
	}
	apply(3, 4)
	apply(5, 5) // Same operation ID succeeds after the missing evidence is supplied.
	apply(6, 6)
	before := m.clone()
	for _, actor := range []string{Producer, Logistics, Retailer, Regulator} {
		got, err := query(m, actor, "GetBatch", []string{"BAT-NORMAL01"})
		b, ok := got.(Batch)
		if err != nil || !ok || b.State != RetailReported || b.OwnerMSP != Retailer || b.CustodianMSP != Retailer || b.Version != 7 {
			t.Fatal("populated shared batch query", got, err)
		}
		for _, test := range []struct {
			fn, filter string
			count      int
		}{
			{"QueryBatchesByOwner", Producer, 0},
			{"QueryBatchesByOwner", Logistics, 0},
			{"QueryBatchesByOwner", Retailer, 1},
			{"QueryBatchesByState", Created, 0},
			{"QueryBatchesByState", RetailReported, 1},
			{"GetBatchHistory", "BAT-NORMAL01", 7},
		} {
			raw, _ := json.Marshal(pageInput{"agrochain.page-request.v1", test.filter, 100, ""})
			got, err := query(m, actor, test.fn, []string{string(raw)})
			page, ok := got.(Page)
			if err != nil || !ok || len(page.Records) != test.count || page.Bookmark != "" {
				t.Fatal(actor, test.fn, got, err)
			}
			if test.fn == "GetBatchHistory" {
				order := []string{"CreateBatch", "OfferPickup", "AcceptPickup", "OfferDelivery", "RecordFreightCost", "AcceptDelivery", "ReportRetailPrice"}
				for i, raw := range page.Records {
					var r Receipt
					if json.Unmarshal(raw, &r) != nil || r.ResultVersion != int64(i+1) || r.Command != order[i] {
						t.Fatal("history lost committed operation ordering", r)
					}
				}
			}
		}
	}
	if !reflect.DeepEqual(before.values, m.values) || !reflect.DeepEqual(before.events, m.events) {
		t.Fatal("shared queries mutated the ledger")
	}
}
