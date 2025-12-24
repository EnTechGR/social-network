# golang-migrate Quick Reference Card

## 📁 File Structure
```
API/pkg/db/migrations/sqlite/
├── 000001_description.up.sql    ← Apply changes
├── 000001_description.down.sql  ← Revert changes
├── 000002_description.up.sql
└── 000002_description.down.sql
```

---

## 🎯 Common Operations

### Create New Migration
```bash
# Manually create files
touch API/pkg/db/migrations/sqlite/000XXX_description.up.sql
touch API/pkg/db/migrations/sqlite/000XXX_description.down.sql
```

### Apply Migrations (Automatic)
```bash
cd API
go run cmd/main.go  # Migrations run on startup
```

### Apply Migrations (Manual)
```bash
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        up
```

### Rollback Last Migration
```bash
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        down 1
```

### Check Current Version
```bash
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        version
```

### Force Version (After Manual Fix)
```bash
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        force 5
```

---

## 📝 SQL Templates

### Add Table
**up.sql:**
```sql
CREATE TABLE IF NOT EXISTS table_name (
    id TEXT PRIMARY KEY,
    column1 TEXT NOT NULL,
    column2 INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_table_column ON table_name(column1);
```

**down.sql:**
```sql
DROP INDEX IF EXISTS idx_table_column;
DROP TABLE IF EXISTS table_name;
```

### Add Column
**up.sql:**
```sql
ALTER TABLE table_name ADD COLUMN new_column TEXT DEFAULT 'default_value';
```

**down.sql (SQLite):**
```sql
-- SQLite requires table recreation to drop column
CREATE TABLE table_name_new AS 
SELECT column1, column2 FROM table_name;  -- exclude new_column

DROP TABLE table_name;
ALTER TABLE table_name_new RENAME TO table_name;
```

### Insert Data
**up.sql:**
```sql
INSERT OR IGNORE INTO table_name (id, name) VALUES 
    ('1', 'Value 1'),
    ('2', 'Value 2');
```

**down.sql:**
```sql
DELETE FROM table_name WHERE id IN ('1', '2');
```

### Modify Column (SQLite)
**up.sql:**
```sql
-- SQLite requires table recreation
CREATE TABLE table_new (
    id TEXT PRIMARY KEY,
    column1 TEXT NOT NULL CHECK (LENGTH(column1) >= 5)  -- new constraint
);

INSERT INTO table_new SELECT * FROM table_name;
DROP TABLE table_name;
ALTER TABLE table_new RENAME TO table_name;
```

---

## 🚨 Common Errors & Fixes

### "Dirty database version X"
```bash
# Check schema
sqlite3 database/forum.db ".schema"

# Fix manually, then force version
migrate -path migrations -database "sqlite3://db.db" force X
```

### "Table already exists"
```sql
-- Always use IF NOT EXISTS
CREATE TABLE IF NOT EXISTS table_name (...);
```

### "Foreign key constraint failed"
```sql
-- Delete in correct order: children first, parents last
DROP TABLE child_table;   -- references parent
DROP TABLE parent_table;  -- parent last
```

### Migration takes too long
```sql
-- Split into smaller migrations
-- Or batch data operations
UPDATE table SET x=y WHERE id IN (
    SELECT id FROM table WHERE x IS NULL LIMIT 1000
);
```

---

## ✅ Best Practices Checklist

- [ ] Both .up.sql and .down.sql files exist
- [ ] Use `IF EXISTS` / `IF NOT EXISTS`
- [ ] Test rollback locally before deploying
- [ ] Keep migrations small and focused
- [ ] Add comments for complex migrations
- [ ] Never modify existing migration files
- [ ] Test on copy of production data
- [ ] Migrations are in version control

---

## 🔍 Verification Commands

```bash
# Check applied migrations
sqlite3 database/forum.db "SELECT * FROM schema_migrations;"

# List all tables
sqlite3 database/forum.db ".tables"

# Check table schema
sqlite3 database/forum.db ".schema table_name"

# Check if column exists
sqlite3 database/forum.db "PRAGMA table_info(table_name);"

# Count rows
sqlite3 database/forum.db "SELECT COUNT(*) FROM table_name;"
```

---

## 📋 Migration Naming Convention

**Format:** `VVVVVV_description.direction.sql`

✅ GOOD:
- `000001_create_base.up.sql`
- `000002_add_followers.up.sql`
- `000003_add_groups.down.sql`

❌ BAD:
- `1_create.sql` (not enough digits)
- `create_users.up.sql` (no version number)
- `000001-create-base.up.sql` (use underscore not dash)

---

## 🎯 SQLite Limitations

| Operation | SQLite Support | Workaround |
|-----------|----------------|------------|
| ADD COLUMN | ✅ Yes | ALTER TABLE |
| DROP COLUMN | ⚠️ v3.35+ only | Recreate table |
| RENAME COLUMN | ⚠️ v3.25+ only | Recreate table |
| ADD CONSTRAINT | ❌ No | Recreate table |
| DROP CONSTRAINT | ❌ No | Recreate table |

**Template for table recreation:**
```sql
CREATE TABLE table_new (...);
INSERT INTO table_new SELECT ... FROM table_old;
DROP TABLE table_old;
ALTER TABLE table_new RENAME TO table_old;
-- Recreate indexes
```

---

## 🛠️ Database Utilities (in your code)

```go
// In models/database_model.go

// Create backup
backupPath, err := createBackup("./database/forum.db")

// Cleanup old backups (older than 30 days)
err := cleanupOldBackups(30)

// Restore from backup
err := RestoreFromBackup(backupPath)

// List available backups
backups, err := ListBackups()

// Cleanup expired OAuth states
err := CleanupExpiredOAuthStates(db)
```

---

## 📞 Help Resources

| Document | Purpose |
|----------|---------|
| COMPLETE_MIGRATION_GUIDE.md | Full migration documentation |
| REDUNDANCY_ANALYSIS.md | What was removed and why |
| CLEANUP_GUIDE.md | Step-by-step cleanup process |
| golang-migrate docs | https://github.com/golang-migrate/migrate |

---

## 🎓 Quick Start for New Developers

```bash
# 1. Clone repo
git clone <repo>
cd API

# 2. Install dependencies
go mod download

# 3. Run application (migrations apply automatically)
go run cmd/main.go

# 4. Verify database
sqlite3 database/forum.db "SELECT * FROM schema_migrations;"

# 5. View schema
sqlite3 database/forum.db ".schema user"
```

---

## 💡 Pro Tips

1. **Always test rollback:** Write .down.sql immediately after .up.sql
2. **Use transactions:** golang-migrate does this automatically per file
3. **Document why:** Add comments explaining complex migrations
4. **Version control:** Commit migration files before applying
5. **Backup first:** Always backup before applying in production
6. **Batch operations:** Split large data migrations into smaller chunks
7. **Monitor timing:** Profile migration execution time in production

---

## 🚀 Deployment Checklist

```bash
# Pre-deployment
- [ ] Migration files tested locally
- [ ] Rollback tested locally
- [ ] Backup verified
- [ ] Migration reviewed by team
- [ ] Performance tested on production-sized data

# Deployment
- [ ] Backup production database
- [ ] Apply migration
- [ ] Verify schema
- [ ] Run smoke tests
- [ ] Monitor application logs

# Post-deployment
- [ ] Verify data integrity
- [ ] Check application performance
- [ ] Document any issues
- [ ] Update team on completion
```

---

**Keep this card handy for daily development! 📌**