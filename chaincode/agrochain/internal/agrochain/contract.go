package agrochain

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// Production always uses ledger trust and the real evidence verifier.
type Contract struct{}

func (*Contract) Init(shim.ChaincodeStubInterface) *peer.Response {
	return shim.Error(fail("UNSUPPORTED_PILOT_OPERATION").Error())
}
func (*Contract) Invoke(stub shim.ChaincodeStubInterface) *peer.Response {
	identity, err := cid.New(stub)
	if err != nil {
		return shim.Error(fail("UNAUTHORIZED_ORGANIZATION").Error())
	}
	msp, err := identity.GetMSPID()
	if err != nil {
		return shim.Error(fail("UNAUTHORIZED_ORGANIZATION").Error())
	}
	role, _, err := identity.GetAttributeValue("agrochain.role")
	if err != nil {
		return shim.Error(fail("UNAUTHORIZED_ROLE").Error())
	}
	fn, args := stub.GetFunctionAndParameters()
	x := execution{Actor: msp, Role: role}
	if isMutation(fn) || fn == "Bootstrap" {
		x, err = executionFromStub(stub, msp, role)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	transient, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fail("INVALID_SCHEMA").Error())
	}
	result, err := dispatch(fabricLedger{stub}, msp, role, fn, args, proposal{x, transient})
	if err != nil {
		return shim.Error(err.Error())
	}
	b, err := canonical(result)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(b)
}

func isMutation(fn string) bool {
	switch fn {
	case "CreateBatch", "OfferPickup", "AcceptPickup", "RecordFreightCost", "OfferDelivery", "AcceptDelivery", "ReportRetailPrice":
		return true
	}
	return false
}

// This context bridge is ready for engine.apply in Stage 4. Client command fields
// cannot provide these values. Fabric proposal time is signed proposal context,
// not an independent wall clock or proof of physical harvest/delivery time.
func executionFromStub(stub shim.ChaincodeStubInterface, msp, role string) (execution, error) {
	ts, err := stub.GetTxTimestamp()
	if err != nil || ts == nil || ts.Seconds < 0 || ts.Nanos < 0 || ts.Nanos > 999999999 || !hexID.MatchString(stub.GetTxID()) {
		return execution{}, fail("INTERNAL_ERROR")
	}
	return execution{Actor: msp, Role: role, TxID: stub.GetTxID(), Time: TxTime{Seconds: strconv.FormatInt(ts.Seconds, 10), Nanos: ts.Nanos}}, nil
}

type proposal struct {
	x         execution
	transient map[string][]byte
}

func dispatch(s ledger, msp, role, fn string, args []string, request ...proposal) (any, error) {
	if err := authorize(msp, role, fn); err != nil {
		return nil, err
	}
	switch fn {
	case "Bootstrap":
		if len(args) != 1 || len(request) != 1 || len(request[0].transient) != 0 {
			return nil, fail("INVALID_SCHEMA")
		}
		return bootstrap(s, request[0].x, []byte(args[0]))
	case "GetConfiguration":
		if len(args) != 0 {
			return nil, fail("INVALID_SCHEMA")
		}
		return getConfig(s)
	case "GetDocument":
		if len(args) != 1 || !validID(args[0], "DOC") {
			return nil, fail("INVALID_IDENTIFIER")
		}
		var env envelope
		found, err := load(s, key("document", args[0]), &env)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fail("DOCUMENT_NOT_FOUND")
		}
		return env, nil
	case "GetPurchase", "GetFreightCost", "GetRetailReport", "VerifyDocument":
		ps, ok := s.(privateLedger)
		if !ok || len(args) != 1 || len(request) != 1 {
			return nil, fail("INVALID_SCHEMA")
		}
		if fn == "VerifyDocument" {
			return verifyOpening(ps, msp, role, args[0], request[0].transient)
		}
		if len(request[0].transient) != 0 {
			return nil, fail("INVALID_SCHEMA")
		}
		return privateQuery(ps, msp, role, fn, args[0], nil)
	case "CreateBatch", "OfferPickup", "AcceptPickup", "RecordFreightCost", "OfferDelivery", "AcceptDelivery", "ReportRetailPrice":
		if len(args) != 1 {
			return nil, fail("INVALID_SCHEMA")
		}
		c, err := parseCommand(fn, args[0])
		if err != nil {
			return nil, err
		}
		if _, err := getConfig(s); err != nil {
			return nil, err
		}
		ps, ok := s.(privateLedger)
		if !ok || len(request) != 1 {
			return nil, fail("MISSING_EVIDENCE")
		}
		r := request[0]
		return (engine{ps, realEvidence{ps, r.transient, r.x}}).apply(c, r.x)
	default:
		return query(s, msp, fn, args)
	}
}

type fabricLedger struct{ stub shim.ChaincodeStubInterface }

func (f fabricLedger) readPrivate(c, k string) ([]byte, error) { return f.stub.GetPrivateData(c, k) }
func (f fabricLedger) writePrivate(c, k string, v []byte) error {
	return f.stub.PutPrivateData(c, k, v)
}

func (f fabricLedger) read(k string) ([]byte, error)  { return f.stub.GetState(k) }
func (f fabricLedger) write(k string, v []byte) error { return f.stub.PutState(k, v) }
func (f fabricLedger) remove(k string) error          { return f.stub.DelState(k) }
func (f fabricLedger) emit(n string, v []byte) error  { return f.stub.SetEvent(n, v) }
func (f fabricLedger) page(kind string, attrs []string, size int32, token string) ([]entry, string, error) {
	bookmark := ""
	prefix := kind + ":" + strings.Join(attrs, ":") + ":"
	if token != "" {
		b, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil || base64.RawURLEncoding.EncodeToString(b) != token || !strings.HasPrefix(string(b), prefix) {
			return nil, "", fail("INVALID_PAGE")
		}
		bookmark = string(b[len(prefix):])
		// LevelDB bookmarks are the next composite key. Do not let a forged
		// token choose a start key outside this exact owner/state/history range.
		if !strings.HasPrefix(bookmark, key(kind, attrs...)) {
			return nil, "", fail("INVALID_PAGE")
		}
	}
	it, meta, err := f.stub.GetStateByPartialCompositeKeyWithPagination(kind, attrs, size, bookmark)
	if err != nil {
		return nil, "", fail("INTERNAL_ERROR")
	}
	defer it.Close()
	items := []entry{}
	for it.HasNext() {
		kv, err := it.Next()
		if err != nil {
			return nil, "", fail("INTERNAL_ERROR")
		}
		items = append(items, entry{kv.Key, kv.Value})
	}
	next := ""
	if meta.Bookmark != "" {
		next = base64.RawURLEncoding.EncodeToString([]byte(prefix + meta.Bookmark))
	}
	return items, next, nil
}

// Keep the public contract surface explicit; no reflection-exposed engine methods.
var _ shim.Chaincode = (*Contract)(nil)
