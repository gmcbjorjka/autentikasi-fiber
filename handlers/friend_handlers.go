package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"autentikasi/database"
	"autentikasi/dto"
	"autentikasi/models"
	"autentikasi/utils"

	"github.com/gofiber/fiber/v2"
)

// SearchUserByPhone -> search user by phone number or name (returns multiple results)
// Accepts ?phone=, ?name=, or auto-detects if query looks like phone
func SearchUserByPhone(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	phone := strings.TrimSpace(c.Query("phone"))
	name := strings.TrimSpace(c.Query("name"))
	query := strings.TrimSpace(c.Query("q")) // generic query param

	// If no explicit phone/name, use 'q' and auto-detect
	if phone == "" && name == "" {
		if query != "" {
			// Auto-detect: check if query LOOKS LIKE a phone number
			// Phone formats: +62..., 62..., 0...
			// Contains ONLY digits, +, and spaces? -> Try as phone first
			hasNonDigitNonPhone := false
			digitCount := 0
			for _, r := range query {
				if r >= '0' && r <= '9' {
					digitCount++
				} else if r == '+' || r == ' ' || r == '-' || r == '(' || r == ')' {
					// Phone format characters - OK
				} else {
					// Has letters or other chars - NOT phone format
					hasNonDigitNonPhone = true
					break
				}
			}

			// Treat as phone if: >= 5 digits AND no non-digit/non-phone-format chars
			if digitCount >= 5 && !hasNonDigitNonPhone {
				phone = query
			} else {
				name = query
			}
		}
	}

	if phone == "" && name == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Phone or name is required")
	}

	// Debug logging
	if name != "" {
		fmt.Printf("[SEARCH DEBUG] Auto-detected as NAME: '%s'\n", name)
	}
	if phone != "" {
		fmt.Printf("[SEARCH DEBUG] Auto-detected as PHONE: '%s'\n", phone)
	}

	var foundUsers []models.User
	found := false

	// Try search by phone first if provided or detected
	if phone != "" {
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, phone)

		if digits != "" {
			variants := []string{digits}
			if strings.HasPrefix(digits, "0") {
				variants = append(variants, "62"+digits[1:])
			} else if strings.HasPrefix(digits, "62") {
				variants = append(variants, "0"+digits[2:])
			}

			// Try phone_digits column first
			if err := database.DB.Where("phone_digits IN ?", variants).Find(&foundUsers).Error; err == nil && len(foundUsers) > 0 {
				found = true
			} else if err := database.DB.Where("REPLACE(REPLACE(REPLACE(phone, '+', ''), ' ', ''), '-', '') IN ?", variants).Find(&foundUsers).Error; err == nil && len(foundUsers) > 0 {
				found = true
			}
		}
	}

	// Try search by name if provided and phone search failed/not provided
	if !found && name != "" {
		if err := database.DB.Where("LOWER(nama) LIKE ?", "%"+strings.ToLower(name)+"%").Find(&foundUsers).Error; err == nil && len(foundUsers) > 0 {
			found = true
		}
	}

	if !found || len(foundUsers) == 0 {
		return utils.Fail(c, fiber.StatusNotFound, "User not found")
	}

	// Build response with multiple users
	var results []fiber.Map
	for _, foundUser := range foundUsers {
		// Check if already friends or has pending request
		var friendship models.Friendship
		err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			user.ID, foundUser.ID, foundUser.ID, user.ID).First(&friendship).Error
		isFriend := err == nil && friendship.Status == "accepted"
		hasPending := err == nil && friendship.Status == "pending"

		results = append(results, fiber.Map{
			"id":          foundUser.ID,
			"nama":        foundUser.Nama,
			"email":       foundUser.Email,
			"phone":       foundUser.Phone,
			"is_friend":   isFriend,
			"has_pending": hasPending,
		})
	}

	// If only 1 result, return it directly (backward compatible)
	// If multiple, return array
	if len(results) == 1 {
		return utils.Ok(c, fiber.StatusOK, results[0])
	}
	return utils.Ok(c, fiber.StatusOK, fiber.Map{"users": results})
}

