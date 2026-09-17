.PHONY: check generate network-up channel-create verify network-down clean-generated bootstrap smoke test
.NOTPARALLEL:

ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

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
