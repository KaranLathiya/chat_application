package groupChat

import (
	"chat_application/api/auth"
	"chat_application/api/chatCommon"
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func CreateGroup(ctx context.Context, input model.NewGroup) (*model.Group, error) {
	var Group model.Group
	db := dal.GetDB()
	adminID := ctx.Value(auth.UserCtxKey).(string)
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.QueryRow("INSERT INTO public.groups( name, admin_id, created_at) VALUES ( $1, $2, $3)  RETURNING id, created_at;", input.Name, adminID, currentFormattedTime).Scan(&Group.ID, &Group.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO public.group_members (group_id, member_id) VALUES ($1, $2)", Group.ID, adminID)
	if err != nil {
		return nil, err
	}

	content := fmt.Sprintf("Group created :- %v", input.Name)
	_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)  RETURNING id, created_at;", Group.ID, adminID, content, currentFormattedTime)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	Group.AdminID = adminID
	Group.Name = input.Name
	return &Group, nil
}

func GroupDetails(ctx context.Context, groupID string) (*model.GroupDetails, error) {
	memberID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()

	var GroupDetails model.GroupDetails
	errIfNoRows := db.QueryRow(
		"SELECT name, admin_id, created_at FROM public.groups WHERE id=$1;", groupID).Scan(&GroupDetails.Name, &GroupDetails.AdminID, &GroupDetails.CreatedAt)
	if errIfNoRows != nil {
		return nil, errIfNoRows
	}
	fmt.Println("group details called")
	var removedFromGroup bool
	errIfNoRows = db.QueryRow(
		"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", memberID, groupID).Scan(&removedFromGroup)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("user is not member of group")
	}
	GroupDetails.GroupID = groupID
	if removedFromGroup {
		return &GroupDetails, nil
	}

	return &GroupDetails, nil
}

func GroupMembers(ctx context.Context, obj *model.GroupDetails) ([]*model.GroupMemberDetails, error) {
	db := dal.GetDB()
	fmt.Println("members finding")
	rows, err := db.Query("SELECT member_id, is_removed, removed_at FROM public.group_members WHERE group_id=$1;", obj.GroupID)
	if err != nil {
		return nil, err
	}
	var GroupMembersDetails []*model.GroupMemberDetails
	for rows.Next() {
		var GroupMemberDetails model.GroupMemberDetails
		err = rows.Scan(&GroupMemberDetails.MemberID, &GroupMemberDetails.IsRemoved, &GroupMemberDetails.RemovedAt)
		if err != nil {
			return nil, err
		}
		GroupMembersDetails = append(GroupMembersDetails, &GroupMemberDetails)
	}
	// obj.GroupMembers = GroupMembersDetails
	return GroupMembersDetails, nil
}

func AddableMembersInGroup(ctx context.Context, input model.AddableMembersInGroupInput) ([]*model.User, error) {
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
	fmt.Println(query)
	rows, err := db.Query(query, filterArgsList...)
	if err != nil {
		return nil, err
	}
	var Users []*model.User
	for rows.Next() {
		var User model.User
		err = rows.Scan(&User.ID, &User.Fullname, &User.Email, &User.IPAddress, &User.Gender)
		if err != nil {
			return nil, err
		}
		Users = append(Users, &User)
	}
	return Users, nil
}

func AddGroupMembers( ctx context.Context, input model.GroupMembersInput) (bool, error) {
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
		_, err := tx.Exec("UPSERT INTO public.group_members (group_id, member_id) VALUES ($1, $2);", input.GroupID, input.MemberID[i])
		if err != nil {
			return false, fmt.Errorf("wrong memberid")
		}
		memberName, err := chatCommon.UserNameByID(input.MemberID[i])
		if err != nil {
			return false, fmt.Errorf("wrong memberid")
		}
		content := fmt.Sprintf("%s added in to the group", memberName)
		_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)  RETURNING id, created_at;", input.GroupID, adminID, content, currentFormattedTime)
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

func RemoveGroupMembers(ctx context.Context, input model.GroupMembersInput) (bool, error) {
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
		_, err = tx.Exec("UPDATE public.group_members SET is_removed=true, removed_at=$1 WHERE group_id=$2 AND member_id=$3;", currentFormattedTime, input.GroupID, input.MemberID[i])
		if err != nil {
			return false, err
		}
		memberName, err := chatCommon.UserNameByID(input.MemberID[i])
		if err != nil {
			return false, fmt.Errorf("wrong memberid")
		}
		content := fmt.Sprintf("%s removed from the group", memberName)
		_, err = tx.Exec("INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)  RETURNING id, created_at;", input.GroupID, adminID, content, currentFormattedTime)
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

func checkUserIsAdminOfGroup(db *sql.DB, adminID string, groupID string) error {
	errIfNoRows := db.QueryRow("SELECT admin_id FROM public.groups WHERE admin_id = $1 AND id = $2;", adminID, groupID).Scan(&adminID)
	fmt.Println(adminID, groupID)
	return errIfNoRows
}
