# Database Migration

This repository now separates database assets into two layers:

- `manifest/sql/mysql/`: versioned schema migrations managed by `golang-migrate`
- `manifest/sql/seed/`: one-off bootstrap seed data for new environments

Legacy snapshot files are kept here:

- `docker/mysql/schema.sql`
- `docker/mysql/seed_data.sql`
- `docker/mysql/init.sql`

Use them as export snapshots, not as the daily source of truth for incremental upgrades.

## Commands

Run from `admin-go/`:

```bash
make db-version
make db-up
make db-down STEPS=1
make db-create NAME=add_some_table
make db-seed
make db-bootstrap
```

## Install migrate

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Rules

1. Every schema change must add an `up` and `down` SQL file under `manifest/sql/mysql/`.
2. Do not edit applied historical migration files.
3. Seed data is only for bootstrap defaults; do not mix operational data into migrations.
4. `init.sql` is a full snapshot for fresh import and audit only, not the primary upgrade path.
