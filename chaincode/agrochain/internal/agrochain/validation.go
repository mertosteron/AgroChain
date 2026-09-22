package agrochain

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var messages = map[string]string{
	"INVALID_SCHEMA":                    "Invalid fields or JSON encoding.",
	"INVALID_IDENTIFIER":                "Invalid identifier.",
	"UNSUPPORTED_SCHEMA_VERSION":        "Unsupported schema version.",
	"INVALID_QUANTITY":                  "Quantity must match the positive whole batch.",
	"INVALID_MONEY":                     "Invalid integer monetary value, currency or tax basis.",
	"UNSUPPORTED_PILOT_OPERATION":       "This operation is not supported by the pilot.",
	"UNAUTHORIZED_ORGANIZATION":         "Organization is not permitted for this operation.",
	"UNAUTHORIZED_ROLE":                 "Certificate role is not permitted for this operation.",
	"WRONG_TRANSFER_RECIPIENT":          "Only the intended recipient may accept this transfer.",
	"BATCH_NOT_FOUND":                   "Batch does not exist.",
	"TRANSFER_NOT_FOUND":                "Transfer does not exist for this batch.",
	"OPERATION_NOT_FOUND":               "Operation does not exist within caller visibility.",
	"BATCH_ALREADY_EXISTS":              "Batch already exists.",
	"TRANSFER_ALREADY_EXISTS":           "Transfer identifier already exists.",
	"LOT_ALREADY_EXISTS":                "Retail lot already exists.",
	"COST_ALREADY_EXISTS":               "Freight cost already exists.",
	"REPORT_ALREADY_EXISTS":             "Retail report already exists.",
	"DUPLICATE_TRANSACTION":             "Operation already committed; query the original receipt.",
	"DOCUMENT_ALREADY_USED":             "Document reference already consumed.",
	"INVALID_STATE_TRANSITION":          "Operation is not allowed in the current state.",
	"VERSION_CONFLICT":                  "Expected batch version does not match.",
	"MISSING_EVIDENCE":                  "Required verified evidence is missing.",
	"SOURCE_BINDING_MISMATCH":           "Verified evidence does not match the batch.",
	"POLICY_MISMATCH":                   "Report policy does not match the verified policy.",
	"INVALID_PAGE":                      "Invalid page size or bookmark.",
	"EVIDENCE_VERIFICATION_UNAVAILABLE": "Stage 3 writes are disabled pending Stage 4 evidence and privacy verification.",
	"INTERNAL_ERROR":                    "Ledger operation failed.",
	"CONFIGURATION_REQUIRED":            "Initialize trusted sources before business operations.",
	"CONFIGURATION_ALREADY_EXISTS":      "Trust configuration is immutable and already initialized.",
	"UNTRUSTED_SOURCE_KEY":              "Source key is not enabled for this claim.",
	"INVALID_SOURCE_SIGNATURE":          "Source signature verification failed.",
	"DOCUMENT_HASH_MISMATCH":            "Document does not match committed evidence.",
	"SOURCE_NONCE_ALREADY_USED":         "Source nonce has already been consumed.",
	"SOURCE_CLAIM_REJECTED":             "Source assertion does not authorize this operation.",
	"PRIVATE_DATA_ACCESS_DENIED":        "Identity cannot access this private record.",
	"PRIVATE_DATA_UNAVAILABLE":          "Required private bytes are unavailable on this peer.",
	"DOCUMENT_NOT_FOUND":                "Document does not exist.",
}

