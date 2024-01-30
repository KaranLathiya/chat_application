package groupChat

import (
	"chat_application/api/auth"
	"chat_application/api/chatCommon"
	"chat_application/api/dal"
	"chat_application/api/dataloader"
	"chat_application/api/user"
	"chat_application/graph/model"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func CreateGroup(ctx context.Context, input model.NewGroup) (*model.Group, error) {
	var group model.Group
	db := dal.GetDB()
	adminID := ctx.Value(auth.UserCtxKey).(string)
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.QueryRow("INSERT INTO public.groups( name, admin_id, created_at) VALUES ( $1, $2, $3)  RETURNING id, created_at;", input.Name, adminID, currentFormattedTime).Scan(&group.ID, &group.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO public.group_members (group_id, member_id) VALUES ($1, $2)", group.ID, adminID)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf("Group created :- %v", input.Name)
	_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4);", group.ID, adminID, content, currentFormattedTime)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	group.AdminID = adminID
	group.Name = input.Name
	return &group, nil
}

func GroupDetails(ctx context.Context, groupID string) (*model.GroupDetails, error) {
	memberID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()

	var groupDetails model.GroupDetails
	errIfNoRows := db.QueryRow(
		"SELECT name, admin_id, created_at FROM public.groups WHERE id=$1;", groupID).Scan(&groupDetails.Name, &groupDetails.AdminID, &groupDetails.CreatedAt)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("invalid groupid")
	}
	fmt.Println("group details called")
	var removedFromGroup bool
	errIfNoRows = db.QueryRow(
		"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", memberID, groupID).Scan(&removedFromGroup)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("user is not member of group")
	}
	groupDetails.GroupID = groupID
	if removedFromGroup {
		groupDetails.IsRemoved = true
		return &groupDetails, nil
	}

	return &groupDetails, nil
}

func GroupMembers(ctx context.Context, obj *model.GroupDetails) ([]*model.GroupMemberDetails, error) {
	if obj.IsRemoved {
		return nil, nil
	}
	db := dal.GetDB()
	fmt.Println("members finding")
	rows, err := db.Query("SELECT member_id, is_removed, removed_at FROM public.group_members WHERE group_id=$1;", obj.GroupID)
	if err != nil {
		return nil, err
	}
	var groupMembersDetails []*model.GroupMemberDetails
	for rows.Next() {
		var groupMemberDetails model.GroupMemberDetails
		err = rows.Scan(&groupMemberDetails.MemberID, &groupMemberDetails.IsRemoved, &groupMemberDetails.RemovedAt)
		if err != nil {
			return nil, err
		}
		groupMembersDetails = append(groupMembersDetails, &groupMemberDetails)
	}
	// obj.GroupMembers = GroupMembersDetails
	return groupMembersDetails, nil
}

