package agrochain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestStrictJSON(t *testing.T) {
	for _, raw := range []string{`{"a":1,"a":2}`, `{"a":null}`, `{"a":1.0}`, `{"a":1e2}`, `{"a":-0}`, `{"a":01}`, `{"a":9223372036854775808}`, `{"__proto__":{}}`, `{"constructor":{}}`, `{"prototype":{}}`, `{"a":"\ud800"}`, `{"a":"e\u0301"}`, `{"a":"\u0000"}`, `{"a":1} {}`, `{"a":NaN}`, "{\"a\":\"\xff\"}", strings.Repeat(" ", 65537)} {
		t.Run(fmt.Sprintf("%x", []byte(raw)[:min(12, len(raw))]), func(t *testing.T) { _, err := strictJSON([]byte(raw)); code(t, err, "INVALID_SCHEMA") })
	}
	// Stage 4 evidence needs booleans and bounded arrays; commands still reject
	// them where a scalar or object is required.
	for _, raw := range []string{`{"a":true}`, `{"a":[]}`, `[1]`} {
		if _, err := strictJSON([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, replacement := range []string{"true", "[]"} {
		raw, _ := json.Marshal(command(0))
		_, err := parseCommand("CreateBatch", strings.Replace(string(raw), `"quantityGrams":100000`, `"quantityGrams":`+replacement, 1))
		code(t, err, "INVALID_SCHEMA")
	}
	c := command(0)
	raw, _ := json.Marshal(c)
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	for k := range obj {
		copy := map[string]any{}
		for key, v := range obj {
			if key != k {
				copy[key] = v
			}
		}
		b, _ := json.Marshal(copy)
		_, err := parseCommand(c.Command, string(b))
		code(t, err, "INVALID_SCHEMA")
	}
	var payload map[string]any
	_ = json.Unmarshal(c.Payload, &payload)
	for k := range payload {
		copy := map[string]any{}
		for key, v := range payload {
			if key != k {
				copy[key] = v
			}
		}
		bad := c
		bad.Payload, _ = json.Marshal(copy)
		b, _ := json.Marshal(bad)
		_, err := parseCommand(c.Command, string(b))
		code(t, err, "INVALID_SCHEMA")
	}
}
func TestInputSchemas(t *testing.T) {
	for _, tc := range []struct{ from, to, want string }{
		{`agrochain.command.v1`, `agrochain.command.v2`, "UNSUPPORTED_SCHEMA_VERSION"},
		{`BAT-NORMAL01`, `BAT-x`, "INVALID_IDENTIFIER"},
		{`OP-STEP0000`, `OP-INVALID!`, "INVALID_IDENTIFIER"},
		{`"expectedVersion":0`, `"expectedVersion":-1`, "VERSION_CONFLICT"},
		{`"quantityGrams":100000`, `"quantityGrams":0`, "INVALID_QUANTITY"},
		{`"quantityGrams":100000`, `"quantityGrams":-1`, "INVALID_QUANTITY"},
		{`"quantityGrams":100000`, `"quantityGrams":100000001`, "INVALID_QUANTITY"},
		{`"quantityGrams":100000`, `"quantityGrams":"100000"`, "INVALID_SCHEMA"},
		{`TOMATO`, `POTATO`, "INVALID_SCHEMA"}, {`STANDARD`, `PREMIUM`, "INVALID_SCHEMA"},
		{`2026-09-15`, `2026-02-30`, "INVALID_SCHEMA"}, {`"07"`, `"82"`, "INVALID_SCHEMA"},
		{`"payload":{`, `"payload":{"unit":"kg",`, "INVALID_SCHEMA"},
		{`"payload":{`, `"payload":{"ownerMsp":"RetailerMSP",`, "INVALID_SCHEMA"},
		{`"payload":{`, `"payload":{"totalKurus":1,`, "INVALID_SCHEMA"},
	} {
		t.Run(tc.to, func(t *testing.T) {
			b, _ := json.Marshal(command(0))
			_, err := parseCommand("CreateBatch", strings.Replace(string(b), tc.from, tc.to, 1))
			code(t, err, tc.want)
		})
	}
}
func TestCanonicalSerialization(t *testing.T) {
	a := map[string]any{"z": int64(100), "a": map[string]any{"c": "<>&", "b": int64(0)}}
	b, err := canonical(a)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"a":{"b":0,"c":"<>&"},"z":100}` {
		t.Fatal(string(b))
	}
	for i := 0; i < 100; i++ {
		got, _ := canonical(map[string]any{"a": a["a"], "z": int64(100)})
		if string(got) != string(b) {
			t.Fatal("unstable bytes")
		}
	}
	if !reflect.DeepEqual(sortedUnique([]string{"b", "a", "b"}), []string{"a", "b"}) {
		t.Fatal("unstable IDs")
	}
}
func TestRoles(t *testing.T) {
	code(t, authorize(Producer, "", "CreateBatch"), "UNAUTHORIZED_ROLE")
	for i := range steps {
		if err := authorize(steps[i].actor, "admin", steps[i].fn); err == nil {
			t.Fatal("admin bypass")
		}
	}
	for _, p := range []struct{ m, r string }{{Producer, "producer"}, {Logistics, "carrier"}, {Retailer, "retailer"}, {Regulator, "auditor"}, {Regulator, "reviewer"}, {Regulator, "oracle"}, {Regulator, "public-reader"}} {
		if err := authorize(p.m, p.r, "GetBatch"); err != nil {
			t.Fatal(err)
		}
	}
	code(t, authorize(Producer, "auditor", "GetBatch"), "UNAUTHORIZED_ROLE")
}
func TestQueries(t *testing.T) {
	m, _ := setup(t, 7)
	for _, tc := range []struct{ fn, id string }{{"GetBatch", "BAT-NORMAL01"}, {"GetTransfer", "TRF-PICKUP01"}, {"GetOperation", "OP-STEP0000"}, {"BatchExists", "BAT-NORMAL01"}} {
		if _, err := query(m, Producer, tc.fn, []string{tc.id}); err != nil {
			t.Fatal(err)
		}
	}
	var got int
	mark := ""
	for {
		raw, _ := json.Marshal(pageInput{"agrochain.page-request.v1", "BAT-NORMAL01", 2, mark})
		v, err := query(m, Producer, "GetBatchHistory", []string{string(raw)})
		if err != nil {
			t.Fatal(err)
		}
		p := v.(Page)
		got += len(p.Records)
		mark = p.Bookmark
		if mark == "" {
			break
		}
	}
	if got != 7 {
		t.Fatal("history pagination", got)
	}
	for _, tc := range []struct {
		fn, filter string
		want       int
	}{{"QueryBatchesByOwner", Producer, 0}, {"QueryBatchesByOwner", Retailer, 1}, {"QueryBatchesByState", RetailReported, 1}, {"QueryBatchesByState", Created, 0}} {
		raw, _ := json.Marshal(pageInput{"agrochain.page-request.v1", tc.filter, 100, ""})
		v, err := query(m, Producer, tc.fn, []string{string(raw)})
		if err != nil || len(v.(Page).Records) != tc.want {
			t.Fatal(tc, err)
		}
	}
	for _, size := range []int32{-1, 0, 101} {
		raw, _ := json.Marshal(pageInput{"agrochain.page-request.v1", Created, size, ""})
		_, err := query(m, Producer, "QueryBatchesByState", []string{string(raw)})
		code(t, err, "INVALID_PAGE")
	}
	for _, tc := range []struct{ fn, id, want string }{{"GetBatch", "BAT-MISSING1", "BATCH_NOT_FOUND"}, {"GetTransfer", "TRF-MISSING1", "TRANSFER_NOT_FOUND"}, {"GetOperation", "OP-MISSING1", "OPERATION_NOT_FOUND"}, {"GetBatch", "bad", "INVALID_IDENTIFIER"}} {
		_, err := query(m, Producer, tc.fn, []string{tc.id})
		code(t, err, tc.want)
	}
}
func TestErrorStability(t *testing.T) {
	for name, msg := range messages {
		err := fail(name)
		code(t, err, name)
		if !strings.Contains(err.Error(), msg) {
			t.Fatal("message drift")
		}
	}
	code(t, fail("untrusted arbitrary error"), "INTERNAL_ERROR")
	m, e := setup(t, 0)
	m.failRead = true
	_, err := run(e, command(0), context(0, Producer))
	code(t, err, "INTERNAL_ERROR")
	m, e = setup(t, 0)
	m.failWrite = true
	_, err = run(e, command(0), context(0, Producer))
	code(t, err, "INTERNAL_ERROR")
	if len(m.events) > 0 {
		t.Fatal("event after failure")
	}
}
