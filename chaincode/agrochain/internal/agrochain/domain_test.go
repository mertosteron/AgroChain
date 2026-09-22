package agrochain

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// This is the only successful evidence implementation. It is excluded from the
// deployment binary by Go's _test.go convention, not a runtime configuration flag.
type fixtureEvidence struct {
	proof verified
	err   error
}

func (f fixtureEvidence) prepare(Command, Batch, EvidenceRefs) (verified, error) {
	return f.proof, f.err
}
func proof() verified {
	return verified{CKS: "DOC-CKS00001", Purchase: "DOC-BUY00001", HKS: "DOC-HKS00001", Manifest: "DOC-MAN00001", Quantity: 100000, PurchasePrice: 2000, PurchaseTotal: 200000, FreightTotal: 20000, RetailPrice: 2800, Currency: "TRY", TaxBasis: "EXCLUDING_TAX", PolicyID: "CFG-PRICE001"}
}

type memory struct {
	values                         map[string][]byte
	events                         []Event
	writes                         int
	failRead, failWrite, failEvent bool
}

func newMemory() *memory { return &memory{values: map[string][]byte{}} }
func (m *memory) read(k string) ([]byte, error) {
	if m.failRead {
		return nil, errors.New("secret storage detail")
	}
	return m.values[k], nil
}
func (m *memory) write(k string, v []byte) error {
	if m.failWrite {
		return errors.New("secret")
	}
	m.values[k] = append([]byte{}, v...)
	m.writes++
	return nil
}
func (m *memory) remove(k string) error {
	if m.failWrite {
		return errors.New("secret")
	}
	delete(m.values, k)
	m.writes++
	return nil
}
func (m *memory) emit(_ string, v []byte) error {
	if m.failEvent {
		return errors.New("secret")
	}
	var e Event
	_ = json.Unmarshal(v, &e)
	m.events = append(m.events, e)
	return nil
}
func (m *memory) page(kind string, attrs []string, size int32, mark string) ([]entry, string, error) {
	if mark != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(mark)
		if err != nil {
			return nil, "", fail("INVALID_PAGE")
		}
		mark = string(decoded)
	}
	keys := []string{}
	prefix := key(kind, attrs...)
	for k := range m.values {
		if strings.HasPrefix(k, prefix) && k > mark {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	next := ""
	if len(keys) > int(size) {
		keys = keys[:size]
		next = base64.RawURLEncoding.EncodeToString([]byte(keys[len(keys)-1]))
	}
	items := []entry{}
	for _, k := range keys {
		items = append(items, entry{k, m.values[k]})
	}
	return items, next, nil
}
func (m *memory) clone() *memory {
	n := newMemory()
	for k, v := range m.values {
		n.values[k] = append([]byte{}, v...)
	}
	n.events = append(n.events, m.events...)
	return n
}

var steps = []struct {
	fn, actor string
	payload   any
}{
	{"CreateBatch", Producer, createInput{"TOMATO", "STANDARD", 100000, "07", "2026-09-15", Logistics, Retailer}},
	{"OfferPickup", Producer, handoffInput{"TRF-PICKUP01", 100000}},
	{"AcceptPickup", Logistics, handoffInput{"TRF-PICKUP01", 100000}},
	{"RecordFreightCost", Logistics, freightInput{"CST-FREIGHT1"}},
	{"OfferDelivery", Logistics, handoffInput{"TRF-DELIVER1", 100000}},
	{"AcceptDelivery", Retailer, handoffInput{"TRF-DELIVER1", 100000}},
	{"ReportRetailPrice", Retailer, reportInput{"RPT-RETAIL01", "LOT-RETAIL01", "CFG-PRICE001"}},
}

func command(i int) Command {
	b, _ := json.Marshal(steps[i].payload)
	return Command{"agrochain.command.v1", fmt.Sprintf("OP-STEP%04d", i), "BAT-NORMAL01", int64(i), steps[i].fn, b}
}
func context(i int, actor string) execution {
	return execution{Actor: actor, Role: map[string]string{Producer: "producer", Logistics: "carrier", Retailer: "retailer", Regulator: "auditor"}[actor], TxID: fmt.Sprintf("%064x", i+1), Time: TxTime{Seconds: "9223372036854775807", Nanos: 123456789}}
}
func run(e engine, c Command, x execution) (Receipt, error) {
	b, _ := json.Marshal(c)
	parsed, err := parseCommand(c.Command, string(b))
	if err != nil {
		return Receipt{}, err
	}
	return e.apply(parsed, x)
}
func setup(t *testing.T, n int) (*memory, engine) {
	t.Helper()
	m := newMemory()
	e := engine{m, fixtureEvidence{proof(), nil}}
	for i := 0; i < n; i++ {
		if _, err := run(e, command(i), context(i, steps[i].actor)); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	return m, e
}
func code(t *testing.T, err error, want string) {
	t.Helper()
	var d *DomainError
	if !errors.As(err, &d) || d.Code != want {
		t.Fatalf("want %s, got %v", want, err)
	}
	var fields map[string]any
	if json.Unmarshal([]byte(err.Error()), &fields) != nil || len(fields) != 4 {
		t.Fatal("unsafe error envelope", err)
	}
}
func reject(t *testing.T, m *memory, e engine, c Command, x execution, want string) {
	t.Helper()
	before := m.clone()
	writes := m.writes
	_, err := run(e, c, x)
	code(t, err, want)
	if !reflect.DeepEqual(before.values, m.values) || !reflect.DeepEqual(before.events, m.events) || m.writes != writes {
		t.Fatal("failed validation produced writes/events")
	}
}
func batch(t *testing.T, m *memory) Batch {
	t.Helper()
	var b Batch
	ok, err := load(m, key("batch", "BAT-NORMAL01"), &b)
	if !ok || err != nil {
		t.Fatal(err)
	}
	return b
}

func TestLifecycle(t *testing.T) {
	m := newMemory()
	e := engine{m, fixtureEvidence{proof(), nil}}
	states := []string{Created, PickupPending, InTransport, InTransport, DeliveryPending, Received, RetailReported}
	for i, s := range states {
		r, err := run(e, command(i), context(i, steps[i].actor))
		if err != nil {
			t.Fatal(err)
		}
		b := batch(t, m)
		if b.State != s || b.Version != int64(i+1) || b.QuantityGrams != 100000 || r.TxTime != context(i, steps[i].actor).Time {
			t.Fatalf("bad step %d: %+v", i, b)
		}
		if i < 5 && b.OwnerMSP != Producer {
			t.Fatal("carrier became owner")
		}
		if i >= 2 && i <= 4 && b.CustodianMSP != Logistics {
			t.Fatal("wrong custody")
		}
		if i >= 5 && (b.OwnerMSP != Retailer || b.CustodianMSP != Retailer) {
			t.Fatal("retailer handoff not atomic")
		}
		if b.CreatedTxID != context(0, Producer).TxID || b.UpdatedTxID != r.TxID {
			t.Fatal("wrong Fabric tx references")
		}
		ev := m.events[i]
		if ev.TxID != r.TxID || ev.ResultVersion != r.ResultVersion || ev.NewState != s || ev.ActorMSP != steps[i].actor || !sort.StringsAreSorted(ev.ObjectIDs) {
			t.Fatal("event mismatch")
		}
	}
	if len(m.events) != 7 {
		t.Fatal("one event per successful command")
	}
	for _, id := range []string{"TRF-PICKUP01", "TRF-DELIVER1"} {
		var tr Transfer
		_, _ = load(m, key("transfer", id), &tr)
		if tr.Status != "ACCEPTED" || tr.OfferedTxID == "" || tr.AcceptedTxID == "" {
			t.Fatal("transfer references lost")
		}
	}
	if len(m.values[key("owner", Producer, "BAT-NORMAL01")]) != 0 || len(m.values[key("state", Created, "BAT-NORMAL01")]) != 0 {
		t.Fatal("stale indexes")
	}
}
func TestReplays(t *testing.T) {
	m, e := setup(t, 7)
	for i := range steps {
		c := command(i)
		reject(t, m, e, c, context(10, steps[i].actor), "DUPLICATE_TRANSACTION")
		c.BatchID = "BAT-CHANGED1"
		reject(t, m, e, c, context(11, steps[i].actor), "DUPLICATE_TRANSACTION")
	}
	// Same ID in a different organization is a different namespace.
	if err := absent(m, key("operation", Regulator, command(0).OperationID), "DUPLICATE_TRANSACTION"); err != nil {
		t.Fatal(err)
	}
	_, err := query(m, Regulator, "GetOperation", []string{command(0).OperationID})
	code(t, err, "OPERATION_NOT_FOUND")
}
func TestRejectedTransitionsAndAuthorization(t *testing.T) {
	for n := 1; n <= 7; n++ {
		t.Run(fmt.Sprintf("state%d", n), func(t *testing.T) {
			m, e := setup(t, n)
			c := command(1)
			c.OperationID = "OP-ILLEGAL1"
			c.ExpectedVersion = int64(n)
			if n == 1 {
				c = command(4)
				c.OperationID = "OP-ILLEGAL1"
				c.ExpectedVersion = int64(n)
			}
			actor := Producer
			if n == 1 {
				actor = Logistics
			}
			reject(t, m, e, c, context(20, actor), "INVALID_STATE_TRANSITION")
		})
	}
	for i := range steps {
		for _, actor := range []string{Producer, Logistics, Retailer, Regulator, "UnknownMSP"} {
			if actor == steps[i].actor {
				continue
			}
			t.Run(fmt.Sprintf("%s/%s", steps[i].fn, actor), func(t *testing.T) {
				m, e := setup(t, i)
				want := "UNAUTHORIZED_ORGANIZATION"
				if member(actor) && (i == 2 || i == 5) {
					want = "WRONG_TRANSFER_RECIPIENT"
				}
				reject(t, m, e, command(i), context(i, actor), want)
			})
		}
	}
}
func TestDuplicatesVersionsAndQuantities(t *testing.T) {
	m, e := setup(t, 1)
	c := command(0)
	c.OperationID = "OP-DUPBATCH"
	reject(t, m, e, c, context(20, Producer), "BATCH_ALREADY_EXISTS")
	c = command(1)
	c.ExpectedVersion = 0
	reject(t, m, e, c, context(20, Producer), "VERSION_CONFLICT")
	for _, n := range []int{1, 2, 4, 5} {
		m, e := setup(t, n)
		c := command(n)
		var p handoffInput
		_ = json.Unmarshal(c.Payload, &p)
		p.QuantityGrams--
		c.Payload, _ = json.Marshal(p)
		reject(t, m, e, c, context(20, steps[n].actor), "INVALID_QUANTITY")
	}
	m, e = setup(t, 3)
	c = command(2)
	c.OperationID = "OP-REPEAT01"
	c.ExpectedVersion = 3
	reject(t, m, e, c, context(20, Logistics), "INVALID_STATE_TRANSITION")
	m, e = setup(t, 4)
	c = command(3)
	c.OperationID = "OP-REPEAT01"
	c.ExpectedVersion = 4
	reject(t, m, e, c, context(20, Logistics), "COST_ALREADY_EXISTS")
	m, e = setup(t, 2)
	c = command(1)
	c.OperationID = "OP-PENDING1"
	c.ExpectedVersion = 2
	reject(t, m, e, c, context(20, Producer), "INVALID_STATE_TRANSITION")
	m, e = setup(t, 2)
	c = command(2)
	c.Payload = []byte(`{"transferId":"TRF-MISSING1","quantityGrams":100000}`)
	reject(t, m, e, c, context(20, Logistics), "TRANSFER_NOT_FOUND")
}
func TestEvidenceAndMoney(t *testing.T) {
	for _, tc := range []struct {
		name   string
		n      int
		change func(*verified)
		want   string
	}{
		{"missing CKS", 0, func(p *verified) { p.CKS = "" }, "MISSING_EVIDENCE"},
		{"quantity binding", 0, func(p *verified) { p.Quantity-- }, "SOURCE_BINDING_MISMATCH"},
		{"duplicate documents", 1, func(p *verified) { p.Purchase = p.CKS }, "DOCUMENT_ALREADY_USED"},
		{"same documents in offer", 1, func(p *verified) { p.Manifest = p.HKS }, "DOCUMENT_ALREADY_USED"},
		{"negative purchase", 1, func(p *verified) { p.PurchasePrice = -1 }, "INVALID_MONEY"},
		{"zero purchase", 1, func(p *verified) { p.PurchasePrice = 0 }, "INVALID_MONEY"},
		{"residual purchase", 1, func(p *verified) { p.PurchaseTotal++ }, "INVALID_MONEY"},
		{"freight negative", 3, func(p *verified) { p.FreightTotal = -1 }, "INVALID_MONEY"},
		{"freight overflow", 3, func(p *verified) { p.FreightTotal = 1000000000001 }, "INVALID_MONEY"},
		{"retail negative", 6, func(p *verified) { p.RetailPrice = -1 }, "INVALID_MONEY"},
		{"retail overflow", 6, func(p *verified) { p.RetailPrice = 1000000001 }, "INVALID_MONEY"},
		{"currency", 6, func(p *verified) { p.Currency = "USD" }, "INVALID_MONEY"},
		{"tax", 6, func(p *verified) { p.TaxBasis = "INCLUDING_TAX" }, "INVALID_MONEY"},
		{"policy", 6, func(p *verified) { p.PolicyID = "CFG-OTHER001" }, "POLICY_MISMATCH"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, e := setup(t, tc.n)
			p := proof()
			tc.change(&p)
			e.evidence = fixtureEvidence{p, nil}
			reject(t, m, e, command(tc.n), context(20, steps[tc.n].actor), tc.want)
		})
	}
	for _, n := range []int{3, 6} {
		m, e := setup(t, n)
		p := proof()
		p.FreightTotal = 0
		p.RetailPrice = 0
		e.evidence = fixtureEvidence{p, nil}
		if _, err := run(e, command(n), context(20, steps[n].actor)); err != nil {
			t.Fatal(err)
		}
		_ = m
	}
	m, e := setup(t, 3)
	c := command(4)
	c.ExpectedVersion = 3
	if _, err := run(e, c, context(20, Logistics)); err != nil {
		t.Fatal(err)
	}
	c = command(5)
	c.ExpectedVersion = 4
	reject(t, m, e, c, context(21, Retailer), "MISSING_EVIDENCE")
	m, e = setup(t, 0)
	e.evidence = closedEvidence{}
	reject(t, m, e, command(0), context(0, Producer), "EVIDENCE_VERIFICATION_UNAVAILABLE")
}
func TestPublicStateAndEventsContainNoPrivateData(t *testing.T) {
	m, _ := setup(t, 7)
	ev, _ := json.Marshal(m.events)
	all := string(ev)
	for _, v := range m.values {
		all += string(v)
	}
	for _, secret := range []string{"priceKurus", "offeredPrice", "totalKurus", "recordSalt", "saltHex", "signature", "attachment", "200000", "2800", "20000"} {
		if strings.Contains(all, secret) {
			t.Fatalf("private data leak %q", secret)
		}
	}
}
func TestFailClosedDispatcher(t *testing.T) {
	for i := range steps {
		m := newMemory()
		c := command(i)
		raw, _ := json.Marshal(c)
		_, err := dispatch(m, steps[i].actor, context(i, steps[i].actor).Role, steps[i].fn, []string{string(raw)})
		code(t, err, "CONFIGURATION_REQUIRED")
		if m.writes != 0 || len(m.events) != 0 {
			t.Fatal("deployed write path enabled")
		}
	}
	for _, fn := range []string{"CancelCustodyTransfer", "SplitBatch", "MergeBatch", "EvaluatePrice", "SetVerifier", "InitLedger"} {
		_, err := dispatch(newMemory(), Regulator, "auditor", fn, []string{"BAT-NORMAL01"})
		code(t, err, "UNSUPPORTED_PILOT_OPERATION")
	}
}
func TestConcurrentAcceptanceMVCCModel(t *testing.T) {
	m, _ := setup(t, 2)
	left, right := m.clone(), m.clone()
	a, b := command(2), command(2)
	b.OperationID = "OP-RACE0001"
	for i, p := range []*memory{left, right} {
		c := a
		if i == 1 {
			c = b
		}
		if _, err := run(engine{p, fixtureEvidence{proof(), nil}}, c, context(i+10, Logistics)); err != nil {
			t.Fatal(err)
		}
	}
	// Both simulations can succeed. A validator comparing the read batch version
	// can commit only one. This models MVCC; it is not a live Fabric race claim.
	observed := batch(t, m).Version
	m = left
	if observed == batch(t, m).Version {
		t.Fatal("read version must conflict after first commit")
	}
	if len(m.events) != 3 {
		t.Fatal("duplicate acceptance committed")
	}
}
