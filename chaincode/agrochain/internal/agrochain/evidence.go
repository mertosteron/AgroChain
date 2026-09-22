package agrochain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"time"
)

const trade = "tradePrivate"
const freight = "freightPrivate"
const audit = "retailAuditPrivate"

type privateLedger interface {
	ledger
	readPrivate(string, string) ([]byte, error)
	writePrivate(string, string, []byte) error
}
type privateWrite struct {
	Collection, Key string
	Value           []byte
}
type sourceKey struct {
	SchemaVersion string   `json:"schemaVersion"`
	IssuerID      string   `json:"issuerId"`
	SourceSystem  string   `json:"sourceSystem"`
	DocumentTypes []string `json:"documentTypes"`
	KeyID         string   `json:"keyId"`
	PublicKey     string   `json:"publicKeyRawBase64url"`
	Enabled       bool     `json:"enabled"`
}
type configuration struct {
	SchemaVersion string      `json:"schemaVersion"`
	PolicyID      string      `json:"policyId"`
	ThresholdBps  int64       `json:"thresholdBps"`
	Sources       []sourceKey `json:"sources"`
}
type header struct {
	SchemaVersion       string `json:"schemaVersion"`
	DocumentID          string `json:"documentId"`
	SourceSystem        string `json:"sourceSystem"`
	SourceMode          string `json:"sourceMode"`
	IssuerID            string `json:"issuerId"`
	KeyID               string `json:"keyId"`
	SourceDocumentID    string `json:"sourceDocumentId"`
	DocumentType        string `json:"documentType"`
	BatchID             string `json:"batchId"`
	OperationID         string `json:"boundOperationId"`
	IssuedAt            string `json:"issuedAt"`
	Nonce               string `json:"nonce"`
	CommitmentAlgorithm string `json:"commitmentAlgorithm"`
	Commitment          string `json:"commitment"`
	SignatureAlgorithm  string `json:"signatureAlgorithm"`
}
type envelope struct {
	Header    header `json:"header"`
	Signature string `json:"signature"`
}
type bundle struct {
	Envelope envelope        `json:"envelope"`
	Body     json.RawMessage `json:"body"`
	Salt     string          `json:"saltHex"`
}
type opening struct {
	Body json.RawMessage `json:"body"`
	Salt string          `json:"saltHex"`
}

var reference = regexp.MustCompile(`^[A-Z0-9_]{1,64}$`)

