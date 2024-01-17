// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type ConversationList struct {
	LastMessageTime string `json:"lastMessageTime"`
	ConversationID  string `json:"conversationId"`
	IsItGroup       bool   `json:"isItGroup"`
}

type DeleteGroupConversationInput struct {
	GroupID   string `json:"groupId"`
	MessageID string `json:"messageId"`
}

type Group struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	AdminID   string `json:"adminId"`
}

type GroupConversation struct {
	ID        string `json:"id"`
	GroupID   string `json:"groupId"`
	SenderID  string `json:"senderId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type GroupConversationPublishedInput struct {
	GroupID  string `json:"groupId"`
	MemberID string `json:"memberId"`
}

type GroupDetails struct {
	GroupID      string                `json:"groupId"`
	Name         string                `json:"name"`
	AdminID      string                `json:"adminId"`
	CreatedAt    string                `json:"createdAt"`
	GroupMembers []*GroupMemberDetails `json:"groupMembers"`
}

type GroupMemberDetails struct {
	MemberID  string  `json:"memberId"`
	IsRemoved bool    `json:"isRemoved"`
	RemovedAt *string `json:"removedAt,omitempty"`
}

type GroupMembersInput struct {
	GroupID  string   `json:"groupId"`
	MemberID []string `json:"memberId"`
}

type Mutation struct {
}

type NewGroup struct {
	Name string `json:"name"`
}

type NewGroupConversation struct {
	GroupID string `json:"groupId"`
	Content string `json:"content"`
}

type NewPersonalConversation struct {
	ReceiverID string `json:"receiverId"`
	Content    string `json:"content"`
}

type PersonalConversation struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
}

type PersonalConversationPublishedInput struct {
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
}

type Query struct {
}

type Subscription struct {
}

type User struct {
	ID        string  `json:"id"`
	Fullname  string  `json:"fullname"`
	Email     string  `json:"email"`
	IPAddress string  `json:"ipAddress"`
	Gender    *string `json:"gender,omitempty"`
}

type UserListInput struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Limit *int    `json:"limit,omitempty"`
	Page  *int    `json:"page,omitempty"`
}
