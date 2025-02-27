package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Model represents an LLM model with its properties
type Model struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Modality string `json:"modality"`
}

// Conversation represents a chat conversation containing multiple messages
type Conversation struct {
	ID       string                 `json:"id"`
	Time     int64                  `json:"time"`
	Models   []Model                `json:"models"`
	Messages []string               `json:"messages"` // Message IDs
	Versions map[string]interface{} `json:"versions"`
	Tools    []string               `json:"tools"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             string `json:"id"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	Unclosed       bool   `json:"unclosed"`
	Generated      bool   `json:"generated"`
	Model          Model  `json:"model"`
	PendingContent string `json:"pendingContent"`
	Editing        bool   `json:"editing"`
	Submitted      bool   `json:"submitted"`
}

// UserData contains all conversations and messages for a specific user
type UserData struct {
	Conversations map[string]Conversation `json:"conversations"`
	Messages      map[string]Message      `json:"messages"`
}

// In-memory storage with mutex for concurrency protection
var (
	storage = make(map[string]UserData) // Map from user token to their data
	mutex   = &sync.RWMutex{}           // Lock for concurrent access to storage
)

// Request and response types
type CheckClientMissingRequest struct {
	Token           string   `json:"token"`
	ConversationIds []string `json:"conversationIds"`
	MessageIds      []string `json:"messageIds"`
}

type CheckClientMissingResponse struct {
	MissingConversationIds []string `json:"missingConversationIds"`
	MissingMessageIds      []string `json:"missingMessageIds"`
}

type GetItemsRequest struct {
	Token           string   `json:"token"`
	ConversationIds []string `json:"conversationIds"`
	MessageIds      []string `json:"messageIds"`
}

type GetItemsResponse struct {
	Conversations map[string]Conversation `json:"conversations"`
	Messages      map[string]Message      `json:"messages"`
}

type CheckServerMissingRequest struct {
	Token           string   `json:"token"`
	ConversationIds []string `json:"conversationIds"`
	MessageIds      []string `json:"messageIds"`
}

type CheckServerMissingResponse struct {
	MissingConversationIds []string `json:"missingConversationIds"`
	MissingMessageIds      []string `json:"missingMessageIds"`
}

type SendItemsRequest struct {
	Token         string                  `json:"token"`
	Conversations map[string]Conversation `json:"conversations"`
	Messages      map[string]Message      `json:"messages"`
}

type SendItemsResponse struct {
	Success bool `json:"success"`
}

type SendSingleItemRequest struct {
	Token        string        `json:"token"`
	Conversation *Conversation `json:"conversation,omitempty"`
	Message      *Message      `json:"message,omitempty"`
}

type SendSingleItemResponse struct {
	Success bool `json:"success"`
}

func main() {
	// Load storage from file if exists
	loadStorageFromFile()

	// Save storage to file periodically
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			saveStorageToFile()
		}
	}()

	// Set up API endpoint handlers
	http.HandleFunc("/api/sync/check-client-missing", handleCheckClientMissing)
	http.HandleFunc("/api/sync/get-items", handleGetItems)
	http.HandleFunc("/api/sync/check-server-missing", handleCheckServerMissing)
	http.HandleFunc("/api/sync/send-items", handleSendItems)
	http.HandleFunc("/api/sync/send-single-item", handleSendSingleItem)

	// Add a simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create CORS middleware handler
	corsHandler := corsMiddleware(http.DefaultServeMux)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}

