package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.43

import (
	"chat_application/api/groupChat"
	"chat_application/api/personalChat"
	"chat_application/api/user"
	"chat_application/graph/model"
	"context"
)

// Sender is the resolver for the sender field.
func (r *groupConversationResolver) Sender(ctx context.Context, obj *model.GroupConversation) (*model.Sender, error) {
	return groupChat.SenderDetails(ctx, obj)
}

// GroupMembers is the resolver for the groupMembers field.
func (r *groupDetailsResolver) GroupMembers(ctx context.Context, obj *model.GroupDetails) ([]*model.GroupMemberDetails, error) {
	return groupChat.GroupMembers(ctx, obj)
}

// CreatePersonalConversation is the resolver for the createPersonalConversation field.
func (r *mutationResolver) CreatePersonalConversation(ctx context.Context, input model.NewPersonalConversation) (*model.PersonalConversation, error) {
	return personalChat.CreatePersonalConversation(ctx, input)
}

// CreateGroupConversation is the resolver for the createGroupConversation field.
func (r *mutationResolver) CreateGroupConversation(ctx context.Context, input model.NewGroupConversation) (*model.GroupConversation, error) {
	return groupChat.CreateGroupConversation(ctx, input)
}

// CreateGroup is the resolver for the createGroup field.
func (r *mutationResolver) CreateGroup(ctx context.Context, input model.NewGroup) (*model.Group, error) {
	return groupChat.CreateGroup(ctx, input)
}

// AddGroupMembers is the resolver for the addGroupMembers field.
func (r *mutationResolver) AddGroupMembers(ctx context.Context, input model.GroupMembersInput) (bool, error) {
	return groupChat.AddGroupMembers(ctx, input)
}

// RemoveGroupMembers is the resolver for the removeGroupMembers field.
func (r *mutationResolver) RemoveGroupMembers(ctx context.Context, input model.GroupMembersInput) (bool, error) {
	return groupChat.RemoveGroupMembers(ctx, input)
}

// DeletePersonalConversation is the resolver for the deletePersonalConversation field.
func (r *mutationResolver) DeletePersonalConversation(ctx context.Context, messageID string) (bool, error) {
	return personalChat.DeletePersonalConversation(ctx, messageID)
}

// DeleteGroupConversation is the resolver for the deleteGroupConversation field.
func (r *mutationResolver) DeleteGroupConversation(ctx context.Context, input model.DeleteGroupConversationInput) (bool, error) {
	return groupChat.DeleteGroupConversation(ctx, input)
}

// ChangeGroupAdmin is the resolver for the changeGroupAdmin field.
func (r *mutationResolver) ChangeGroupAdmin(ctx context.Context, input model.ChangeGroupAdminInput) (bool, error) {
	return groupChat.ChangeGroupAdmin(ctx, input)
}

// ChangeGroupName is the resolver for the changeGroupName field.
func (r *mutationResolver) ChangeGroupName(ctx context.Context, input model.ChangeGroupNameInput) (bool, error) {
	return groupChat.ChangeGroupName(ctx, input)
}

// LeaveGroup is the resolver for the leaveGroup field.
func (r *mutationResolver) LeaveGroup(ctx context.Context, groupID string) (bool, error) {
	return groupChat.LeaveGroup(ctx, groupID)
}

// GroupDetails is the resolver for the GroupDetails field.
func (r *queryResolver) GroupDetails(ctx context.Context, groupID string) (*model.GroupDetails, error) {
	return groupChat.GroupDetails(ctx, groupID)
}

// UserList is the resolver for the UserList field.
func (r *queryResolver) UserList(ctx context.Context, input *model.UserListInput) ([]*model.User, error) {
	return user.UserList(ctx, input)
}

// MembersListThatCanJoinTheGroup is the resolver for the MembersListThatCanJoinTheGroup field.
func (r *queryResolver) MembersListThatCanJoinTheGroup(ctx context.Context, input model.MembersListThatCanJoinTheGroupInput) ([]*model.User, error) {
	return groupChat.MembersListThatCanJoinTheGroup(ctx, input)
}

// UserDetailsByID is the resolver for the UserDetailsByID field.
func (r *queryResolver) UserDetailsByID(ctx context.Context, userID string) (*model.User, error) {
	return user.UserDetailsByID(ctx, userID)
}

// UserNameByID is the resolver for the UserNameById field.
func (r *queryResolver) UserNameByID(ctx context.Context, userID string) (string, error) {
	return user.UserNameByID(ctx, userID)
}

// PersonalConversationRecords is the resolver for the PersonalConversationRecords field.
func (r *queryResolver) PersonalConversationRecords(ctx context.Context, limit *int, offset *int, receiverID string) ([]*model.PersonalConversation, error) {
	return personalChat.PersonalConversationRecords(ctx, limit, offset, receiverID)
}

// GroupConversationRecords is the resolver for the GroupConversationRecords field.
func (r *queryResolver) GroupConversationRecords(ctx context.Context, limit *int, offset *int, groupID string) ([]*model.GroupConversation, error) {
	return groupChat.GroupConversationRecords(ctx, limit, offset, groupID)
}

// RecentConversationList is the resolver for the RecentConversationList field.
func (r *queryResolver) RecentConversationList(ctx context.Context, limit *int, offset *int) ([]*model.ConversationList, error) {
	return user.RecentConversationList(ctx, limit, offset)
}

// PersonalConversationNotification is the resolver for the personalConversationNotification field.
func (r *subscriptionResolver) PersonalConversationNotification(ctx context.Context, input model.PersonalConversationNotificationInput) (<-chan *model.PersonalConversation, error) {
	return personalChat.PersonalConversationNotification(ctx, input)
}

// GroupConversationNotification is the resolver for the groupConversationNotification field.
func (r *subscriptionResolver) GroupConversationNotification(ctx context.Context, input model.GroupConversationNotificationInput) (<-chan *model.GroupConversation, error) {
	return groupChat.GroupConversationNotification(ctx, input)
}

// GroupConversation returns GroupConversationResolver implementation.
func (r *Resolver) GroupConversation() GroupConversationResolver {
	return &groupConversationResolver{r}
}

// GroupDetails returns GroupDetailsResolver implementation.
func (r *Resolver) GroupDetails() GroupDetailsResolver { return &groupDetailsResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type groupConversationResolver struct{ *Resolver }
type groupDetailsResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
