.PHONY: single single-clean three

CONFIG := config.yaml
MAIN := cmd/main.go

DATA_DIR := temp/data

SINGLE_DIR := single-run
THREE_DIR := three-run

DATA_SINGLE := $(DATA_DIR)/$(SINGLE_DIR)
DATA_THREE := $(DATA_DIR)/$(THREE_DIR)

THREE_TMUX_SESSION := three-tmux

help:
	@go run $(MAIN) -help

single:
	@mkdir -p $(DATA_SINGLE)
	@go run $(MAIN) -config $(CONFIG) -data $(DATA_SINGLE)

single-clean:
	@rm -rf $(DATA_SINGLE)

three-tmux:
	@mkdir -p $(DATA_THREE)
	@for i in 1 2 3 ; do mkdir -p $(DATA_THREE)/$$i ; done
	@tmux new-session -d -s $(THREE_TMUX_SESSION) \
		"bash -c 'go run $(MAIN) -config $(CONFIG) -data $(DATA_THREE)/1 ; bash'"
	@sleep 3
	@tmux split-window -v \
		"bash -c 'go run $(MAIN) -config $(CONFIG) -data $(DATA_THREE)/2 -internal-port 3001 -public-port 8091 -join-to localhost:8090 ; bash'"
	@tmux split-window -h \
		"bash -c 'go run $(MAIN) -config $(CONFIG) -data $(DATA_THREE)/3 -internal-port 3002 -public-port 8092 -join-to localhost:8090 ; bash'"
	@tmux select-layout tiled
	@tmux attach -t $(THREE_TMUX_SESSION)

three-restart-leader:
	@go run $(MAIN) -config $(CONFIG) -data $(DATA_THREE)/1

three-tmux-clean:
	@rm -rf $(DATA_THREE)
	@tmux kill-session -tf $(THREE_TMUX_SESSION)

three-docker:
	@docker compose up -d

three-docker-clean:
	@docker compose down --volumes --rmi all
