package handlers

import (
	"strconv"

	"autentikasi/database"
	"autentikasi/dto"
	"autentikasi/models"
	"autentikasi/utils"

	"github.com/gofiber/fiber/v2"
)

// SearchUserByPhone -> search user by phone number
func SearchUserByPhone(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	phone := c.Query("phone")
	if phone == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Phone number is required")
	}

	var foundUser models.User
	if err := database.DB.Where("phone = ?", phone).First(&foundUser).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "User not found")
	}

	// Check if already friends or has pending request
	var friendship models.Friendship
	err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, foundUser.ID, foundUser.ID, user.ID).First(&friendship).Error
	isFriend := err == nil && friendship.Status == "accepted"
	hasPending := err == nil && friendship.Status == "pending"

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"id":          foundUser.ID,
		"nama":        foundUser.Nama,
		"email":       foundUser.Email,
		"phone":       foundUser.Phone,
		"is_friend":   isFriend,
		"has_pending": hasPending,
	})
}

// SendFriendRequest -> send friend request by phone
func SendFriendRequest(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.SendFriendRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}

	var friend models.User
	if err := database.DB.Where("phone = ?", req.Phone).First(&friend).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "User not found")
	}

	if friend.ID == user.ID {
		return utils.Fail(c, fiber.StatusBadRequest, "Cannot send friend request to yourself")
	}

	// Check if already friends or pending
	var existing models.Friendship
	err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friend.ID, friend.ID, user.ID).First(&existing).Error
	if err == nil {
		if existing.Status == "accepted" {
			return utils.Fail(c, fiber.StatusBadRequest, "Already friends")
		}
		if existing.Status == "pending" {
			return utils.Fail(c, fiber.StatusBadRequest, "Friend request already sent")
		}
	}

	// Create friendship with pending status
	friendship := models.Friendship{
		UserID:   user.ID,
		FriendID: friend.ID,
		Status:   "pending",
	}
	if err := database.DB.Create(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to send friend request")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Friend request sent"})
}

// AcceptFriendRequest -> accept friend request
func AcceptFriendRequest(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid friendship ID")
	}

	var friendship models.Friendship
	if err := database.DB.Where("id = ? AND (user_id = ? OR friend_id = ?) AND status = 'pending'",
		id, user.ID, user.ID).First(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Friend request not found")
	}

	friendship.Status = "accepted"
	if err := database.DB.Save(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to accept friend request")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Friend request accepted"})
}

// RejectFriendRequest -> reject friend request
func RejectFriendRequest(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid friendship ID")
	}

	var friendship models.Friendship
	if err := database.DB.Where("id = ? AND (user_id = ? OR friend_id = ?) AND status = 'pending'",
		id, user.ID, user.ID).First(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Friend request not found")
	}

	if err := database.DB.Delete(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to reject friend request")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Friend request rejected"})
}

// ListFriends -> list accepted friends
func ListFriends(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var friendships []models.Friendship
	if err := database.DB.Where("(user_id = ? OR friend_id = ?) AND status = 'accepted'",
		user.ID, user.ID).Find(&friendships).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to fetch friends")
	}

	var friends []fiber.Map
	for _, f := range friendships {
		friendID := f.FriendID
		if f.FriendID == user.ID {
			friendID = f.UserID
		}
		var friend models.User
		if err := database.DB.First(&friend, friendID).Error; err != nil {
			continue
		}
		friends = append(friends, fiber.Map{
			"id":    friend.ID,
			"nama":  friend.Nama,
			"email": friend.Email,
			"phone": friend.Phone,
		})
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"friends": friends})
}

// ListPendingRequests -> list pending friend requests (sent and received)
func ListPendingRequests(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var friendships []models.Friendship
	if err := database.DB.Where("(user_id = ? OR friend_id = ?) AND status = 'pending'",
		user.ID, user.ID).Find(&friendships).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to fetch pending requests")
	}

	var sent []fiber.Map
	var received []fiber.Map
	for _, f := range friendships {
		if f.UserID == user.ID {
			// sent request
			var friend models.User
			if err := database.DB.First(&friend, f.FriendID).Error; err != nil {
				continue
			}
			sent = append(sent, fiber.Map{
				"id":           f.ID,
				"friend_id":    friend.ID,
				"friend_nama":  friend.Nama,
				"friend_email": friend.Email,
			})
		} else {
			// received request
			var friend models.User
			if err := database.DB.First(&friend, f.UserID).Error; err != nil {
				continue
			}
			received = append(received, fiber.Map{
				"id":           f.ID,
				"friend_id":    friend.ID,
				"friend_nama":  friend.Nama,
				"friend_email": friend.Email,
			})
		}
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"sent":     sent,
		"received": received,
	})
}