// corsMiddleware adds CORS headers to each response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// handleCheckClientMissing finds what items the client is missing compared to the server
func handleCheckClientMissing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckClientMissingRequest
	if err := parseBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Get user data or create new if not exists
	mutex.RLock()
	userData, exists := storage[req.Token]
	mutex.RUnlock()

	if !exists {
		// No data for this token
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CheckClientMissingResponse{
			MissingConversationIds: []string{},
			MissingMessageIds:      []string{},
		})
		return
	}

	// Find missing conversations
	missingConversations := []string{}
	for convId := range userData.Conversations {
		found := false
		for _, clientConvId := range req.ConversationIds {
			if convId == clientConvId {
				found = true
				break
			}
		}
		if !found {
			missingConversations = append(missingConversations, convId)
		}
	}

	// Find missing messages
	missingMessages := []string{}
	for msgId := range userData.Messages {
		found := false
		for _, clientMsgId := range req.MessageIds {
			if msgId == clientMsgId {
				found = true
				break
			}
		}
		if !found {
			missingMessages = append(missingMessages, msgId)
		}
	}

	resp := CheckClientMissingResponse{
		MissingConversationIds: missingConversations,
		MissingMessageIds:      missingMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleGetItems returns the requested conversations and messages
func handleGetItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GetItemsRequest
	if err := parseBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	userData, exists := storage[req.Token]
	mutex.RUnlock()

	if !exists {
		// Return empty data if user not found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GetItemsResponse{
			Conversations: make(map[string]Conversation),
			Messages:      make(map[string]Message),
		})
		return
	}

	// Collect requested conversations
	conversations := make(map[string]Conversation)
	for _, convId := range req.ConversationIds {
		if conv, ok := userData.Conversations[convId]; ok {
			conversations[convId] = conv
		}
	}

	// Collect requested messages
	messages := make(map[string]Message)
	for _, msgId := range req.MessageIds {
		if msg, ok := userData.Messages[msgId]; ok {
			messages[msgId] = msg
		}
	}

	resp := GetItemsResponse{
		Conversations: conversations,
		Messages:      messages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleCheckServerMissing finds what items the server is missing compared to the client
func handleCheckServerMissing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckServerMissingRequest
	if err := parseBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	mutex.RLock()
	userData, exists := storage[req.Token]
	mutex.RUnlock()

	if !exists {
		// All client items are missing from server
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CheckServerMissingResponse{
			MissingConversationIds: req.ConversationIds,
			MissingMessageIds:      req.MessageIds,
		})
		return
	}

	// Find which conversations the server is missing
	missingConversations := []string{}
	for _, clientConvId := range req.ConversationIds {
		if _, ok := userData.Conversations[clientConvId]; !ok {
			missingConversations = append(missingConversations, clientConvId)
		}
	}

	// Find which messages the server is missing
	missingMessages := []string{}
	for _, clientMsgId := range req.MessageIds {
		if _, ok := userData.Messages[clientMsgId]; !ok {
			missingMessages = append(missingMessages, clientMsgId)
		}
	}

	resp := CheckServerMissingResponse{
		MissingConversationIds: missingConversations,
		MissingMessageIds:      missingMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleSendItems receives multiple items from the client and stores them
func handleSendItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendItemsRequest
	if err := parseBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Get or create user data
	userData, exists := storage[req.Token]
	if !exists {
		userData = UserData{
			Conversations: make(map[string]Conversation),
			Messages:      make(map[string]Message),
		}
	}

	// Add conversations
	for id, conv := range req.Conversations {
		userData.Conversations[id] = conv
	}

	// Add messages
	for id, msg := range req.Messages {
		userData.Messages[id] = msg
	}

	// Update storage
	storage[req.Token] = userData

	// Save to file after updates
	go saveStorageToFile()

	resp := SendItemsResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleSendSingleItem receives a single item from the client and stores it
func handleSendSingleItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendSingleItemRequest
	if err := parseBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if req.Conversation == nil && req.Message == nil {
		http.Error(w, "Either conversation or message must be provided", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Get or create user data
	userData, exists := storage[req.Token]
	if !exists {
		userData = UserData{
			Conversations: make(map[string]Conversation),
			Messages:      make(map[string]Message),
		}
	}

	// Add conversation if provided
	if req.Conversation != nil {
		conv := *req.Conversation
		if conv.ID == "" {
			http.Error(w, "Conversation ID is required", http.StatusBadRequest)
			return
		}
		userData.Conversations[conv.ID] = conv
	}

	// Add message if provided
	if req.Message != nil {
		msg := *req.Message
		if msg.ID == "" {
			http.Error(w, "Message ID is required", http.StatusBadRequest)
			return
		}
		userData.Messages[msg.ID] = msg
	}

	// Update storage
	storage[req.Token] = userData

	// Save to file after updates
	go saveStorageToFile()

	resp := SendSingleItemResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// parseBody extracts and parses JSON from the request body
func parseBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	return nil
}

// saveStorageToFile persists the in-memory storage to a file
func saveStorageToFile() {
	mutex.RLock()
	defer mutex.RUnlock()

	data, err := json.Marshal(storage)
	if err != nil {
		log.Printf("Failed to marshal storage data: %v", err)
		return
	}

	// Write to a temporary file first for atomic replacement
	tempFile := "storage.json.temp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		log.Printf("Failed to write temporary storage file: %v", err)
		return
	}

	// Rename to do an atomic replace (safer)
	if err := os.Rename(tempFile, "storage.json"); err != nil {
		log.Printf("Failed to rename storage file: %v", err)
	}
}

// loadStorageFromFile loads the storage data from file into memory
func loadStorageFromFile() {
	data, err := os.ReadFile("storage.json")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read storage file: %v", err)
		}
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if err := json.Unmarshal(data, &storage); err != nil {
		log.Printf("Failed to unmarshal storage data: %v", err)
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
