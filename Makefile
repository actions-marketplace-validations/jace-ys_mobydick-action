include .env

.PHONY: bin/action

bin/action:
	cd bin && go run main.go distribute \
		--organisation ${ORGANISATION} \
		--token ${TOKEN} \
		--private
