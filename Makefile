.PHONY: help tag-patch tag-minor tag-major

help:
	@echo "Available targets:"
	@echo "  tag-patch   - Increment patch version (v0.0.X -> v0.0.X+1)"
	@echo "  tag-minor   - Increment minor version (v0.X.0 -> v0.X+1.0)"
	@echo "  tag-major   - Increment major version (vX.0.0 -> vX+1.0.0)"


# Default to v0.0.0 if no tags exist
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)

# Parse current version
MAJOR := $(shell echo $(CURRENT_TAG) | awk -F. '{print $$1}' | sed 's/v//')
MINOR := $(shell echo $(CURRENT_TAG) | awk -F. '{print $$2}')
PATCH := $(shell echo $(CURRENT_TAG) | awk -F. '{print $$3}')

tag-patch:
	@echo "Current version: $(CURRENT_TAG)"
	@new_patch=$$(($(PATCH) + 1)); \
	NEW_TAG="v$(MAJOR).$(MINOR).$$new_patch"; \
	echo "New version: $$NEW_TAG"; \
	git tag $$NEW_TAG; \
	echo "Created tag $$NEW_TAG"

tag-minor:
	@echo "Current version: $(CURRENT_TAG)"
	@new_minor=$$(($(MINOR) + 1)); \
	NEW_TAG="v$(MAJOR).$$new_minor.0"; \
	echo "New version: $$NEW_TAG"; \
	git tag $$NEW_TAG; \
	echo "Created tag $$NEW_TAG"

tag-major:
	@echo "Current version: $(CURRENT_TAG)"
	@new_major=$$(($(MAJOR) + 1)); \
	NEW_TAG="v$$new_major.0.0"; \
	echo "New version: $$NEW_TAG"; \
	git tag $$NEW_TAG; \
	echo "Created tag $$NEW_TAG"
