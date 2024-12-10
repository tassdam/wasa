package database

// Group represents a group (like a group chat).
type Group struct {
	ID   string
	Name string
}

// GroupMember represents a user’s membership in a group.
type GroupMember struct {
	UserID   string
	JoinedAt string // Adjust type if you store as DATETIME
}

// CreateGroup creates a new group.
func (db *appdbimpl) CreateGroup(g Group) error {
	_, err := db.c.Exec(`INSERT INTO groups (id, name) VALUES (?, ?)`,
		g.ID, g.Name)
	return err
}

// AddUserToGroup adds a user to a group.
func (db *appdbimpl) AddUserToGroup(groupID, userID string, joinedAt string) error {
	_, err := db.c.Exec(`INSERT INTO group_members (group_id, user_id, joined_at) VALUES (?, ?, ?)`,
		groupID, userID, joinedAt)
	return err
}

// LeaveGroup removes the authenticated user from a group.
func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	_, err := db.c.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID)
	return err
}

// SetGroupName updates the group's name.
func (db *appdbimpl) SetGroupName(groupID, newName string) error {
	_, err := db.c.Exec(`UPDATE groups SET name = ? WHERE id = ?`,
		newName, groupID)
	return err
}

// GetGroup retrieves group details.
func (db *appdbimpl) GetGroup(groupID string) (Group, error) {
	var g Group
	err := db.c.QueryRow(`SELECT id, name FROM groups WHERE id = ?`,
		groupID).Scan(&g.ID, &g.Name)
	return g, err
}
