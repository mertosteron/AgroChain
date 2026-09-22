.PHONY: check generate network-up channel-create verify network-down clean-generated bootstrap smoke test
.NOTPARALLEL:

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
JAVA ?= java

check:
	bash "$(ROOT)network/scripts/prerequisites.sh"
generate:
	bash "$(ROOT)network/scripts/generate.sh"
network-up:
	bash "$(ROOT)network/scripts/up.sh"
channel-create:
	bash "$(ROOT)network/scripts/create-channel.sh"
verify:
	bash "$(ROOT)network/scripts/verify.sh"
network-down:
	bash "$(ROOT)network/scripts/down.sh"
clean-generated:
	bash "$(ROOT)network/scripts/clean-generated.sh"
bootstrap: check generate network-up channel-create
smoke:
	bash "$(ROOT)network/scripts/smoke.sh"
test:
	python3 -m unittest discover -s "$(ROOT)network/tests" -v

.PHONY: chaincode-prerequisites chaincode-deps chaincode-format chaincode-test chaincode-build chaincode-package chaincode-install chaincode-approve chaincode-commit chaincode-check chaincode-deploy chaincode-upgrade chaincode-integration
export CHAINCODE_VERSION CHAINCODE_SEQUENCE
chaincode-prerequisites:
	bash "$(ROOT)network/scripts/chaincode-toolchain.sh" install
chaincode-deps:
	bash "$(ROOT)network/scripts/chaincode-build.sh" deps
chaincode-format:
	bash "$(ROOT)network/scripts/chaincode-build.sh" format
chaincode-test chaincode-build:
	bash "$(ROOT)network/scripts/chaincode-build.sh" test
chaincode-package chaincode-install chaincode-approve chaincode-commit chaincode-check chaincode-deploy chaincode-upgrade:
	bash "$(ROOT)network/scripts/chaincode.sh" $(patsubst chaincode-%,%,$@)
chaincode-integration:
	python3 "$(ROOT)chaincode/agrochain/test/integration/fabric_test.py"
.PHONY: chaincode-restart-check
chaincode-restart-check:
	python3 "$(ROOT)chaincode/agrochain/test/integration/restart_test.py"

# Verify the approved Stage 3 boundary on an already deployed network.
# This preserves identities/volumes; it commits health probes and restarts nodes.
.PHONY: stage3-check
stage3-check: test chaincode-test verify chaincode-check chaincode-integration chaincode-restart-check

.PHONY: privacy-prepare privacy-integration privacy-vectors stage4-check
privacy-prepare:
	python3 "$(ROOT)network/scripts/privacy-prepare.py"
privacy-integration:
	python3 "$(ROOT)chaincode/agrochain/test/integration/privacy_test.py"
privacy-vectors:
	$(JAVA) "$(ROOT)chaincode/agrochain/test/VerifyVector.java"
stage4-check: test chaincode-test privacy-vectors verify chaincode-check chaincode-integration privacy-integration

.PHONY: backend-prerequisites backend-prepare backend-test backend-build backend-run backend-integration backend-gateway-test stage5-check
backend-prerequisites:
	python3 "$(ROOT)backend/scripts/toolchain.py"
backend-prepare: privacy-prepare
	python3 "$(ROOT)backend/scripts/prepare.py"
backend-test:
	bash "$(ROOT)backend/mvnw" test
backend-build:
	bash "$(ROOT)backend/mvnw" package
backend-run:
	AGROCHAIN_ROOT="$(ROOT)" "$(ROOT)network/tools/jdk-21.0.12.1+1/bin/java" -jar "$(ROOT)backend/target/agrochain-backend-0.3.0.jar"
backend-integration:
	python3 "$(ROOT)backend/test/integration_test.py"
backend-gateway-test:
	AGROCHAIN_ROOT="$(ROOT)" AGROCHAIN_LIVE_TESTS=true bash "$(ROOT)backend/mvnw" -Dtest=GatewayRecoveryTest test
stage5-check: backend-build chaincode-check backend-integration backend-gateway-test
