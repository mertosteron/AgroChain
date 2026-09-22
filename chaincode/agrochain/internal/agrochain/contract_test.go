package agrochain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/attrmgr"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeStub struct {
	shim.ChaincodeStubInterface
	memory    *memory
	creator   []byte
	fn        string
	args      []string
	timestamp *timestamppb.Timestamp
	txid      string
	err       error
	iterator  *fakeIterator
	bookmark  string
}

func (s *fakeStub) GetCreator() ([]byte, error)                     { return s.creator, s.err }
func (s *fakeStub) GetTransient() (map[string][]byte, error)        { return nil, nil }
func (s *fakeStub) GetFunctionAndParameters() (string, []string)    { return s.fn, s.args }
func (s *fakeStub) GetTxTimestamp() (*timestamppb.Timestamp, error) { return s.timestamp, s.err }
func (s *fakeStub) GetTxID() string                                 { return s.txid }
func (s *fakeStub) GetState(k string) ([]byte, error)               { return s.memory.read(k) }
func (s *fakeStub) PutState(k string, v []byte) error               { return s.memory.write(k, v) }
func (s *fakeStub) DelState(k string) error                         { return s.memory.remove(k) }
func (s *fakeStub) SetEvent(n string, v []byte) error               { return s.memory.emit(n, v) }
func (s *fakeStub) GetStateByPartialCompositeKeyWithPagination(_ string, _ []string, _ int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	s.bookmark = bookmark
	return s.iterator, &peer.QueryResponseMetadata{Bookmark: key("owner", Producer, "BAT-NORMAL02")}, s.err
}

type fakeIterator struct {
	entries []*queryresult.KV
	index   int
	closed  bool
	err     error
}

func (i *fakeIterator) HasNext() bool { return i.index < len(i.entries) }
func (i *fakeIterator) Next() (*queryresult.KV, error) {
	if i.err != nil {
		return nil, i.err
	}
	v := i.entries[i.index]
	i.index++
	return v, nil
}
func (i *fakeIterator) Close() error { i.closed = true; return nil }
func identity(t *testing.T, org string) []byte {
	t.Helper()
	return identityWithAttributes(t, org, nil)
}
func identityWithAttributes(t *testing.T, org string, attributes []byte) []byte {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	c := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "User1", OrganizationalUnit: []string{"client"}}, NotBefore: time.Unix(0, 0), NotAfter: time.Unix(2000000000, 0)}
	if attributes != nil {
		c.ExtraExtensions = []pkix.Extension{{Id: attrmgr.AttrOID, Value: attributes}}
	}
	b, err := x509.CreateCertificate(rand.Reader, c, c, &k.PublicKey, k)
	if err != nil {
		t.Fatal(err)
	}
	serialized, err := proto.Marshal(&msp.SerializedIdentity{Mspid: org, IdBytes: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})})
	if err != nil {
		t.Fatal(err)
	}
	return serialized
}

