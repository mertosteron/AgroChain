package agrochain

import (
	"encoding/json"
	"sort"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

type entry struct {
	Key   string
	Value []byte
}
type ledger interface {
	read(string) ([]byte, error)
	write(string, []byte) error
	remove(string) error
	page(string, []string, int32, string) ([]entry, string, error)
	emit(string, []byte) error
}

func key(kind string, attrs ...string) string {
	k, err := shim.CreateCompositeKey(kind, attrs)
	if err != nil {
		panic("invalid internal composite key")
	}
	return k
}
func load(s ledger, k string, dst any) (bool, error) {
	b, err := s.read(k)
	if err != nil {
		return false, fail("INTERNAL_ERROR")
	}
	if len(b) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return false, fail("INTERNAL_ERROR")
	}
	return true, nil
}
func absent(s ledger, k, code string) error {
	b, err := s.read(k)
	if err != nil {
		return fail("INTERNAL_ERROR")
	}
	if len(b) != 0 {
		return fail(code)
	}
	return nil
}

type writePlan map[string][]byte

func (p writePlan) put(k string, v any) error {
	b, err := canonical(v)
	if err != nil {
		return err
	}
	p[k] = b
	return nil
}

// Validate and serialize everything before touching the stub. Fabric validates
// and atomically commits the entire RW set; a failed flush must never be submitted.
func (p writePlan) flush(s ledger, event Event) error {
	b, err := canonical(event)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if p[k] == nil {
			err = s.remove(k)
		} else {
			err = s.write(k, p[k])
		}
		if err != nil {
			return fail("INTERNAL_ERROR")
		}
	}
	if err := s.emit(event.EventType, b); err != nil {
		return fail("INTERNAL_ERROR")
	}
	return nil
}
