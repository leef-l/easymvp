.PHONY: db-up
db-up:
	@./hack/db-migrate.sh up

.PHONY: db-down
db-down:
	@./hack/db-migrate.sh down $(if $(STEPS),$(STEPS),1)

.PHONY: db-version
db-version:
	@./hack/db-migrate.sh version

.PHONY: db-status
db-status: db-version

.PHONY: db-create
db-create:
	@test -n "$(NAME)" || (echo "usage: make db-create NAME=add_xxx" && exit 1)
	@./hack/db-migrate.sh create "$(NAME)"

.PHONY: db-goto
db-goto:
	@test -n "$(VERSION)" || (echo "usage: make db-goto VERSION=1" && exit 1)
	@./hack/db-migrate.sh goto "$(VERSION)"

.PHONY: db-force
db-force:
	@test -n "$(VERSION)" || (echo "usage: make db-force VERSION=1" && exit 1)
	@./hack/db-migrate.sh force "$(VERSION)"

.PHONY: db-seed
db-seed:
	@GF_GCFG_PATH=app/mvp/manifest/config GF_GCFG_FILE=config.yaml \
		go run ./app/mvp/tools/dbctl seed -file manifest/sql/seed/mysql_seed.sql $(if $(FORCE),-force,)

.PHONY: db-bootstrap
db-bootstrap:
	@$(MAKE) db-up
	@$(MAKE) db-seed
