package dto

type SendFriendRequest struct {
	Phone string `json:"phone" validate:"required"`
}

type AcceptFriendRequest struct {
	FriendshipID uint64 `json:"friendship_id" validate:"required"`
}

type RejectFriendRequest struct {
	FriendshipID uint64 `json:"friendship_id" validate:"required"`
}
