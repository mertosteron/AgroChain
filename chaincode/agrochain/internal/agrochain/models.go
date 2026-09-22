package agrochain

import "encoding/json"

const (
	Producer        = "ProducerMSP"
	Logistics       = "LogisticsMSP"
	Retailer        = "RetailerMSP"
	Regulator       = "RegulatorMSP"
	Created         = "CREATED"
	PickupPending   = "PICKUP_PENDING"
	InTransport     = "IN_TRANSPORT"
	DeliveryPending = "DELIVERY_PENDING"
	Received        = "RECEIVED"
	RetailReported  = "RETAIL_REPORTED"
)

type TxTime struct {
	Seconds string `json:"seconds"`
	Nanos   int32  `json:"nanos"`
}

type Command struct {
	SchemaVersion   string          `json:"schemaVersion"`
	OperationID     string          `json:"operationId"`
	BatchID         string          `json:"batchId"`
	ExpectedVersion int64           `json:"expectedVersion"`
	Command         string          `json:"command"`
	Payload         json.RawMessage `json:"payload"`
}

type Batch struct {
	SchemaVersion       string `json:"schemaVersion"`
	BatchID             string `json:"batchId"`
	ProductCode         string `json:"productCode"`
	GradeCode           string `json:"gradeCode"`
	QuantityGrams       int64  `json:"quantityGrams"`
	OriginRegionCode    string `json:"originRegionCode"`
	HarvestDate         string `json:"harvestDate"`
	ProducerMSP         string `json:"producerMsp"`
	LogisticsMSP        string `json:"logisticsMsp"`
	IntendedRetailerMSP string `json:"intendedRetailerMsp"`
	OwnerMSP            string `json:"ownerMsp"`
	CustodianMSP        string `json:"custodianMsp"`
	State               string `json:"state"`
	Version             int64  `json:"version"`
	RetailLotID         string `json:"retailLotId,omitempty"`
	CreatedTxID         string `json:"createdTxId"`
	UpdatedTxID         string `json:"updatedTxId"`
	TxTime              TxTime `json:"txTime"`
}

type EvidenceRefs struct {
	SchemaVersion       string `json:"schemaVersion"`
	BatchID             string `json:"batchId"`
	CKSDocumentID       string `json:"cksDocumentId,omitempty"`
	PurchaseDocumentID  string `json:"purchaseDocumentId,omitempty"`
	HKSDocumentID       string `json:"hksDocumentId,omitempty"`
	TransportDocumentID string `json:"transportDocumentId,omitempty"`
	FreightCostID       string `json:"freightCostId,omitempty"`
	RetailReportID      string `json:"retailReportId,omitempty"`
}

type Transfer struct {
	SchemaVersion     string `json:"schemaVersion"`
	TransferID        string `json:"transferId"`
	BatchID           string `json:"batchId"`
	Kind              string `json:"kind"`
	Status            string `json:"status"`
	FromMSP           string `json:"fromMsp"`
	ToMSP             string `json:"toMsp"`
	QuantityGrams     int64  `json:"quantityGrams"`
	OfferOperationID  string `json:"offerOperationId"`
	AcceptOperationID string `json:"acceptOperationId,omitempty"`
	OfferedTxID       string `json:"offeredTxId"`
	AcceptedTxID      string `json:"acceptedTxId,omitempty"`
}

type Receipt struct {
	SchemaVersion string   `json:"schemaVersion"`
	OperationID   string   `json:"operationId"`
	BatchID       string   `json:"batchId"`
	Command       string   `json:"command"`
	ActorMSP      string   `json:"actorMsp"`
	TxID          string   `json:"txId"`
	TxTime        TxTime   `json:"txTime"`
	ResultVersion int64    `json:"resultVersion"`
	ObjectIDs     []string `json:"objectIds"`
}

type Event struct {
	SchemaVersion string   `json:"schemaVersion"`
	EventType     string   `json:"eventType"`
	OperationID   string   `json:"operationId"`
	BatchID       string   `json:"batchId"`
	ActorMSP      string   `json:"actorMsp"`
	TxID          string   `json:"txId"`
	TxTime        TxTime   `json:"txTime"`
	ResultVersion int64    `json:"resultVersion"`
	ObjectIDs     []string `json:"objectIds"`
	PreviousState string   `json:"previousState"`
	NewState      string   `json:"newState"`
}

type Lot struct {
	SchemaVersion string `json:"schemaVersion"`
	RetailLotID   string `json:"retailLotId"`
	BatchID       string `json:"batchId"`
	QuantityGrams int64  `json:"quantityGrams"`
	RetailerMSP   string `json:"retailerMsp"`
	ReportID      string `json:"reportId"`
	TxID          string `json:"txId"`
}

type createInput struct {
	ProductCode         string `json:"productCode"`
	GradeCode           string `json:"gradeCode"`
	QuantityGrams       int64  `json:"quantityGrams"`
	OriginRegionCode    string `json:"originRegionCode"`
	HarvestDate         string `json:"harvestDate"`
	LogisticsMSP        string `json:"logisticsMsp"`
	IntendedRetailerMSP string `json:"intendedRetailerMsp"`
}
type handoffInput struct {
	TransferID    string `json:"transferId"`
	QuantityGrams int64  `json:"quantityGrams"`
}
type freightInput struct {
	CostID string `json:"costId"`
}
type reportInput struct {
	ReportID    string `json:"reportId"`
	RetailLotID string `json:"retailLotId"`
	PolicyID    string `json:"policyId"`
}

// Verified evidence is an internal Stage 4 seam, never a deserializable public
// argument. Only _test.go provides a successful provider in Stage 3.
type verified struct {
	Public                                                            writePlan
	Private                                                           []privateWrite
	CKS, Purchase, HKS, Manifest                                      string
	Quantity, PurchasePrice, PurchaseTotal, FreightTotal, RetailPrice int64
	Currency, TaxBasis, PolicyID                                      string
}
type evidenceProvider interface {
	prepare(Command, Batch, EvidenceRefs) (verified, error)
}
type closedEvidence struct{}

func (closedEvidence) prepare(Command, Batch, EvidenceRefs) (verified, error) {
	return verified{}, fail("EVIDENCE_VERIFICATION_UNAVAILABLE")
}
