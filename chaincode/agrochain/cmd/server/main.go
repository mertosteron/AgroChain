// Only transport startup reads deployment configuration/files. Endorsement logic
// in internal/agrochain does not read files, environment, clocks or networks.
package main

import (
	"log"
	"os"

	"agrochain/chaincode/internal/agrochain"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

func main() {
	cert, err := os.ReadFile("/tls/server.crt")
	if err != nil {
		log.Fatal("chaincode TLS certificate unavailable")
	}
	key, err := os.ReadFile("/tls/server.key")
	if err != nil {
		log.Fatal("chaincode TLS key unavailable")
	}
	ccid := os.Getenv("CHAINCODE_ID")
	if ccid == "" {
		log.Fatal("chaincode package ID required")
	}
	server := shim.ChaincodeServer{CCID: ccid, Address: "0.0.0.0:9999", CC: &agrochain.Contract{}, TLSProps: shim.TLSProperties{Disabled: false, Key: key, Cert: cert}}
	if err := server.Start(); err != nil {
		log.Fatal("chaincode server stopped")
	}
}
