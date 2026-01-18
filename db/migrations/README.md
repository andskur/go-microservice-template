# Database Migrations

This directory contains SQL migration files managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Structure

- Migrations are numbered sequentially: `000001`, `000002`, etc.
- Each migration has two files:
  - `*.up.sql` - Applied when migrating forward
  - `*.down.sql` - Applied when rolling back

## Usage

### Create new migration
```bash
make migrate-create NAME=add_user_roles
```

### Apply migrations
```bash
make migrate-up
```

### Rollback last migration
```bash
make migrate-down
```

### Check current version
```bash
make migrate-version
```

### Force version (recovery from dirty state)
```bash
make migrate-force VERSION=2
```

### Drop all tables (⚠️ DANGER)
```bash
make migrate-drop
```

## Guidelines

- **Always test down migrations** - ensure they cleanly reverse up migrations
- **Keep migrations small** - one logical change per migration
- **Don't modify applied migrations** - create new migrations to fix issues
- **Use IF EXISTS/IF NOT EXISTS** - makes migrations idempotent and safe to re-run
- **Order matters** - migrations run in numerical order
- **Test locally first** - verify migrations work before committing

## Schema Overview

- `000001_initial_migration` - Creates `update_updated_at_column()` trigger function
- `000002_users_table` - Creates users table with status enum, indexes, and update trigger

## Migration Naming Convention

Use descriptive names that indicate the change:
- ✅ Good: `add_user_roles`, `create_sessions_table`, `add_email_verification`
- ❌ Bad: `update`, `fix`, `new_stuff`