// SendFriendRequest -> send friend request by phone
func SendFriendRequest(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.SendFriendRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[FRIEND_REQUEST_ERROR] BodyParser error: %v\n", err)
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}

	// normalize request phone to digits and find by phone_digits
	rq := strings.TrimSpace(req.Phone)
	if rq == "" {
		fmt.Printf("[FRIEND_REQUEST_ERROR] Empty phone field\n")
		return utils.Fail(c, fiber.StatusBadRequest, "Phone number is required")
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, rq)
	if digits == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Phone number is invalid")
	}

	var friend models.User
	if err := database.DB.Where("phone_digits = ?", digits).First(&friend).Error; err != nil {
		// fallback to exact phone comparison
		if err := database.DB.Where("phone = ?", req.Phone).First(&friend).Error; err != nil {
			fmt.Printf("[FRIEND_REQUEST_ERROR] User not found by phone: %s (digits: %s)\n", req.Phone, digits)
			return utils.Fail(c, fiber.StatusNotFound, "User not found")
		}
	}
	phoneStr := ""
	if friend.Phone != nil {
		phoneStr = *friend.Phone
	}
	fmt.Printf("[FRIEND_REQUEST] Found user: ID=%d, Name=%s, Phone=%s\n", friend.ID, friend.Nama, phoneStr)

	if friend.ID == user.ID {
		fmt.Printf("[FRIEND_REQUEST_ERROR] Cannot send request to self (UserID=%d)\n", user.ID)
		return utils.Fail(c, fiber.StatusBadRequest, "Cannot send friend request to yourself")
	}

	// Check if already friends or pending
	var existing models.Friendship
	err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friend.ID, friend.ID, user.ID).First(&existing).Error
	if err == nil {
		fmt.Printf("[FRIEND_REQUEST_ERROR] Friendship exists with status: %s\n", existing.Status)
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
		fmt.Printf("[FRIEND_REQUEST_ERROR] Failed to create friendship: %v\n", err)
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to send friend request")
	}
	fmt.Printf("[FRIEND_REQUEST] Success: UserID=%d -> FriendID=%d\n", user.ID, friend.ID)

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

// AcceptFriendRequestByPhone -> accept friend request using phone number in body
func AcceptFriendRequestByPhone(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.SendFriendRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[ACCEPT_ERROR] BodyParser error: %v\n", err)
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}

	// normalize phone to digits
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Phone number is required")
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Find friend by phone
	var friend models.User
	if err := database.DB.Where("phone_digits = ?", digits).First(&friend).Error; err != nil {
		if err := database.DB.Where("phone = ?", phone).First(&friend).Error; err != nil {
			return utils.Fail(c, fiber.StatusNotFound, "User not found")
		}
	}

	// Find pending friendship (request from friend to user)
	var friendship models.Friendship
	if err := database.DB.Where("(user_id = ? AND friend_id = ?) AND status = 'pending'",
		friend.ID, user.ID).First(&friendship).Error; err != nil {
		fmt.Printf("[ACCEPT_ERROR] No pending request from this user: %v\n", err)
		return utils.Fail(c, fiber.StatusNotFound, "No pending friend request from this user")
	}

	friendship.Status = "accepted"
	if err := database.DB.Save(&friendship).Error; err != nil {
		fmt.Printf("[ACCEPT_ERROR] Failed to save: %v\n", err)
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to accept friend request")
	}

	fmt.Printf("[ACCEPT] Success: UserID=%d accepted request from UserID=%d\n", user.ID, friend.ID)
	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Friend request accepted"})
}

// RejectFriendRequestByPhone -> reject friend request using phone number in body
func RejectFriendRequestByPhone(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var req dto.SendFriendRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[REJECT_ERROR] BodyParser error: %v\n", err)
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid JSON body")
	}

	// normalize phone to digits
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "Phone number is required")
	}

	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Find friend by phone
	var friend models.User
	if err := database.DB.Where("phone_digits = ?", digits).First(&friend).Error; err != nil {
		if err := database.DB.Where("phone = ?", phone).First(&friend).Error; err != nil {
			return utils.Fail(c, fiber.StatusNotFound, "User not found")
		}
	}

	// Find pending friendship (request from friend to user)
	var friendship models.Friendship
	if err := database.DB.Where("(user_id = ? AND friend_id = ?) AND status = 'pending'",
		friend.ID, user.ID).First(&friendship).Error; err != nil {
		fmt.Printf("[REJECT_ERROR] No pending request from this user: %v\n", err)
		return utils.Fail(c, fiber.StatusNotFound, "No pending friend request from this user")
	}

	if err := database.DB.Delete(&friendship).Error; err != nil {
		fmt.Printf("[REJECT_ERROR] Failed to delete: %v\n", err)
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to reject friend request")
	}

	fmt.Printf("[REJECT] Success: UserID=%d rejected request from UserID=%d\n", user.ID, friend.ID)
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

		// Check debt status: is current user in debt to this friend?
		isDebt := f.DebtUserID != nil && *f.DebtUserID == user.ID

		friends = append(friends, fiber.Map{
			"id":      friend.ID,
			"nama":    friend.Nama,
			"email":   friend.Email,
			"phone":   friend.Phone,
			"is_debt": isDebt,
			"status":  f.Status,
		})
	}

	// Ensure friends is never null, return empty array if no friends
	if friends == nil {
		friends = []fiber.Map{}
	}

	fmt.Printf("[FRIENDS_DEBUG] UserID=%d has %d accepted friends\n", user.ID, len(friends))
	return utils.Ok(c, fiber.StatusOK, fiber.Map{"friends": friends})
}

