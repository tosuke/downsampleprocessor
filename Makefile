SRC_DIR := $(shell git rev-parse --show-toplevel)

GOCMD ?= go

TOOLS_MOD_DIR := $(SRC_DIR)/internal/tools
TOOLS_BIN_DIR := $(SRC_DIR)/.tools
TOOLS_PKG_NAMES := $(shell grep -E '\s+_\s+".+"' < $(TOOLS_MOD_DIR)/tools.go | tr -d ' \t_"')
TOOLS_BIN_NAMES := $(addprefix $(TOOLS_BIN_DIR)/, $(notdir $(TOOLS_PKG_NAMES)))

.PHONY: generate
generate:
	PATH=$(TOOLS_BIN_DIR):$$PATH $(GOCMD) generate ./...

.PHONY: tools
tools: $(TOOLS_BIN_NAMES)

$(TOOLS_BIN_DIR):
	mkdir -p $@
$(TOOLS_BIN_NAMES): $(TOOLS_BIN_DIR) $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum
	cd $(TOOLS_MOD_DIR) && $(GOCMD) build -o $@ -trimpath $(filter %/$(notdir $@),$(TOOLS_PKG_NAMES))
