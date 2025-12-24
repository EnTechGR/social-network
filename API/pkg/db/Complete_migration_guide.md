# Complete Migration Guide: golang-migrate for Social Network

## Table of Contents
1. [Overview](#overview)
2. [Migration System Architecture](#migration-system-architecture)
3. [File Structure](#file-structure)
4. [Migration Basics](#migration-basics)
5. [Creating New Migrations](#creating-new-migrations)
6. [Migration Operations](#migration-operations)
7. [Production Best Practices](#production-best-practices)
8. [Rollback Strategies](#rollback-strategies)
9. [Testing Migrations](#testing-migrations)
10. [Common Patterns](#common-patterns)
11. [Troubleshooting](#troubleshooting)

---

## Overview

### What is golang-migrate?

`golang-migrate` is a database migration tool that reads migration files from various sources (file system, S3, GitHub, etc.) and applies them in the correct order to your database.

### Why Use golang-migrate?

✅ **Version Control**: Each migration is numbered and tracked
✅ **Idempotent**: Migrations run exactly once
✅ **Reversible**: Every migration has an "up" and "down"
✅ **Database Agnostic**: Works with SQLite, PostgreSQL, MySQL, etc.
✅ **Production Ready**: Battle-tested in thousands of applications

---

## Migration System Architecture

### How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Startup                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  InitDB() in database_model.go              │
│                                                               │
│  1. Opens database connection                                │
│  2. Calls sqlite.Migrate(db)                                 │
│  3. Seeds categories                                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│             sqlite.Migrate() in pkg/db/sqlite/sqlite.go     │
│                                                               │
│  1. Creates sqlite3 driver instance                          │
│  2. Resolves migrations directory path                       │
│  3. Creates migrate instance                                 │
│  4. Runs m.Up() to apply pending migrations                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     golang-migrate Library                   │
│                                                               │
│  1. Checks schema_migrations table                           │
│  2. Identifies pending migrations                            │
│  3. Applies migrations in order                              │
│  4. Records applied versions                                 │
└─────────────────────────────────────────────────────────────┘
```

### schema_migrations Table

golang-migrate automatically creates a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);
```

- **version**: Migration version number (e.g., 1, 2, 3)
- **dirty**: Indicates if migration failed mid-execution

---

## File Structure

### Required Directory Layout

```
API/
├── pkg/
│   └── db/
│       ├── migrations/
│       │   └── sqlite/
│       │       ├── 000001_create_base.up.sql
│       │       ├── 000001_create_base.down.sql
│       │       ├── 000002_add_groups.up.sql
│       │       ├── 000002_add_groups.down.sql
│       │       ├── 000003_add_chat.up.sql
│       │       └── 000003_add_chat.down.sql
│       └── sqlite/
│           └── sqlite.go  (migration runner)
└── models/
    └── database_model.go  (InitDB, utilities)
```

### Naming Convention

**Format:** `{version}_{description}.{direction}.sql`

- **version**: 6-digit number (000001, 000002, etc.)
- **description**: Snake_case description (create_users, add_groups)
- **direction**: `up` or `down`

**Examples:**
```
000001_create_base.up.sql       # Initial schema creation
000001_create_base.down.sql     # Rollback initial schema
000002_add_follow_system.up.sql # Add followers feature
000002_add_follow_system.down.sql # Remove followers feature
```

---

## Migration Basics

### The Migration Pair Concept

Every migration consists of TWO files:

1. **UP migration** (`.up.sql`): Applies changes forward
2. **DOWN migration** (`.down.sql`): Reverts changes backward

**Example:**

`000002_add_status_column.up.sql`:
```sql
ALTER TABLE posts ADD COLUMN status TEXT DEFAULT 'published';
```

`000002_add_status_column.down.sql`:
```sql
ALTER TABLE posts DROP COLUMN status;
```

### Atomic Transactions

Each migration file is wrapped in a transaction automatically:
- If ANY statement fails, the ENTIRE migration is rolled back
- Database remains in consistent state
- Migration marked as "dirty" if partial failure occurs

---

## Creating New Migrations

### Step-by-Step: Adding a New Table

Let's add a "likes" table to track post likes.

#### Step 1: Create Migration Files

Create two new files in `API/pkg/db/migrations/sqlite/`:

**File: `000002_add_likes_table.up.sql`**
```sql
-- Create likes table to track user likes on posts
CREATE TABLE IF NOT EXISTS likes (
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);

-- Add index for efficient queries by post
CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id);

-- Add index for efficient queries by user
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
```

**File: `000002_add_likes_table.down.sql`**
```sql
-- Remove indexes first (reverse order)
DROP INDEX IF EXISTS idx_likes_user_id;
DROP INDEX IF EXISTS idx_likes_post_id;

-- Remove table
DROP TABLE IF EXISTS likes;
```

#### Step 2: Test the Migration

```bash
# Start your application
cd API
go run cmd/main.go
```

Output should show:
```
Database migrations applied via golang-migrate.
Categories populated (duplicates ignored).
```

#### Step 3: Verify in Database

```bash
sqlite3 API/database/forum.db

-- Check if table was created
.schema likes

-- Check migration version
SELECT * FROM schema_migrations;
```

Expected output:
```
version | dirty
--------|------
1       | false
2       | false
```

---

### Step-by-Step: Modifying an Existing Table

Let's add a "pinned" column to posts.

#### Step 1: Create Migration Files

**File: `000003_add_pinned_to_posts.up.sql`**
```sql
-- Add pinned column to posts table
ALTER TABLE posts ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT 0;

-- Add index for pinned posts queries
CREATE INDEX IF NOT EXISTS idx_posts_pinned ON posts(pinned, created_at DESC);

-- Optional: Set some existing posts as pinned (data migration)
-- UPDATE posts SET pinned = 1 WHERE post_id IN ('abc', 'def');
```

**File: `000003_add_pinned_to_posts.down.sql`**
```sql
-- Remove index
DROP INDEX IF EXISTS idx_posts_pinned;

-- SQLite doesn't support DROP COLUMN before version 3.35.0
-- For older SQLite versions, you need to recreate the table:

-- 1. Create new table without pinned column
CREATE TABLE posts_backup (
    post_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT,
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public','followers','private')),
    title TEXT CHECK (LENGTH(title) <= 200),
    content TEXT CHECK (LENGTH(content) <= 2000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);

-- 2. Copy data (excluding pinned column)
INSERT INTO posts_backup 
SELECT post_id, user_id, group_id, visibility, title, content, 
       created_at, updated_at, deleted_at 
FROM posts;

-- 3. Drop old table
DROP TABLE posts;

-- 4. Rename backup to original name
ALTER TABLE posts_backup RENAME TO posts;

-- 5. Recreate indexes that were on posts table
CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_posts_group ON posts(group_id);
```

---

### Step-by-Step: Inserting Data

Let's add default categories via migration.

**⚠️ Important Note:** For this project, categories are seeded by application code (`populateCategories()` in `database_model.go`). However, for demonstration, here's how you would do it via migration:

**File: `000004_seed_categories.up.sql`**
```sql
-- Insert default categories
-- Using INSERT OR IGNORE to make it idempotent
INSERT OR IGNORE INTO categories (category_id, name) VALUES 
    (1, 'Drama'),
    (2, 'Fantasy & Sci-Fi'),
    (3, 'Mystery & Thriller'),
    (4, 'Romance'),
    (5, 'Horror'),
    (6, 'Non-Fiction'),
    (7, 'Young Adult & Kids');
```

**File: `000004_seed_categories.down.sql`**
```sql
-- Remove seeded categories
DELETE FROM categories WHERE category_id IN (1, 2, 3, 4, 5, 6, 7);

-- Reset auto-increment counter
DELETE FROM sqlite_sequence WHERE name = 'categories';
```

---

### Step-by-Step: Deleting a Table

Let's say we want to remove an obsolete table.

**File: `000005_remove_old_logs.up.sql`**
```sql
-- Drop indexes first
DROP INDEX IF EXISTS idx_logs_timestamp;
DROP INDEX IF EXISTS idx_logs_user;

-- Drop table
DROP TABLE IF EXISTS old_logs;
```

**File: `000005_remove_old_logs.down.sql`**
```sql
-- Recreate the table
CREATE TABLE IF NOT EXISTS old_logs (
    log_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON old_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_user ON old_logs(user_id);

-- Note: Historical data cannot be restored
```

---

## Migration Operations

### Running Migrations (Automatic)

Migrations run automatically when your application starts:

```go
// In database_model.go
func InitDB() (*sql.DB, error) {
    // ... database connection code ...
    
    // Apply migrations
    if err := dbmigrate.Migrate(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to apply migrations: %v", err)
    }
    
    // ... seed categories ...
}
```

### Running Migrations (Manual via CLI)

For development/testing, you can use the migrate CLI:

```bash
# Install migrate CLI
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Navigate to your project
cd API

# Run migrations up
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        up

# Run migrations down (rollback all)
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        down

# Rollback specific number of migrations
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        down 2  # Rollback last 2 migrations

# Go to specific version
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        goto 3

# Check current version
migrate -path pkg/db/migrations/sqlite \
        -database "sqlite3://database/forum.db?_foreign_keys=on" \
        version
```

---

## Production Best Practices

### 1. Always Create Down Migrations

Even if you think you'll never roll back:
```sql
-- ❌ BAD: No down migration
-- 000006_add_feature.up.sql exists
-- 000006_add_feature.down.sql MISSING

-- ✅ GOOD: Both migrations exist
-- 000006_add_feature.up.sql
-- 000006_add_feature.down.sql
```

### 2. Keep Migrations Small and Focused

```sql
-- ❌ BAD: One giant migration doing everything
-- 000002_update_entire_schema.up.sql (adds 10 tables, modifies 5 tables, adds 20 indexes)

-- ✅ GOOD: Separate migrations for each logical change
-- 000002_add_likes_table.up.sql
-- 000003_add_shares_table.up.sql
-- 000004_modify_posts_add_pinned.up.sql
```

### 3. Test Migrations Before Production

```bash
# 1. Backup production database
cp production.db production_backup.db

# 2. Apply migration to test database
migrate -path migrations -database "sqlite3://test.db" up

# 3. Run application tests
go test ./...

# 4. Manually verify critical queries still work

# 5. Test rollback
migrate -path migrations -database "sqlite3://test.db" down 1

# 6. Only then deploy to production
```

### 4. Use Transactions (Implicit)

golang-migrate wraps each migration in a transaction automatically, but be aware:

```sql
-- ✅ This entire file is one transaction
CREATE TABLE new_table (...);
CREATE INDEX idx_new_table ON new_table(...);
INSERT INTO new_table VALUES (...);

-- If ANY statement fails, ALL statements are rolled back
```

### 5. Handle Data Migrations Carefully

When moving data between tables:

```sql
-- File: 000007_restructure_notifications.up.sql

-- 1. Create new table
CREATE TABLE notifications_new (
    notification_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    data TEXT NOT NULL,  -- JSON blob
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Migrate data with transformation
INSERT INTO notifications_new (notification_id, user_id, type, data, created_at)
SELECT 
    notification_id,
    user_id,
    type,
    json_object(
        'post_id', post_id,
        'comment_id', comment_id,
        'from_user_id', from_user_id
    ) as data,
    created_at
FROM notifications;

-- 3. Drop old table
DROP TABLE notifications;

-- 4. Rename new table
ALTER TABLE notifications_new RENAME TO notifications;

-- 5. Recreate indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
```

### 6. Add Columns with Defaults

When adding non-nullable columns:

```sql
-- ✅ GOOD: Provides default value
ALTER TABLE posts ADD COLUMN status TEXT NOT NULL DEFAULT 'published';

-- ❌ BAD: No default for non-nullable column
-- ALTER TABLE posts ADD COLUMN status TEXT NOT NULL;
-- This fails if table already has data
```

### 7. Use IF EXISTS / IF NOT EXISTS

Makes migrations idempotent:

```sql
-- ✅ GOOD: Safe to run multiple times
CREATE TABLE IF NOT EXISTS new_table (...);
CREATE INDEX IF NOT EXISTS idx_name ON table(column);
DROP TABLE IF EXISTS old_table;
DROP INDEX IF EXISTS idx_old;

-- ❌ BAD: Fails if run twice
-- CREATE TABLE new_table (...);
-- CREATE INDEX idx_name ON table(column);
```

### 8. Document Complex Migrations

```sql
-- File: 000008_optimize_message_queries.up.sql

/*
 * Migration: Optimize message queries
 * 
 * Problem: Message list queries are slow with 10,000+ messages
 * 
 * Solution: Add composite index on (receiver_id, created_at DESC)
 * 
 * Performance impact:
 *   - Query time: 450ms → 12ms (97% improvement)
 *   - Index size: ~2MB for 10,000 messages
 * 
 * Rollback impact: Query performance will degrade
 */

CREATE INDEX IF NOT EXISTS idx_messages_receiver_created 
ON messages(receiver_id, created_at DESC);
```

---

## Rollback Strategies

### Automatic Rollback (Transaction Failure)

If any statement in a migration fails, the entire migration is automatically rolled back:

```sql
-- File: 000009_bad_migration.up.sql
CREATE TABLE test_table (id INTEGER);
CREATE INDEX idx_test ON test_table(nonexistent_column);  -- FAILS
-- Transaction is automatically rolled back, test_table is NOT created
```

### Manual Rollback (Down Migration)

To manually roll back the last migration:

```bash
# Using migrate CLI
migrate -path migrations -database "sqlite3://forum.db" down 1

# Or programmatically in Go
m, err := migrate.NewWithDatabaseInstance(...)
if err := m.Steps(-1); err != nil {  // Roll back 1 migration
    log.Fatal(err)
}
```

### Rollback Considerations

**What can be easily rolled back:**
- ✅ Adding tables (just drop them)
- ✅ Adding columns (drop the column or recreate table)
- ✅ Adding indexes (just drop them)
- ✅ Inserting data (just delete it)

**What's harder to roll back:**
- ⚠️ Dropping tables (data is lost)
- ⚠️ Dropping columns (data is lost)
- ⚠️ Data transformations (original data may be lost)
- ⚠️ Schema changes that break old code

**Production Rollback Strategy:**

1. **Blue-Green Deployment:**
   - Keep old version running
   - Deploy new version with migrations to separate environment
   - Switch traffic after verification
   - Can switch back if issues arise

2. **Feature Flags:**
   ```go
   if featureFlags["new_likes_system"] {
       // Use new likes table
   } else {
       // Use old reactions system
   }
   ```

3. **Backward-Compatible Migrations:**
   - Add new column with NULL allowed
   - Gradually migrate data
   - Make column NOT NULL in later migration
   - Drop old column in even later migration

---

## Testing Migrations

### Unit Test for Migration Files

```go
// File: pkg/db/sqlite/sqlite_test.go

package sqlite

import (
    "database/sql"
    "testing"
    _ "github.com/mattn/go-sqlite3"
)

func TestMigrations(t *testing.T) {
    // Create in-memory database
    db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Run migrations
    if err := Migrate(db); err != nil {
        t.Fatalf("Migration failed: %v", err)
    }

    // Verify tables were created
    tables := []string{"user", "posts", "comments", "categories"}
    for _, table := range tables {
        var count int
        query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
        if err := db.QueryRow(query, table).Scan(&count); err != nil {
            t.Fatal(err)
        }
        if count != 1 {
            t.Errorf("Table %s was not created", table)
        }
    }

    // Verify schema_migrations table exists
    var version int
    err = db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
    if err != nil {
        t.Fatalf("Failed to query schema_migrations: %v", err)
    }
    t.Logf("Current migration version: %d", version)
}
```

### Integration Test

```go
// File: models/database_test.go

package models

import (
    "testing"
    "os"
)

func TestDatabaseInitialization(t *testing.T) {
    // Use test database
    os.Setenv("DB_PATH", "test_forum.db")
    defer os.Remove("test_forum.db")

    // Initialize database
    db, err := InitDB()
    if err != nil {
        t.Fatalf("InitDB failed: %v", err)
    }
    defer db.Close()

    // Test that categories were seeded
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
    if err != nil {
        t.Fatal(err)
    }
    if count == 0 {
        t.Error("No categories were seeded")
    }

    // Test that schema is correct
    var tableName string
    err = db.QueryRow(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name='user'
    `).Scan(&tableName)
    if err != nil {
        t.Fatal("User table not found")
    }
}
```

---

## Common Patterns

### Pattern 1: Adding a Feature with Multiple Tables

When adding a complex feature like "groups":

**File: `000010_add_groups.up.sql`**
```sql
-- 1. Create main table
CREATE TABLE IF NOT EXISTS groups (
    group_id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Create related tables
CREATE TABLE IF NOT EXISTS group_members (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner','admin','member')),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_posts (
    post_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 3. Create indexes
CREATE INDEX IF NOT EXISTS idx_groups_owner ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_public ON groups(is_public);
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_group_members_group ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_posts_group ON group_posts(group_id);
CREATE INDEX IF NOT EXISTS idx_group_posts_user ON group_posts(user_id);
```

**File: `000010_add_groups.down.sql`**
```sql
-- Drop in reverse order (indexes, then tables, respecting foreign keys)
DROP INDEX IF EXISTS idx_group_posts_user;
DROP INDEX IF EXISTS idx_group_posts_group;
DROP INDEX IF EXISTS idx_group_members_group;
DROP INDEX IF EXISTS idx_group_members_user;
DROP INDEX IF EXISTS idx_groups_public;
DROP INDEX IF EXISTS idx_groups_owner;

DROP TABLE IF EXISTS group_posts;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
```

---

### Pattern 2: Renaming a Column (SQLite)

SQLite doesn't support `ALTER TABLE RENAME COLUMN` in older versions:

**File: `000011_rename_user_age_to_birthdate.up.sql`**
```sql
-- SQLite requires recreating the table to rename a column

-- 1. Create new table with correct column name
CREATE TABLE user_new (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,  -- renamed from age
    avatar_url TEXT,
    nickname TEXT UNIQUE,
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Copy data, transforming age to date_of_birth
INSERT INTO user_new
SELECT 
    user_id,
    email,
    username,
    first_name,
    last_name,
    date('now', '-' || age || ' years') as date_of_birth,  -- Convert age to birth date
    avatar_url,
    nickname,
    about_me,
    gender,
    is_private,
    created_at
FROM user;

-- 3. Drop old table
DROP TABLE user;

-- 4. Rename new table
ALTER TABLE user_new RENAME TO user;

-- 5. Recreate any indexes that were on user table
-- (none in this case, but if there were...)
```

---

### Pattern 3: Adding Constraints

**File: `000012_add_email_validation.up.sql`**
```sql
-- SQLite doesn't allow adding constraints to existing columns
-- Must recreate table

CREATE TABLE user_new (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (
        email LIKE '%_@__%.__%'  -- Basic email validation
    ),
    username TEXT NOT NULL UNIQUE CHECK (
        LENGTH(username) >= 3 AND LENGTH(username) <= 50
    ),
    -- ... other columns ...
);

-- Copy data (will fail if data doesn't meet new constraints)
INSERT INTO user_new SELECT * FROM user;

-- Drop and rename
DROP TABLE user;
ALTER TABLE user_new RENAME TO user;
```

---

### Pattern 4: Data Migration with Transformation

**File: `000013_split_user_name.up.sql`**
```sql
-- Scenario: Split "full_name" into "first_name" and "last_name"

-- 1. Add new columns
ALTER TABLE user ADD COLUMN first_name TEXT;
ALTER TABLE user ADD COLUMN last_name TEXT;

-- 2. Migrate data
UPDATE user 
SET 
    first_name = CASE 
        WHEN full_name LIKE '% %' 
        THEN SUBSTR(full_name, 1, INSTR(full_name, ' ') - 1)
        ELSE full_name
    END,
    last_name = CASE 
        WHEN full_name LIKE '% %' 
        THEN SUBSTR(full_name, INSTR(full_name, ' ') + 1)
        ELSE ''
    END;

-- 3. Make columns non-nullable (now that they have data)
-- (Requires recreating table in SQLite)
CREATE TABLE user_new (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,  -- Now NOT NULL
    last_name TEXT NOT NULL,   -- Now NOT NULL
    -- ... other columns (excluding full_name) ...
);

INSERT INTO user_new SELECT 
    user_id, email, username, first_name, last_name, ...
FROM user;

DROP TABLE user;
ALTER TABLE user_new RENAME TO user;
```

---

## Troubleshooting

### Problem: "Dirty Database"

**Symptom:**
```
Error: Dirty database version 5. Fix and force version.
```

**Cause:** A migration partially failed, leaving the database in an inconsistent state.

**Solution:**
```bash
# 1. Check what's in the database
sqlite3 database/forum.db ".schema"

# 2. Manually fix the schema if needed

# 3. Force the version back
migrate -path migrations -database "sqlite3://forum.db" force 4

# 4. Re-run migrations
migrate -path migrations -database "sqlite3://forum.db" up
```

---

### Problem: Migration Fails with "Table Already Exists"

**Symptom:**
```
Error: table "posts" already exists
```

**Cause:** Migration doesn't use `IF NOT EXISTS`.

**Solution 1: Fix the migration file:**
```sql
-- Change from:
CREATE TABLE posts (...);

-- To:
CREATE TABLE IF NOT EXISTS posts (...);
```

**Solution 2: Mark as applied:**
```bash
# If you know the table is correct, just mark migration as applied
migrate -path migrations -database "sqlite3://forum.db" force 5
```

---

### Problem: "Cannot Roll Back Migration"

**Symptom:**
```
Error: Cannot execute "DROP TABLE users" because it contains data
```

**Cause:** Down migration tries to drop table with data.

**Solution:** Make rollback safer:
```sql
-- In .down.sql file
-- Option 1: Backup before drop
CREATE TABLE users_backup AS SELECT * FROM users;
DROP TABLE users;

-- Option 2: Use a flag table to prevent accidental rollback
CREATE TABLE IF NOT EXISTS _migration_guards (
    migration INT PRIMARY KEY,
    allow_rollback BOOLEAN DEFAULT 0
);

-- Check before dropping
-- (This would require custom logic in your app)
```

---

### Problem: "Foreign Key Constraint Failed"

**Symptom:**
```
Error: FOREIGN KEY constraint failed
```

**Cause:** Trying to insert/delete data that violates foreign key.

**Solution:** Always delete in correct order:
```sql
-- DOWN migration for related tables
-- ✅ CORRECT ORDER: Child tables first, parent tables last
DROP TABLE group_posts;      -- references groups
DROP TABLE group_members;    -- references groups
DROP TABLE groups;           -- parent table last

-- ❌ WRONG ORDER: Parent first breaks foreign keys
-- DROP TABLE groups;
-- DROP TABLE group_posts;  -- FAILS: still references groups
```

---

### Problem: "Migration Takes Too Long"

**Symptom:** Migration running for minutes on large tables.

**Cause:** Creating indexes on large tables, or complex data migrations.

**Solutions:**

1. **Create indexes CONCURRENTLY (if supported):**
   ```sql
   -- Not supported in SQLite, but in PostgreSQL:
   CREATE INDEX CONCURRENTLY idx_name ON table(column);
   ```

2. **Split into smaller migrations:**
   ```sql
   -- Instead of one big migration:
   -- 000015_big_change.up.sql (adds 10 indexes)

   -- Split into:
   -- 000015_add_index_1.up.sql
   -- 000016_add_index_2.up.sql
   -- ...
   ```

3. **Batch data migrations:**
   ```sql
   -- Instead of updating all rows at once:
   -- UPDATE posts SET status = 'published';  -- millions of rows

   -- Batch the updates:
   UPDATE posts SET status = 'published' WHERE post_id IN (
       SELECT post_id FROM posts WHERE status IS NULL LIMIT 10000
   );
   -- Run multiple times or use application code
   ```

---

## Summary: Key Takeaways

### ✅ DO:
- Always create both `.up.sql` and `.down.sql` files
- Use `IF EXISTS` / `IF NOT EXISTS` for idempotency
- Keep migrations small and focused
- Test migrations on a copy of production data
- Document complex migrations
- Use transactions (automatic)
- Add columns with defaults
- Delete in order (child → parent)

### ❌ DON'T:
- Don't modify existing migration files after they've been applied
- Don't put application logic in migrations
- Don't forget to handle SQLite's limited ALTER TABLE support
- Don't create migrations without testing rollback
- Don't mix DDL and DML excessively in one migration
- Don't forget foreign key constraints

### Migration Workflow:

```
1. Write migration files (000XXX_description.{up,down}.sql)
2. Test locally:
   - Run up migration
   - Verify schema
   - Test application
   - Run down migration
   - Verify rollback worked
3. Commit migration files to git
4. Deploy to staging
5. Test in staging environment
6. Deploy to production
7. Monitor for errors
```

---

## Quick Reference Card

```bash
# Create new migration (manual)
touch API/pkg/db/migrations/sqlite/000XXX_description.up.sql
touch API/pkg/db/migrations/sqlite/000XXX_description.down.sql

# Run migrations (automatic on app start)
go run cmd/main.go

# Run migrations (manual via CLI)
migrate -path pkg/db/migrations/sqlite -database "sqlite3://database/forum.db" up

# Rollback last migration
migrate -path pkg/db/migrations/sqlite -database "sqlite3://database/forum.db" down 1

# Check current version
migrate -path pkg/db/migrations/sqlite -database "sqlite3://database/forum.db" version

# Force version (after manual fixes)
migrate -path pkg/db/migrations/sqlite -database "sqlite3://database/forum.db" force VERSION

# Check schema
sqlite3 database/forum.db ".schema TABLE_NAME"

# Check migrations table
sqlite3 database/forum.db "SELECT * FROM schema_migrations;"
```

---

This is your complete guide to using golang-migrate in your social network project. Keep this document updated as you add new migration patterns!