// ListPendingRequests -> list pending friend requests (sent and received)
func ListPendingRequests(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	fmt.Printf("[PENDING_DEBUG] UserID=%d fetching pending requests\n", user.ID)

	var friendships []models.Friendship
	if err := database.DB.Where("(user_id = ? OR friend_id = ?) AND status = 'pending'",
		user.ID, user.ID).Find(&friendships).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to fetch pending requests")
	}

	fmt.Printf("[PENDING_DEBUG] Found %d pending friendships\n", len(friendships))

	var sent []fiber.Map
	var received []fiber.Map
	for _, f := range friendships {
		fmt.Printf("[PENDING_DEBUG] Friendship ID=%d: UserID=%d, FriendID=%d, Status=%s\n", f.ID, f.UserID, f.FriendID, f.Status)

		if f.UserID == user.ID {
			// sent request
			fmt.Printf("[PENDING_DEBUG] This is a SENT request (I am UserID=%d)\n", user.ID)
			var friend models.User
			if err := database.DB.First(&friend, f.FriendID).Error; err != nil {
				fmt.Printf("[PENDING_DEBUG] Error fetching friend %d: %v\n", f.FriendID, err)
				continue
			}
			phoneStr := ""
			if friend.Phone != nil {
				phoneStr = *friend.Phone
			}
			fmt.Printf("[PENDING_DEBUG] Added SENT: friend_id=%d, name=%s, phone=%s\n", friend.ID, friend.Nama, phoneStr)
			sent = append(sent, fiber.Map{
				"id":           f.ID,
				"friend_id":    friend.ID,
				"friend_nama":  friend.Nama,
				"friend_phone": phoneStr,
			})
		} else {
			// received request
			fmt.Printf("[PENDING_DEBUG] This is a RECEIVED request (UserID=%d sent to me)\n", f.UserID)
			var friend models.User
			if err := database.DB.First(&friend, f.UserID).Error; err != nil {
				fmt.Printf("[PENDING_DEBUG] Error fetching friend %d: %v\n", f.UserID, err)
				continue
			}
			phoneStr := ""
			if friend.Phone != nil {
				phoneStr = *friend.Phone
			}
			fmt.Printf("[PENDING_DEBUG] Added RECEIVED: friend_id=%d, name=%s, phone=%s\n", friend.ID, friend.Nama, phoneStr)
			received = append(received, fiber.Map{
				"id":           f.ID,
				"friend_id":    friend.ID,
				"friend_nama":  friend.Nama,
				"friend_phone": phoneStr,
			})
		}
	}

	fmt.Printf("[PENDING_DEBUG] Final result: %d sent, %d received\n", len(sent), len(received))
	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"sent":     sent,
		"received": received,
	})
}

// DeleteFriend -> remove a friendship (soft delete)
func DeleteFriend(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	friendIDStr := c.Params("id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid friend ID")
	}

	// Find the friendship record
	var friendship models.Friendship
	if err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friendID, friendID, user.ID).First(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Friendship not found")
	}

	// Soft delete the friendship
	if err := database.DB.Delete(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to delete friendship")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{"message": "Friend removed"})
}

// ToggleDebt -> toggle debt status for a friend
func ToggleDebt(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	friendIDStr := c.Params("id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Invalid friend ID")
	}

	// Find the friendship record
	var friendship models.Friendship
	if err := database.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		user.ID, friendID, friendID, user.ID).First(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Friendship not found")
	}

	// Toggle debt: if current user is in debt, remove it; otherwise mark current user as in debt
	if friendship.DebtUserID != nil && *friendship.DebtUserID == user.ID {
		friendship.DebtUserID = nil
	} else {
		uid := user.ID
		friendship.DebtUserID = &uid
	}

	if err := database.DB.Save(&friendship).Error; err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "Failed to update debt status")
	}

	return utils.Ok(c, fiber.StatusOK, fiber.Map{
		"message":      "Debt status toggled",
		"debt_user_id": friendship.DebtUserID,
	})
}