// Current source schemas contain only bounded ASCII fields and integer values.
// canonical() is JCS-compatible on this validated subset, never general JSON.
func exact(raw []byte, dst any) error {
	t := reflect.TypeOf(dst).Elem()
	fields := []string{}
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Tag.Get("json"))
	}
	return decode(raw, dst, fields...)
}
func encodedExact(v any, dst any) error {
	raw, err := canonical(v)
	if err != nil {
		return err
	}
	return exact(raw, dst)
}
func rawURL(v string, n int) ([]byte, error) {
	b, e := base64.RawURLEncoding.Strict().DecodeString(v)
	if e != nil || len(b) != n || base64.RawURLEncoding.EncodeToString(b) != v {
		return nil, fail("INVALID_SCHEMA")
	}
	return b, nil
}
func validDateTime(v string) bool {
	t, e := time.Parse("2006-01-02T15:04:05.000Z", v)
	return e == nil && t.Format("2006-01-02T15:04:05.000Z") == v
}
func commitment(body json.RawMessage, salt string) (string, error) {
	if !hexID.MatchString(salt) {
		return "", fail("INVALID_SCHEMA")
	}
	v, e := strictJSON(body)
	if e != nil {
		return "", e
	}
	b, e := canonical(v)
	if e != nil {
		return "", e
	}
	s, _ := hex.DecodeString(salt)
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(b)))
	h := sha256.New()
	h.Write([]byte("AgroChain/document/v1\x00"))
	h.Write(s)
	h.Write(size[:])
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}
func getConfig(s ledger) (configuration, error) {
	var c configuration
	found, e := load(s, key("configuration", "v1"), &c)
	if e != nil {
		return c, e
	}
	if !found {
		return c, fail("CONFIGURATION_REQUIRED")
	}
	return c, nil
}
func bootstrap(s ledger, x execution, raw []byte) (any, error) {
	if e := authorize(x.Actor, x.Role, "Bootstrap"); e != nil {
		return nil, e
	}
	if e := absent(s, key("configuration", "v1"), "CONFIGURATION_ALREADY_EXISTS"); e != nil {
		return nil, e
	}
	var c configuration
	if e := exact(raw, &c); e != nil {
		return nil, e
	}
	if c.SchemaVersion != "agrochain.bootstrap.v1" || !validID(c.PolicyID, "CFG") || c.ThresholdBps < 0 || c.ThresholdBps > 100000 || len(c.Sources) != 4 {
		return nil, fail("INVALID_SCHEMA")
	}
	types := map[string][]string{"CKS": {"PRODUCER_ELIGIBILITY"}, "EFATURA": {"FREIGHT_INVOICE", "PURCHASE_INVOICE"}, "HKS": {"TRADE_NOTIFICATION"}, "UETDS": {"TRANSPORT_MANIFEST"}}
	seen := map[string]bool{}
	seenKeys := map[string]bool{}
	plan := writePlan{}
	last := ""
	for _, k := range c.Sources {
		registryKey := key("sourceKey", k.IssuerID, k.KeyID)
		if seenKeys[registryKey] {
			return nil, fail("INVALID_SCHEMA")
		}
		seenKeys[registryKey] = true
		if e := encodedExact(k, &sourceKey{}); e != nil {
			return nil, e
		}
		if k.SchemaVersion != "agrochain.source-key.v1" || !reference.MatchString(k.IssuerID) || !reference.MatchString(k.KeyID) || !k.Enabled || !reflect.DeepEqual(k.DocumentTypes, types[k.SourceSystem]) || seen[k.SourceSystem] || k.SourceSystem <= last {
			return nil, fail("INVALID_SCHEMA")
		}
		if _, e := rawURL(k.PublicKey, 32); e != nil {
			return nil, e
		}
		seen[k.SourceSystem] = true
		last = k.SourceSystem
		if e := plan.put(key("sourceKey", k.IssuerID, k.KeyID), struct {
			sourceKey
			CreatedTxID string `json:"createdTxId"`
		}{k, x.TxID}); e != nil {
			return nil, e
		}
	}
	// Validate nested source objects before Unmarshal can discard unknown fields.
	v, _ := strictJSON(raw)
	for _, item := range v.(map[string]any)["sources"].([]any) {
		if e := encodedExact(item, &sourceKey{}); e != nil {
			return nil, e
		}
	}
	if e := plan.put(key("configuration", "v1"), c); e != nil {
		return nil, e
	}
	policy := map[string]any{"schemaVersion": "agrochain.price-policy.v1", "policyId": c.PolicyID, "ruleVersion": "price-increase-v1", "thresholdBps": c.ThresholdBps, "currency": "TRY", "taxBasis": "EXCLUDING_TAX", "comparison": "STRICT_GREATER", "createdTxId": x.TxID}
	if e := plan.put(key("policy", c.PolicyID), policy); e != nil {
		return nil, e
	}
	keys := []string{}
	for k := range plan {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if e := s.write(k, plan[k]); e != nil {
			return nil, fail("INTERNAL_ERROR")
		}
	}
	return c, nil
}
func verifyBundle(s ledger, b bundle) error {
	h := b.Envelope.Header
	if h.SchemaVersion != "agrochain.source-document.v1" || h.SourceMode != "SIMULATED" || !validID(h.DocumentID, "DOC") || !validID(h.SourceDocumentID, "DOC") || !validID(h.BatchID, "BAT") || !validID(h.OperationID, "OP") || !reference.MatchString(h.IssuerID) || !reference.MatchString(h.KeyID) || !hexID.MatchString(h.Nonce) || !hexID.MatchString(h.Commitment) || !validDateTime(h.IssuedAt) || h.CommitmentAlgorithm != "SHA256_SALTED_JCS_V1" || h.SignatureAlgorithm != "Ed25519" {
		return fail("INVALID_SCHEMA")
	}
	var k sourceKey
	found, e := load(s, key("sourceKey", h.IssuerID, h.KeyID), &k)
	if e != nil {
		return e
	}
	if !found || !k.Enabled || k.SourceSystem != h.SourceSystem {
		return fail("UNTRUSTED_SOURCE_KEY")
	}
	allowed := false
	for _, t := range k.DocumentTypes {
		if t == h.DocumentType {
			allowed = true
		}
	}
	if !allowed {
		return fail("UNTRUSTED_SOURCE_KEY")
	}
	pub, e := rawURL(k.PublicKey, 32)
	if e != nil {
		return fail("UNTRUSTED_SOURCE_KEY")
	}
	sig, e := rawURL(b.Envelope.Signature, 64)
	if e != nil {
		return fail("INVALID_SOURCE_SIGNATURE")
	}
	msg, e := canonical(h)
	if e != nil {
		return e
	}
	if !ed25519.Verify(pub, append([]byte("AgroChain/source-signature/v1\x00"), msg...), sig) {
		return fail("INVALID_SOURCE_SIGNATURE")
	}
	c, e := commitment(b.Body, b.Salt)
	if e != nil {
		return e
	}
	if c != h.Commitment {
		return fail("DOCUMENT_HASH_MISMATCH")
	}
	return nil
}
func decodeBundles(raw []byte) ([]bundle, error) {
	v, e := strictJSON(raw)
	if e != nil {
		return nil, e
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fail("INVALID_SCHEMA")
	}
	out := []bundle{}
	last := ""
	for _, item := range arr {
		var b bundle
		if e := encodedExact(item, &b); e != nil {
			return nil, e
		}
		obj := item.(map[string]any)
		var env envelope
		if e := encodedExact(obj["envelope"], &env); e != nil {
			return nil, e
		}
		if e := encodedExact(obj["envelope"].(map[string]any)["header"], &header{}); e != nil {
			return nil, e
		}
		if b.Envelope.Header.DocumentID <= last {
			return nil, fail("INVALID_SCHEMA")
		}
		last = b.Envelope.Header.DocumentID
		out = append(out, b)
	}
	return out, nil
}
