package database

import (
	"database/sql"
)

// AddGroupMember adds a user to a group.
// Assumes groupId and userId exist, and joinedAt is the timestamp of when the user joined.
// Returns the inserted GroupMember or an error.
func (db *appdbimpl) AddGroupMember(groupId string, userId string, joinedAt string) (GroupMember, error) {
	gm := GroupMember{
		UserId:   userId,
		JoinedAt: joinedAt,
	}
	_, err := db.c.Exec(`INSERT INTO group_members (groupId, userId, joinedAt) VALUES (?, ?, ?)`,
		groupId, userId, joinedAt)
	if err != nil {
		return gm, err
	}
	return gm, nil
}

// RemoveUserFromGroup removes the authenticated user from a group.
// If no row is affected, it returns ErrGroupMemberDoesNotExist.
func (db *appdbimpl) RemoveUserFromGroup(groupId string, userId string) error {
	res, err := db.c.Exec(`DELETE FROM group_members WHERE groupId=? AND userId=?`, groupId, userId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrGroupMemberDoesNotExist
	}
	return nil
}

// UpdateGroupName sets a new name for the group.
// If no group is updated, returns ErrGroupDoesNotExist.
func (db *appdbimpl) UpdateGroupName(groupId string, newName string) (Group, error) {
	var g Group
	res, err := db.c.Exec(`UPDATE groups SET name=? WHERE id=?`, newName, groupId)
	if err != nil {
		return g, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return g, err
	} else if affected == 0 {
		return g, ErrGroupDoesNotExist
	}

	// Return the updated group
	if err := db.c.QueryRow(`SELECT id, name FROM groups WHERE id=?`, groupId).Scan(&g.Id, &g.Name); err != nil {
		if err == sql.ErrNoRows {
			return g, ErrGroupDoesNotExist
		}
		return g, err
	}
	return g, nil
}

// SetGroupPhoto updates the group's photo stored in `groups` table.
// This assumes you've added a `photo BLOB` column to your `groups` table.
// If you haven't, you need to update the schema accordingly.
// If no group is found, returns ErrGroupDoesNotExist.
func (db *appdbimpl) SetGroupPhoto(groupId string, photo []byte) (Group, error) {
	var g Group
	res, err := db.c.Exec(`UPDATE groups SET photo=? WHERE id=?`, photo, groupId)
	if err != nil {
		return g, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return g, err
	} else if affected == 0 {
		return g, ErrGroupDoesNotExist
	}

	// Return the group without the photo field if your schema does not define it in the Group struct
	// If you need to retrieve and return the photo, modify your Group struct and SELECT it here.
	if err := db.c.QueryRow(`SELECT id, name FROM groups WHERE id=?`, groupId).Scan(&g.Id, &g.Name); err != nil {
		if err == sql.ErrNoRows {
			return g, ErrGroupDoesNotExist
		}
		return g, err
	}

	return g, nil
}

// (Optional) GetGroupById fetches a group by id, if needed for verification.
func (db *appdbimpl) GetGroupById(groupId string) (Group, error) {
	var g Group
	if err := db.c.QueryRow(`SELECT id, name FROM groups WHERE id=?`, groupId).Scan(&g.Id, &g.Name); err != nil {
		if err == sql.ErrNoRows {
			return g, ErrGroupDoesNotExist
		}
		return g, err
	}
	return g, nil
}