func TestFabricCertificateRolesCannotEnableWrites(t *testing.T) {
	roles := map[string]string{Producer: "producer", Logistics: "carrier", Retailer: "retailer"}
	for i, step := range steps {
		for _, role := range []string{roles[step.actor], "admin", "oracle", "auditor"} {
			t.Run(step.fn+"/"+role, func(t *testing.T) {
				attributes, _ := json.Marshal(map[string]any{"attrs": map[string]string{"agrochain.role": role}})
				raw, _ := json.Marshal(command(i))
				s := &fakeStub{memory: newMemory(), creator: identityWithAttributes(t, step.actor, attributes), fn: step.fn, args: []string{string(raw)}, txid: strings.Repeat("b", 64), timestamp: &timestamppb.Timestamp{Seconds: 1789549200}}
				r := (&Contract{}).Invoke(s)
				want := "UNAUTHORIZED_ROLE"
				if role == roles[step.actor] {
					want = "CONFIGURATION_REQUIRED"
				}
				if r.Status != 500 || !strings.Contains(r.Message, want) || s.memory.writes != 0 || len(s.memory.events) != 0 {
					t.Fatal("certificate role bypassed deployment boundary", r)
				}
			})
		}
	}
	s := &fakeStub{memory: newMemory(), creator: identityWithAttributes(t, Producer, []byte(`{"attrs":`)), fn: "Health"}
	r := (&Contract{}).Invoke(s)
	if r.Status != 500 || !strings.Contains(r.Message, "UNAUTHORIZED_ORGANIZATION") {
		t.Fatal("malformed certificate attributes accepted", r)
	}
}
func TestFabricIdentityAndFailClosed(t *testing.T) {
	cc := &Contract{}
	for _, org := range []string{Producer, Logistics, Retailer, Regulator, "UnknownMSP"} {
		s := &fakeStub{memory: newMemory(), creator: identity(t, org), fn: "Health", txid: strings.Repeat("a", 64), timestamp: &timestamppb.Timestamp{Seconds: 1789549200, Nanos: 987654321}}
		r := cc.Invoke(s)
		if org == "UnknownMSP" {
			if r.Status != 500 || !strings.Contains(r.Message, "UNAUTHORIZED_ORGANIZATION") {
				t.Fatal(r)
			}
			continue
		}
		if r.Status != 200 {
			t.Fatal(r)
		}
		x, err := executionFromStub(s, org, "")
		if err != nil || x.TxID != s.txid || x.Time.Seconds != "1789549200" || x.Time.Nanos != 987654321 {
			t.Fatal(x, err)
		}
		s.fn = "CreateBatch"
		c, _ := json.Marshal(command(0))
		s.args = []string{string(c)}
		r = cc.Invoke(s)
		want := "UNAUTHORIZED_ORGANIZATION"
		if org == Producer {
			want = "UNAUTHORIZED_ROLE"
		}
		if r.Status != 500 || !strings.Contains(r.Message, want) || s.memory.writes != 0 {
			t.Fatal(r)
		}
	}
	s := &fakeStub{creator: []byte("not an identity"), memory: newMemory()}
	if cc.Invoke(s).Status != 500 {
		t.Fatal("bad identity accepted")
	}
	if cc.Init(s).Status != 500 {
		t.Fatal("init mutation enabled")
	}
	s.err = errors.New("private error")
	_, err := executionFromStub(s, Producer, "")
	code(t, err, "INTERNAL_ERROR")
	s.err = nil
	s.timestamp = &timestamppb.Timestamp{Seconds: -1}
	_, err = executionFromStub(s, Producer, "")
	code(t, err, "INTERNAL_ERROR")
}
func TestFabricPaginationAndLedgerBridge(t *testing.T) {
	s := &fakeStub{memory: newMemory(), iterator: &fakeIterator{entries: []*queryresult.KV{{Key: "k", Value: []byte(`"BAT-NORMAL01"`)}}}}
	f := fabricLedger{s}
	items, next, err := f.page("owner", []string{Producer}, 1, "")
	if err != nil || len(items) != 1 || next == "" || !s.iterator.closed {
		t.Fatal(items, next, err)
	}
	s.iterator = &fakeIterator{}
	_, _, err = f.page("owner", []string{Producer}, 1, next)
	if err != nil || s.bookmark != key("owner", Producer, "BAT-NORMAL02") {
		t.Fatal(err)
	}
	for _, bad := range []string{"not?base64", base64.RawURLEncoding.EncodeToString([]byte("owner:RetailerMSP:next")), base64.RawURLEncoding.EncodeToString([]byte("owner:ProducerMSP:" + key("history", "BAT-NORMAL01"))), next + "="} {
		_, _, err = f.page("owner", []string{Producer}, 1, bad)
		code(t, err, "INVALID_PAGE")
	}
	s.err = errors.New("secret")
	_, _, err = f.page("owner", []string{Producer}, 1, "")
	code(t, err, "INTERNAL_ERROR")
	s.err = nil
	s.iterator = &fakeIterator{entries: []*queryresult.KV{{Key: "k"}}, err: errors.New("secret")}
	_, _, err = f.page("owner", []string{Producer}, 1, "")
	code(t, err, "INTERNAL_ERROR")
	if !s.iterator.closed {
		t.Fatal("iterator leaked")
	}
	if err := f.write("x", []byte("value")); err != nil {
		t.Fatal(err)
	}
	if b, _ := f.read("x"); string(b) != "value" {
		t.Fatal("bad bridge")
	}
	if err := f.remove("x"); err != nil {
		t.Fatal(err)
	}
	e, _ := canonical(Event{SchemaVersion: "agrochain.event.v1"})
	if err := f.emit("Test", e); err != nil {
		t.Fatal(err)
	}
}