func MembersListThatCanJoinTheGroup(ctx context.Context, input model.MembersListThatCanJoinTheGroupInput) ([]*model.User, error) {
	db := dal.GetDB()
	userID := ctx.Value(auth.UserCtxKey).(string)
	err := checkUserIsAdminOfGroup(db, userID, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("user is not admin group")
	}
	offset := (*input.Page - 1) * *input.Limit
	var where, orderBy []string
	var andKeyword string
	var filterArgsList []interface{}

	if input.Name != nil && *input.Name != "" {
		where = append(where, "fullname ILIKE '%' || ? || '%'")
		filterArgsList = append(filterArgsList, *input.Name)
		orderBy = append(orderBy, "POSITION (LOWER('"+*input.Name+"') IN LOWER(fullname)) ASC")
	}
	if input.Email != nil && *input.Email != "" {
		where = append(where, "email ILIKE '%' || ? || '%'")
		filterArgsList = append(filterArgsList, *input.Email)
		orderBy = append(orderBy, "POSITION (LOWER('"+*input.Email+"') IN LOWER(email)) ASC")
	}
	if len(where) > 0 {
		postCondition := strings.Join(where, " OR ")
		andKeyword = "AND (" + postCondition + ")"
	} else {
		orderBy = append(orderBy, "fullname ASC")
	}
	preCondition := "id IN (SELECT id FROM public.users EXCEPT SELECT member_id FROM public.group_members WHERE group_id = " + input.GroupID + " AND is_removed = false )"
	query := fmt.Sprintf("SELECT id, fullname, email, ip_address, gender FROM public.users WHERE %s %s ORDER BY %v LIMIT %d OFFSET %d", preCondition, andKeyword, strings.Join(orderBy, " , "), *input.Limit, offset)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	// fmt.Println(query)
	rows, err := db.Query(query, filterArgsList...)
	if err != nil {
		return nil, err
	}
	var users []*model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.FullName, &user.Email, &user.IPAddress, &user.Gender)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func AddGroupMembers(ctx context.Context, input model.GroupMembersInput) (bool, error) {
	adminID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	err := checkUserIsAdminOfGroup(db, adminID, input.GroupID)
	if err != nil {
		return false, fmt.Errorf("user is not admin group")
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	for i, _ := range input.MemberID {
		var removedFromGroup bool
		errIfNoRows := tx.QueryRow(
			"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", input.MemberID[i], input.GroupID).Scan(&removedFromGroup)
		if errIfNoRows != nil {
			_, err := tx.Exec("INSERT INTO public.group_members (group_id, member_id) VALUES ($1, $2);", input.GroupID, input.MemberID[i])
			if err != nil {
				return false, fmt.Errorf("invalid memberid or group id")
			}
		} else if removedFromGroup {
			_, err := tx.Exec("UPDATE public.group_members SET is_removed = false, removed_at = null where group_id = $1 and member_id = $2 ;", input.GroupID, input.MemberID[i])
			if err != nil {
				return false, fmt.Errorf("invalid memberid")
			}
		} else {
			continue
		}
		memberName, err := user.UserNameByID(ctx, input.MemberID[i])
		if err != nil {
			return false, fmt.Errorf("invalid memberid")
		}
		content := fmt.Sprintf("%s added in to the group", memberName)
		_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4);", input.GroupID, adminID, content, currentFormattedTime)
		if err != nil {
			return false, err
		}
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

// func currentGroupMembers(db *sql.DB, groupID string) ([]string, error) {
// 	rows, err := db.Query("SELECT member_id FROM public.group_members WHERE group_id=$1 and is_removed = false;", groupID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var groupMembers []string
// 	var groupMember string
// 	for rows.Next() {
// 		err = rows.Scan(&groupMember)
// 		if err != nil {
// 			return nil, err
// 		}
// 		groupMembers = append(groupMembers, groupMember)
// 	}
// 	// obj.GroupMembers = GroupMembersDetails
// 	return groupMembers, nil
// }

func RemoveGroupMembers(ctx context.Context, input model.GroupMembersInput) (bool, error) {
	adminID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	err := checkUserIsAdminOfGroup(db, adminID, input.GroupID)
	if err != nil {
		return false, fmt.Errorf("user is not admin of group")
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	for i, _ := range input.MemberID {
		if adminID == input.MemberID[i] {
			return false, fmt.Errorf("admin can't leave group without assigning new admin")
		}
		_, err = tx.Exec("UPDATE public.group_members SET is_removed=true, removed_at=$1 WHERE group_id=$2 AND member_id=$3;", currentFormattedTime, input.GroupID, input.MemberID[i])
		if err != nil {
			return false, err
		}
		memberName, err := user.UserNameByID(ctx, input.MemberID[i])
		if err != nil {
			return false, fmt.Errorf("invalid memberid")
		}
		content := fmt.Sprintf("%s removed from the group", memberName)
		_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4);", input.GroupID, adminID, content, currentFormattedTime)
		if err != nil {
			return false, err
		}
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func ChangeGroupAdmin(ctx context.Context, input model.ChangeGroupAdminInput) (bool, error) {
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	adminID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	err := checkUserIsAdminOfGroup(db, adminID, input.GroupID)
	if err != nil {
		return false, fmt.Errorf("user is not admin group")
	}
	if adminID == input.NewAdminID {
		return false, fmt.Errorf("member is already admin of group")
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE public.groups SET admin_id=$1 WHERE id=$2 AND admin_id=$3;", input.NewAdminID, input.GroupID, adminID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false, fmt.Errorf("invalid memberid")
	}
	newAdminName, err := user.UserNameByID(ctx, input.NewAdminID)
	if err != nil {
		return false, fmt.Errorf("invalid memberid")
	}
	content := fmt.Sprintf("%s is the new admin of the group", newAdminName)
	_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)", input.GroupID, adminID, content, currentFormattedTime)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func ChangeGroupName(ctx context.Context, input model.ChangeGroupNameInput) (bool, error) {
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	adminID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	err := checkUserIsAdminOfGroup(db, adminID, input.GroupID)
	if err != nil {
		return false, fmt.Errorf("user is not admin group")
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE public.groups SET name=$1 WHERE name != $1 AND id=$2 AND admin_id=$3 ;", input.NewGroupName, input.GroupID, adminID)
	if err != nil {
		return false, err
	}
	RowsAffected, _ := result.RowsAffected()
	fmt.Println(RowsAffected)
	if RowsAffected == 0 {
		return false, fmt.Errorf("no change in group name")
	}
	content := fmt.Sprintf("Group name changed to :- %s ", input.NewGroupName)
	_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)", input.GroupID, adminID, content, currentFormattedTime)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func LeaveGroup(ctx context.Context, groupID string) (bool, error) {
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	memberID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	err := checkUserIsAdminOfGroup(db, memberID, groupID)
	if err == nil {
		return false, fmt.Errorf("admin can not leave group directly")
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec("UPDATE public.group_members SET is_removed=true, removed_at=$1 WHERE is_removed=false AND group_id=$2 AND member_id=$3;", currentFormattedTime, groupID, memberID)
	if err != nil {
		return false, err
	}
	RowsAffected, _ := result.RowsAffected()
	fmt.Println(RowsAffected)
	if RowsAffected == 0 {
		return false, fmt.Errorf("user is not member of this group")
	}
	memberName, err := user.UserNameByID(ctx, memberID)
	if err != nil {
		return false, fmt.Errorf("invalid memberid")
	}
	content := fmt.Sprintf("%s left the group", memberName)
	_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4);", groupID, memberID, content, currentFormattedTime)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func SenderDetails(ctx context.Context, obj *model.GroupConversation) (*model.Sender, error) {
	user, err := ctx.Value(dataloader.CtxKey).(*dataloader.UserLoader).Load(*obj.SenderID)
	return user, err
}

func checkUserIsAdminOfGroup(db *sql.DB, adminID string, groupID string) error {
	errIfNoRows := db.QueryRow("SELECT admin_id FROM public.groups WHERE admin_id = $1 AND id = $2;", adminID, groupID).Scan(&adminID)
	// fmt.Println(adminID, groupID)
	return errIfNoRows
}
