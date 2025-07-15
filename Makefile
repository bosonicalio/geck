# Makefile for Generating Go Mocks
#
# This Makefile provides a target to automate the creation of mocks using `mockgen`.
# It requires the `FILEPATH` of the source file to be passed as an argument.
#
# Usage:
#   make gen-mock FILEPATH=path/to/your/file.go

# =============================================================================
#  HELPER FUNCTIONS
# =============================================================================

# Function to add a suffix to each directory in a path.
define add-suffix
    $(if $(subst  ,/,$(strip $(foreach dir,$(subst /, ,$(patsubst %/,%,$(dir $(1)))),$(dir)$(2)))),$(subst  ,/,$(strip $(foreach dir,$(subst /, ,$(patsubst %/,%,$(dir $(1)))),$(dir)$(2))))/,)$(notdir $(1))
endef

# Function to get the name of the immediate parent folder.
define get-closest-folder
    $(if $(filter-out ./ .,$(patsubst %/,%,$(dir $(1)))),$(notdir $(patsubst %/,%,$(dir $(1)))))
endef

# =============================================================================
#  TARGETS
# =============================================================================

# Target to generate a mock for the given FILEPATH.
gen-mock:
	$(eval DEST_PATH := $(strip $(subst //,/,$(call add-suffix, ${FILEPATH},mock))))
	$(eval PACKAGE_NAME := $(strip $(call get-closest-folder, ${FILEPATH}))mock)
	@echo "➡️  Generating mock for [$(FILEPATH)]"
	mockgen -source=${FILEPATH} \
		-destination=$(DEST_PATH) \
		-package=$(PACKAGE_NAME)
	@echo "✅  Successfully created mock at $(DEST_PATH)"

run-all-tests:
	go test ./... -tags=integration,e2e -v -cover -race -p=10
