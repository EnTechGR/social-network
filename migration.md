# Migration Execution Guide

## Prerequisites
Before running the migration, ensure you have:
- ✅ Updated `schema_config.go` with the new user table definition
- ✅ Updated `CURRENT_DB_VERSION` constant in `database_model.go` from `8` to `9`
- ✅ Added the new migration (version 9) to the `GetMigrations()` function

## Steps to Execute Migration

### 1. Backup Your Database (Automatic)
The migration system automatically creates a backup before running migrations. Backups are stored in `./database/backups/` with a timestamp.

### 2. Run Your Application
Simply start your application as normal:

```bash
go run main.go
```

or

```bash
go run .
```

### 3. What Happens Automatically

When your application starts, the `InitDB()` function will:

1. **Detect** that the current database version (8) is less than `CURRENT_DB_VERSION` (9)
2. **Create a backup** of your database
3. **Execute migration version 9**:
   - Add `first_name` column to user table
   - Add `last_name` column to user table
   - Add `age` column to user table
   - Add `gender` column to user table
   - Update existing users with default values
4. **Update** the database version to 9
5. **Display** success messages in the console

### 4. Expected Console Output

You should see output similar to:

```
Running 1 migration(s)...
✅ Database backup created: ./database/backups/forum_backup_20241104_153045.db
Applying migration 9: Add user profile fields (first_name, last_name, age, gender)
Migration 9 completed
🎉 All migrations completed successfully!
📁 Backup stored at: ./database/backups/forum_backup_20241104_153045.db
Database initialization and migrations completed successfully.
```

## Verification

After the migration completes, you can verify the changes:

### Using SQLite CLI
```bash
sqlite3 ./database/forum.db
```

Then run:
```sql
PRAGMA table_info(user);
```

You should see the new columns: `first_name`, `last_name`, `age`, `gender`

### Check Database Version
```sql
SELECT * FROM database_version ORDER BY version DESC LIMIT 1;
```

Should show version `9`.

## Rollback (If Needed)

If something goes wrong, you can restore from the automatic backup:

1. Locate the backup file in `./database/backups/`
2. Use the restore function or manually replace the database file

```bash
cp ./database/backups/forum_backup_TIMESTAMP.db ./database/forum.db
```

## Troubleshooting

### Migration Fails
- Check the error message carefully
- Look for the backup file path in the error output
- Restore from backup if needed
- Fix the issue and try again

### Column Already Exists Error
If you've already manually added columns, the migration might fail. You can:
1. Update the migration logic to check if columns exist (similar to version 7 & 8)
2. Or restore from a backup before the manual changes

### Application Logic Updates Required

After the migration succeeds, you'll need to update:
1. **User registration handlers** to accept new fields
2. **User models/structs** to include new fields
3. **Validation logic** for age (13-120), gender values, etc.
4. **Frontend forms** to collect the new information

## Next Steps

After successful migration:
1. ✅ Update your User struct in Go code
2. ✅ Update registration handler to accept new fields
3. ✅ Update login handler to accept username OR email
4. ✅ Add validation for new fields
5. ✅ Update frontend registration form
