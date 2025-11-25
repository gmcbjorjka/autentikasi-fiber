package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Quick test untuk SendFriendRequest endpoint
func testSendFriendRequest() {
	// First, login ke backend untuk dapatkan token
	loginResp, err := http.Post(
		"http://localhost:3000/api/v1/auth/login",
		"application/json",
		bytes.NewBufferString(`{"email":"teguh@example.com","password":"password123"}`),
	)
	if err != nil {
		fmt.Printf("Login error: %v\n", err)
		return
	}
	defer loginResp.Body.Close()

	var loginBody map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginBody)
	token := loginBody["data"].(map[string]interface{})["token"].(string)
	fmt.Printf("Token: %s\n", token)

	// Test 1: Send friend request by phone
	fmt.Println("\n=== Test 1: Send friend request by phone ===")
	client := &http.Client{}

	payload := map[string]string{
		"phone": "085972745905",
	}
	jsonPayload, _ := json.Marshal(payload)
	fmt.Printf("Request body: %s\n", string(jsonPayload))

	req, _ := http.NewRequest("POST", "http://localhost:3000/api/v1/friends/request", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", string(body))
}

/*
func main() {
	testSendFriendRequest()
}
*/