type DomainError struct {
	SchemaVersion string `json:"schemaVersion"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	Retryable     bool   `json:"retryable"`
}

func (e *DomainError) Error() string { b, _ := json.Marshal(e); return string(b) }
func fail(code string) error {
	message, ok := messages[code]
	if !ok {
		code, message = "INTERNAL_ERROR", messages["INTERNAL_ERROR"]
	}
	return &DomainError{"agrochain.error.v1", code, message, false}
}

var identifier = regexp.MustCompile(`^[A-Z]+-[A-Z0-9]{8,32}$`)
var integer = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
var hexID = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validID(s, prefix string) bool {
	return strings.HasPrefix(s, prefix+"-") && identifier.MatchString(s)
}
func member(m string) bool { return m == Producer || m == Logistics || m == Retailer || m == Regulator }
func validState(s string) bool {
	switch s {
	case Created, PickupPending, InTransport, DeliveryPending, Received, RetailReported:
		return true
	}
	return false
}

// All Stage 3 public input strings have ASCII-only grammars. Rejecting non-ASCII
// also rejects non-NFC, lone escaped surrogates, and replacement characters.
func strictJSON(raw []byte) (any, error) {
	if len(raw) == 0 || len(raw) > 65536 || !utf8.Valid(raw) {
		return nil, fail("INVALID_SCHEMA")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	v, err := tokenValue(d, 0)
	if err != nil {
		return nil, fail("INVALID_SCHEMA")
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, fail("INVALID_SCHEMA")
	}
	return v, nil
}
func tokenValue(d *json.Decoder, depth int) (any, error) {
	if depth > 8 {
		return nil, fail("INVALID_SCHEMA")
	}
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch x := t.(type) {
	case json.Delim:
		if x == '[' {
			a := []any{}
			for d.More() {
				if len(a) >= 16 {
					return nil, fail("INVALID_SCHEMA")
				}
				v, err := tokenValue(d, depth+1)
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return nil, fail("INVALID_SCHEMA")
			}
			return a, nil
		}
		if x != '{' {
			return nil, fail("INVALID_SCHEMA")
		}
		m := map[string]any{}
		for d.More() {
			kt, err := d.Token()
			if err != nil {
				return nil, err
			}
			k, ok := kt.(string)
			if !ok || !safeString(k) || k == "__proto__" || k == "constructor" || k == "prototype" || len(m) >= 32 {
				return nil, fail("INVALID_SCHEMA")
			}
			if _, exists := m[k]; exists {
				return nil, fail("INVALID_SCHEMA")
			}
			v, err := tokenValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return nil, fail("INVALID_SCHEMA")
		}
		return m, nil
	case json.Number:
		if !integer.MatchString(string(x)) {
			return nil, fail("INVALID_SCHEMA")
		}
		if _, err := strconv.ParseInt(string(x), 10, 64); err != nil {
			return nil, fail("INVALID_SCHEMA")
		}
		return x, nil
	case string:
		if !safeString(x) {
			return nil, fail("INVALID_SCHEMA")
		}
		return x, nil
	case bool:
		return x, nil
	default:
		return nil, fail("INVALID_SCHEMA")
	}
}
func safeString(s string) bool {
	if len(s) > 512 {
		return false
	}
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}
	return true
}
func decode(raw []byte, dst any, fields ...string) error {
	v, err := strictJSON(raw)
	if err != nil {
		return err
	}
	m, ok := v.(map[string]any)
	if !ok || len(m) != len(fields) {
		return fail("INVALID_SCHEMA")
	}
	for _, k := range fields {
		if _, ok := m[k]; !ok {
			return fail("INVALID_SCHEMA")
		}
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fail("INVALID_SCHEMA")
	}
	return nil
}
func parseCommand(fn, raw string) (Command, error) {
	var c Command
	err := decode([]byte(raw), &c, "schemaVersion", "operationId", "batchId", "expectedVersion", "command", "payload")
	if err != nil {
		return c, err
	}
	if c.SchemaVersion != "agrochain.command.v1" {
		return c, fail("UNSUPPORTED_SCHEMA_VERSION")
	}
	if c.Command != fn {
		return c, fail("INVALID_SCHEMA")
	}
	if !validID(c.BatchID, "BAT") || !validID(c.OperationID, "OP") {
		return c, fail("INVALID_IDENTIFIER")
	}
	if c.ExpectedVersion < 0 || c.ExpectedVersion > 2147483647 {
		return c, fail("VERSION_CONFLICT")
	}
	return c, validatePayload(c)
}
func validatePayload(c Command) error {
	switch c.Command {
	case "CreateBatch":
		var p createInput
		if err := decode(c.Payload, &p, "productCode", "gradeCode", "quantityGrams", "originRegionCode", "harvestDate", "logisticsMsp", "intendedRetailerMsp"); err != nil {
			return err
		}
		if p.QuantityGrams < 1 || p.QuantityGrams > 100000000 {
			return fail("INVALID_QUANTITY")
		}
		if p.ProductCode != "TOMATO" || p.GradeCode != "STANDARD" || p.LogisticsMSP != Logistics || p.IntendedRetailerMSP != Retailer {
			return fail("INVALID_SCHEMA")
		}
		r, err := strconv.Atoi(p.OriginRegionCode)
		if err != nil || len(p.OriginRegionCode) != 2 || r < 1 || r > 81 {
			return fail("INVALID_SCHEMA")
		}
		t, err := time.Parse("2006-01-02", p.HarvestDate)
		if err != nil || t.Format("2006-01-02") != p.HarvestDate {
			return fail("INVALID_SCHEMA")
		}
	case "OfferPickup", "AcceptPickup", "OfferDelivery", "AcceptDelivery":
		var p handoffInput
		if err := decode(c.Payload, &p, "transferId", "quantityGrams"); err != nil {
			return err
		}
		if !validID(p.TransferID, "TRF") {
			return fail("INVALID_IDENTIFIER")
		}
		if p.QuantityGrams < 1 || p.QuantityGrams > 100000000 {
			return fail("INVALID_QUANTITY")
		}
	case "RecordFreightCost":
		var p freightInput
		if err := decode(c.Payload, &p, "costId"); err != nil {
			return err
		}
		if !validID(p.CostID, "CST") {
			return fail("INVALID_IDENTIFIER")
		}
	case "ReportRetailPrice":
		var p reportInput
		if err := decode(c.Payload, &p, "reportId", "retailLotId", "policyId"); err != nil {
			return err
		}
		if !validID(p.ReportID, "RPT") || !validID(p.RetailLotID, "LOT") || !validID(p.PolicyID, "CFG") {
			return fail("INVALID_IDENTIFIER")
		}
	default:
		return fail("UNSUPPORTED_PILOT_OPERATION")
	}
	return nil
}

func authorize(actor, role, fn string) error {
	if !member(actor) {
		return fail("UNAUTHORIZED_ORGANIZATION")
	}
	expected, expectedRole := "", ""
	switch fn {
	case "Bootstrap":
		expected, expectedRole = Regulator, "admin"
	case "CreateBatch", "OfferPickup":
		expected, expectedRole = Producer, "producer"
	case "AcceptPickup", "RecordFreightCost", "OfferDelivery":
		expected, expectedRole = Logistics, "carrier"
	case "AcceptDelivery", "ReportRetailPrice":
		expected, expectedRole = Retailer, "retailer"
	}
	if expected != "" && actor != expected {
		if fn == "AcceptPickup" || fn == "AcceptDelivery" {
			return fail("WRONG_TRANSFER_RECIPIENT")
		}
		return fail("UNAUTHORIZED_ORGANIZATION")
	}
	// Health exposes no business data and remains available to channel members.
	if fn == "Health" {
		return nil
	}
	if role == "" {
		return fail("UNAUTHORIZED_ROLE")
	}
	if role != "" {
		if expectedRole != "" && role != expectedRole {
			return fail("UNAUTHORIZED_ROLE")
		}
		if expectedRole == "" {
			allowed := map[string][]string{Producer: {"producer"}, Logistics: {"carrier"}, Retailer: {"retailer"}, Regulator: {"reviewer", "auditor", "oracle", "public-reader"}}
			ok := false
			for _, r := range allowed[actor] {
				if r == role {
					ok = true
				}
			}
			if !ok {
				return fail("UNAUTHORIZED_ROLE")
			}
		}
	}
	return nil
}

// Canonical ordering for typed public records. All strings are ASCII and all
// numbers integers; the restricted encoding is JCS-compatible. Stage 4 must add
// the full cross-language cryptographic canonicalization suite before hashing.
func canonical(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fail("INTERNAL_ERROR")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var x any
	if err := d.Decode(&x); err != nil {
		return nil, fail("INTERNAL_ERROR")
	}
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if err := e.Encode(x); err != nil {
		return nil, fail("INTERNAL_ERROR")
	}
	return bytes.TrimSuffix(out.Bytes(), []byte{'\n'}), nil
}
func sortedUnique(ids []string) []string {
	sort.Strings(ids)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if len(out) == 0 || out[len(out)-1] != id {
			out = append(out, id)
		}
	}
	return out
}
