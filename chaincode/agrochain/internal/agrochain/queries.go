package agrochain

import "encoding/json"

// Set only at build time by the reproducible release script, never at endorsement.
var ReleaseVersion = "0.2.1"

type pageInput struct {
	SchemaVersion string `json:"schemaVersion"`
	Filter        string `json:"filter"`
	PageSize      int32  `json:"pageSize"`
	Bookmark      string `json:"bookmark"`
}
type Page struct {
	SchemaVersion string            `json:"schemaVersion"`
	Records       []json.RawMessage `json:"records"`
	Bookmark      string            `json:"bookmark"`
}

func query(s ledger, actor, fn string, args []string) (any, error) {
	if fn == "Health" {
		if len(args) != 0 {
			return nil, fail("INVALID_SCHEMA")
		}
		_, configErr := getConfig(s)
		if configErr != nil && configErr.(*DomainError).Code != "CONFIGURATION_REQUIRED" {
			return nil, configErr
		}
		return struct {
			SchemaVersion        string `json:"schemaVersion"`
			Version              string `json:"version"`
			WritesEnabled        bool   `json:"writesEnabled"`
			EvidenceVerification string `json:"evidenceVerification"`
		}{"agrochain.health.v1", ReleaseVersion, configErr == nil, "REQUIRED_STAGE_4"}, nil
	}
	if len(args) != 1 {
		return nil, fail("INVALID_SCHEMA")
	}
	switch fn {
	case "GetBatch", "BatchExists":
		if !validID(args[0], "BAT") {
			return nil, fail("INVALID_IDENTIFIER")
		}
		var b Batch
		ok, err := load(s, key("batch", args[0]), &b)
		if err != nil {
			return nil, err
		}
		if fn == "BatchExists" {
			return ok, nil
		}
		if !ok {
			return nil, fail("BATCH_NOT_FOUND")
		}
		return b, nil
	case "GetTransfer":
		if !validID(args[0], "TRF") {
			return nil, fail("INVALID_IDENTIFIER")
		}
		var t Transfer
		ok, err := load(s, key("transfer", args[0]), &t)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fail("TRANSFER_NOT_FOUND")
		}
		return t, nil
	case "GetOperation":
		if !validID(args[0], "OP") {
			return nil, fail("INVALID_IDENTIFIER")
		}
		var r Receipt
		ok, err := load(s, key("operation", actor, args[0]), &r)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fail("OPERATION_NOT_FOUND")
		}
		return r, nil
	case "GetBatchHistory", "QueryBatchesByOwner", "QueryBatchesByState":
		var p pageInput
		if err := decode([]byte(args[0]), &p, "schemaVersion", "filter", "pageSize", "bookmark"); err != nil {
			return nil, err
		}
		if p.SchemaVersion != "agrochain.page-request.v1" {
			return nil, fail("UNSUPPORTED_SCHEMA_VERSION")
		}
		if p.PageSize < 1 || p.PageSize > 100 {
			return nil, fail("INVALID_PAGE")
		}
		kind := ""
		switch fn {
		case "GetBatchHistory":
			if !validID(p.Filter, "BAT") {
				return nil, fail("INVALID_IDENTIFIER")
			}
			kind = "history"
		case "QueryBatchesByOwner":
			if !member(p.Filter) {
				return nil, fail("UNAUTHORIZED_ORGANIZATION")
			}
			kind = "owner"
		case "QueryBatchesByState":
			if !validState(p.Filter) {
				return nil, fail("INVALID_STATE_TRANSITION")
			}
			kind = "state"
		}
		// Bookmarks are opaque, scoped to this query by the adapter. They cannot
		// supply arbitrary world-state keys or an unbounded result size.
		items, next, err := s.page(kind, []string{p.Filter}, p.PageSize, p.Bookmark)
		if err != nil {
			return nil, err
		}
		out := Page{SchemaVersion: "agrochain.page.v1", Records: []json.RawMessage{}, Bookmark: next}
		for _, item := range items {
			var value any
			if kind == "history" {
				var r Receipt
				if json.Unmarshal(item.Value, &r) != nil {
					return nil, fail("INTERNAL_ERROR")
				}
				value = r
			} else {
				var id string
				if json.Unmarshal(item.Value, &id) != nil || !validID(id, "BAT") {
					return nil, fail("INTERNAL_ERROR")
				}
				var b Batch
				ok, err := load(s, key("batch", id), &b)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, fail("INTERNAL_ERROR")
				}
				value = b
			}
			b, err := canonical(value)
			if err != nil {
				return nil, err
			}
			out.Records = append(out.Records, b)
		}
		return out, nil
	default:
		return nil, fail("UNSUPPORTED_PILOT_OPERATION")
	}
}
