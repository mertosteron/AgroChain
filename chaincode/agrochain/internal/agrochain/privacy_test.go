package agrochain

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type privateMemory struct {
	*memory
	private map[string][]byte
}

func (m *privateMemory) readPrivate(c, k string) ([]byte, error)  { return m.private[c+k], nil }
func (m *privateMemory) writePrivate(c, k string, v []byte) error { m.private[c+k] = v; return nil }
func salt() string                                                { b := make([]byte, 32); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func bytesOf(v any) []byte                                        { b, _ := canonical(v); return b }
func privacySetup(t *testing.T) (*privateMemory, map[string]ed25519.PrivateKey) {
	t.Helper()
	m := &privateMemory{newMemory(), map[string][]byte{}}
	keys := map[string]ed25519.PrivateKey{}
	cfg := configuration{"agrochain.bootstrap.v1", "CFG-PRICE001", 5000, nil}
	for _, sys := range []string{"CKS", "EFATURA", "HKS", "UETDS"} {
		pub, k, _ := ed25519.GenerateKey(rand.Reader)
		keys[sys] = k
		types := map[string][]string{"CKS": {"PRODUCER_ELIGIBILITY"}, "EFATURA": {"FREIGHT_INVOICE", "PURCHASE_INVOICE"}, "HKS": {"TRADE_NOTIFICATION"}, "UETDS": {"TRANSPORT_MANIFEST"}}
		cfg.Sources = append(cfg.Sources, sourceKey{"agrochain.source-key.v1", "SIM_" + sys, sys, types[sys], "KEY_" + sys, base64.RawURLEncoding.EncodeToString(pub), true})
	}
	x := context(50, Regulator)
	x.Role = "admin"
	if _, err := bootstrap(m, x, bytesOf(cfg)); err != nil {
		t.Fatal(err)
	}
	return m, keys
}
func signed(sys, kind, id string, body any, c Command, keys map[string]ed25519.PrivateKey) bundle {
	b := bundle{Body: bytesOf(body), Salt: salt()}
	commit, _ := commitment(b.Body, b.Salt)
	b.Envelope.Header = header{"agrochain.source-document.v1", id, sys, "SIMULATED", "SIM_" + sys, "KEY_" + sys, id, kind, c.BatchID, c.OperationID, "2026-09-21T09:00:00.000Z", salt(), "SHA256_SALTED_JCS_V1", commit, "Ed25519"}
	b.Envelope.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(keys[sys], append([]byte("AgroChain/source-signature/v1\x00"), bytesOf(b.Envelope.Header)...)))
	return b
}
func transientFor(i int, c Command, keys map[string]ed25519.PrivateKey) map[string][]byte {
	out := map[string][]byte{}
	docs := []bundle{}
	switch i {
	case 0:
		docs = append(docs, signed("CKS", "PRODUCER_ELIGIBILITY", "DOC-CKS00001", map[string]any{"producerMsp": Producer, "productCode": "TOMATO", "gradeCode": "STANDARD", "quantityGrams": 100000, "originRegionCode": "07", "harvestDate": "2026-09-15", "eligible": true}, c, keys))
	case 1:
		docs = append(docs, signed("EFATURA", "PURCHASE_INVOICE", "DOC-BUY00001", purchaseBody{Producer, Retailer, 100000, 2000, 200000, "TRY", "EXCLUDING_TAX", strings.Repeat("a", 64), "application/xml"}, c, keys))
		docs = append(docs, signed("HKS", "TRADE_NOTIFICATION", "DOC-HKS00001", map[string]any{"producerMsp": Producer, "retailerMsp": Retailer, "productCode": "TOMATO", "quantityGrams": 100000, "notified": true}, c, keys))
		docs = append(docs, signed("UETDS", "TRANSPORT_MANIFEST", "DOC-MAN00001", map[string]any{"carrierMsp": Logistics, "consignorMsp": Producer, "consigneeMsp": Retailer, "productCode": "TOMATO", "quantityGrams": 100000, "declaredDepartureAt": "2026-09-21T09:00:00.000Z"}, c, keys))
	case 3:
		docs = append(docs, signed("EFATURA", "FREIGHT_INVOICE", "DOC-FREIGHT1", freightBody{Logistics, Retailer, 100000, 20000, "TRY", "EXCLUDING_TAX", strings.Repeat("b", 64), "application/xml"}, c, keys))
		out["recordSaltHex"] = bytesOf(salt())
	case 6:
		out["retailReportInput"] = bytesOf(retailInput{2800, "TRY", "EXCLUDING_TAX", "2026-09-21T09:30:00.000Z", salt()})
	}
	if len(docs) > 0 {
		sort.Slice(docs, func(i, j int) bool { return docs[i].Envelope.Header.DocumentID < docs[j].Envelope.Header.DocumentID })
		out["evidence"] = bytesOf(docs)
	}
	return out
}
func TestVerifiedPrivateLifecycle(t *testing.T) {
	m, keys := privacySetup(t)
	for i, step := range steps {
		c := command(i)
		x := context(i, step.actor)
		_, err := dispatch(m, x.Actor, x.Role, c.Command, []string{string(bytesOf(c))}, proposal{x, transientFor(i, c, keys)})
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	if batch(t, m.memory).State != RetailReported || len(m.private) != 4 {
		t.Fatal("missing public/private records")
	}
	for _, fn := range []string{"GetPurchase", "GetFreightCost", "GetRetailReport"} {
		for _, actor := range []string{Producer, Logistics, Retailer, Regulator} {
			x := context(70, actor)
			_, err := privateQuery(m, actor, x.Role, fn, "BAT-NORMAL01", nil)
			collection := map[string]string{"GetPurchase": trade, "GetFreightCost": freight, "GetRetailReport": audit}[fn]
			if mayRead(actor, x.Role, collection) {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				code(t, err, "PRIVATE_DATA_ACCESS_DENIED")
			}
		}
	}
	_, err := privateQuery(m, Regulator, "public-reader", "GetPurchase", "BAT-NORMAL01", nil)
	code(t, err, "PRIVATE_DATA_ACCESS_DENIED")
	var buy opening
	_ = privateLoad(m, trade, key("documentOpening", "DOC-BUY00001"), &buy)
	if _, err := verifyOpening(m, Producer, "producer", "DOC-BUY00001", map[string][]byte{"opening": bytesOf(buy)}); err != nil {
		t.Fatal(err)
	}
	buy.Salt = salt()
	_, err = verifyOpening(m, Producer, "producer", "DOC-BUY00001", map[string][]byte{"opening": bytesOf(buy)})
	code(t, err, "DOCUMENT_HASH_MISMATCH")
	delete(m.private, trade+key("documentOpening", "DOC-BUY00001"))
	_, err = privateQuery(m, Producer, "producer", "GetPurchase", "BAT-NORMAL01", nil)
	code(t, err, "PRIVATE_DATA_UNAVAILABLE")
	for _, raw := range m.values {
		for _, bad := range []string{"priceKurusPerKg", "totalKurus", "saltHex", "recordSaltHex", "attachmentDigest", "offeredPriceKurusPerKg"} {
			if strings.Contains(string(raw), bad) {
				t.Fatal("private field in shared state", bad)
			}
		}
	}
}
func TestEvidenceTamperingAndAtomicRejection(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		change     func(*bundle, map[string]ed25519.PrivateKey)
	}{
		{"salt", "DOCUMENT_HASH_MISMATCH", func(b *bundle, _ map[string]ed25519.PrivateKey) { b.Salt = salt() }},
		{"signature", "INVALID_SOURCE_SIGNATURE", func(b *bundle, _ map[string]ed25519.PrivateKey) {
			b.Envelope.Signature = base64.RawURLEncoding.EncodeToString(make([]byte, 64))
		}},
		{"header", "INVALID_SOURCE_SIGNATURE", func(b *bundle, _ map[string]ed25519.PrivateKey) { b.Envelope.Header.BatchID = "BAT-WRONG001" }},
		{"trusted binding", "SOURCE_BINDING_MISMATCH", func(b *bundle, k map[string]ed25519.PrivateKey) {
			b.Envelope.Header.OperationID = "OP-WRONG001"
			b.Envelope.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(k["CKS"], append([]byte("AgroChain/source-signature/v1\x00"), bytesOf(b.Envelope.Header)...)))
		}},
		{"unknown key", "UNTRUSTED_SOURCE_KEY", func(b *bundle, _ map[string]ed25519.PrivateKey) { b.Envelope.Header.KeyID = "UNKNOWN" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, k := privacySetup(t)
			c := command(0)
			trans := transientFor(0, c, k)
			docs, _ := decodeBundles(trans["evidence"])
			tc.change(&docs[0], k)
			trans["evidence"] = bytesOf(docs)
			before := m.clone()
			x := context(0, Producer)
			_, err := dispatch(m, x.Actor, x.Role, c.Command, []string{string(bytesOf(c))}, proposal{x, trans})
			code(t, err, tc.want)
			if !reflect.DeepEqual(before.values, m.values) || len(m.private) != 0 || len(m.events) != 0 {
				t.Fatal("rejected evidence wrote data")
			}
		})
	}
}
func TestSourceReplayAndBootstrapProtection(t *testing.T) {
	m, k := privacySetup(t)
	c := command(0)
	x := context(0, Producer)
	trans := transientFor(0, c, k)
	docs, _ := decodeBundles(trans["evidence"])
	if _, err := dispatch(m, x.Actor, x.Role, c.Command, []string{string(bytesOf(c))}, proposal{x, trans}); err != nil {
		t.Fatal(err)
	}
	_, err := dispatch(m, x.Actor, x.Role, c.Command, []string{string(bytesOf(c))}, proposal{x, trans})
	code(t, err, "DUPLICATE_TRANSACTION")
	for _, kind := range []string{"source", "nonce"} {
		c2 := command(0)
		c2.BatchID = "BAT-OTHER001"
		c2.OperationID = "OP-OTHER001"
		var body any
		_ = json.Unmarshal(docs[0].Body, &body)
		b := signed("CKS", "PRODUCER_ELIGIBILITY", "DOC-OTHER001", body, c2, k)
		if kind == "source" {
			b.Envelope.Header.SourceDocumentID = docs[0].Envelope.Header.SourceDocumentID
		} else {
			b.Envelope.Header.Nonce = docs[0].Envelope.Header.Nonce
		}
		b.Envelope.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(k["CKS"], append([]byte("AgroChain/source-signature/v1\x00"), bytesOf(b.Envelope.Header)...)))
		_, err := dispatch(m, x.Actor, x.Role, c2.Command, []string{string(bytesOf(c2))}, proposal{x, map[string][]byte{"evidence": bytesOf([]bundle{b})}})
		want := "DOCUMENT_ALREADY_USED"
		if kind == "nonce" {
			want = "SOURCE_NONCE_ALREADY_USED"
		}
		code(t, err, want)
	}
	_, err = bootstrap(m, x, []byte(`{}`))
	code(t, err, "UNAUTHORIZED_ORGANIZATION")
	x = context(0, Regulator)
	x.Role = "admin"
	_, err = bootstrap(m, x, []byte(`{}`))
	code(t, err, "CONFIGURATION_ALREADY_EXISTS")
}
func TestDocumentedCommitmentVector(t *testing.T) {
	body := json.RawMessage(`{"sellerMsp":"ProducerMSP","buyerMsp":"RetailerMSP","quantityGrams":100000,"priceKurusPerKg":2000,"totalKurus":200000,"currency":"TRY","taxBasis":"EXCLUDING_TAX","attachmentDigest":"a95bd76451b27eaa935bd2bac3586c94165d58a6622154395f9530cdfa0c5a56","attachmentMediaType":"application/xml"}`)
	got, err := commitment(body, "202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f")
	if err != nil || got != "4e1259fa1bf2ab64d37e9026c276fbe2dc06872f2deb800cbc87f83ffd15940a" {
		t.Fatal(got, err)
	}
	h := header{"agrochain.source-document.v1", "DOC-INVOICE01", "EFATURA", "SIMULATED", "SIM_EFATURA", "SIM_EFATURA_KEY01", "DOC-SOURCE001", "PURCHASE_INVOICE", "BAT-NORMAL01", "OP-OFFER001", "2026-09-16T09:00:00.000Z", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "SHA256_SALTED_JCS_V1", got, "Ed25519"}
	pub, _ := rawURL("njeHGEyz0EF6HP6iEDUHPftD5gwevFdrlCcOnJgTFkY", 32)
	sig, _ := rawURL("2LPDyrKB3jnVnrDMXr4_I1A1qA8R6Ty-69K8PS67kYM5u2x63GIr1Pm5LETIRAfW_xRSng2yyJraa8Ac41_5DQ", 64)
	if !ed25519.Verify(pub, append([]byte("AgroChain/source-signature/v1\x00"), bytesOf(h)...), sig) {
		t.Fatal("documented Ed25519 signature differs from Java vector")
	}
	h.BatchID = "BAT-CHANGED1"
	if ed25519.Verify(pub, append([]byte("AgroChain/source-signature/v1\x00"), bytesOf(h)...), sig) {
		t.Fatal("changed header verified")
	}
	for _, bad := range []string{`{"a":1.1}`, `{"a":1,"a":2}`, `{"a":"\ud800"}`, `{"a":null}`} {
		_, err := commitment([]byte(bad), salt())
		code(t, err, "INVALID_SCHEMA")
	}
}

func TestBootstrapRejectsCollidingRegistryKeys(t *testing.T) {
	m, _ := privacySetup(t)
	cfg, err := getConfig(m)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Sources[1].IssuerID = cfg.Sources[0].IssuerID
	cfg.Sources[1].KeyID = cfg.Sources[0].KeyID
	fresh := &privateMemory{newMemory(), map[string][]byte{}}
	x := context(0, Regulator)
	x.Role = "admin"
	_, err = bootstrap(fresh, x, bytesOf(cfg))
	code(t, err, "INVALID_SCHEMA")
	if len(fresh.values) != 0 {
		t.Fatal("invalid registry partially written")
	}
}
