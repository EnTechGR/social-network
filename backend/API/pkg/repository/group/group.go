package group

import (
	"database/sql"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"time"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create inserts a new group and automatically adds the creator as a member
func (r *GroupRepository) Create(group models.Group) (*models.Group, error) {
	group.ID = utils.GenerateUUID()
	group.CreatedAt = time.Now()

	// Start transaction to create group and add owner as member atomically
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert group
	_, err = tx.Exec(
		`INSERT INTO groups (group_id, owner_id, title, description, created_at) 
		VALUES (?, ?, ?, ?, ?)`,
		group.ID, group.OwnerID, group.Title, group.Description, group.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Add creator as a member
	_, err = tx.Exec(
		`INSERT INTO group_members (group_id, user_id, joined_at) 
		VALUES (?, ?, ?)`,
		group.ID, group.OwnerID, group.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &group, nil
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(groupID string) (*models.Group, error) {
	var g models.Group
	err := r.db.QueryRow(
		`SELECT group_id, owner_id, title, description, created_at 
		FROM groups WHERE group_id = ?`,
		groupID,
	).Scan(&g.ID, &g.OwnerID, &g.Title, &g.Description, &g.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// GetByIDWithOwner retrieves a group with owner details and member count
func (r *GroupRepository) GetByIDWithOwner(groupID string) (*models.GroupWithOwner, error) {
	var g models.GroupWithOwner
	err := r.db.QueryRow(
		`SELECT 
			g.group_id, g.owner_id, u.nickname, g.title, g.description, g.created_at,
			COUNT(gm.user_id) as member_count
		FROM groups g
		INNER JOIN user u ON g.owner_id = u.user_id
		LEFT JOIN group_members gm ON g.group_id = gm.group_id
		WHERE g.group_id = ?
		GROUP BY g.group_id`,
		groupID,
	).Scan(&g.ID, &g.OwnerID, &g.OwnerNickname, &g.Title, &g.Description, &g.CreatedAt, &g.MemberCount)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// GetAll retrieves all groups with owner details and member count (for browsing)
func (r *GroupRepository) GetAll() ([]models.GroupWithOwner, error) {
	rows, err := r.db.Query(
		`SELECT 
			g.group_id, g.owner_id, u.nickname, g.title, g.description, g.created_at,
			COUNT(gm.user_id) as member_count
		FROM groups g
		INNER JOIN user u ON g.owner_id = u.user_id
		LEFT JOIN group_members gm ON g.group_id = gm.group_id
		GROUP BY g.group_id
		ORDER BY g.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.GroupWithOwner
	for rows.Next() {
		var g models.GroupWithOwner
		if err := rows.Scan(&g.ID, &g.OwnerID, &g.OwnerNickname, &g.Title, &g.Description, &g.CreatedAt, &g.MemberCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetUserGroups retrieves all groups a user is a member of
func (r *GroupRepository) GetUserGroups(userID string) ([]models.GroupWithOwner, error) {
	rows, err := r.db.Query(
		`SELECT 
			g.group_id, g.owner_id, u.nickname, g.title, g.description, g.created_at,
			COUNT(gm2.user_id) as member_count
		FROM groups g
		INNER JOIN user u ON g.owner_id = u.user_id
		INNER JOIN group_members gm ON g.group_id = gm.group_id
		LEFT JOIN group_members gm2 ON g.group_id = gm2.group_id
		WHERE gm.user_id = ?
		GROUP BY g.group_id
		ORDER BY g.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.GroupWithOwner
	for rows.Next() {
		var g models.GroupWithOwner
		if err := rows.Scan(&g.ID, &g.OwnerID, &g.OwnerNickname, &g.Title, &g.Description, &g.CreatedAt, &g.MemberCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// Update updates a group's title and description
func (r *GroupRepository) Update(groupID, title string, description *string) error {
	result, err := r.db.Exec(
		`UPDATE groups SET title = ?, description = ? WHERE group_id = ?`,
		title, description, groupID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete removes a group (CASCADE will handle members, invites, events, etc.)
func (r *GroupRepository) Delete(groupID string) error {
	result, err := r.db.Exec(`DELETE FROM groups WHERE group_id = ?`, groupID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// IsUserMember checks if a user is a member of a group
func (r *GroupRepository) IsUserMember(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsUserOwner checks if a user is the owner of a group
func (r *GroupRepository) IsUserOwner(groupID, userID string) (bool, error) {
	var ownerID string
	err := r.db.QueryRow(
		`SELECT owner_id FROM groups WHERE group_id = ?`,
		groupID,
	).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return ownerID == userID, nil
}

// GetMemberCount returns the number of members in a group
func (r *GroupRepository) GetMemberCount(groupID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM group_members WHERE group_id = ?`,
		groupID,
	).Scan(&count)
	return count, err
}