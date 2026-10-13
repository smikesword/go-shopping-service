package main

import (
	// "io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	// "encoding/json"

	gomock "go.uber.org/mock/gomock"
)



func TestHandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockRepositoryInterface(ctrl)
	db = mock

	mock.EXPECT().GetUser(LoginRequest{Username: "admin", Password: "admin"}).Return(
		&UserDTO{
			Role:     RoleAdmin,
			Username: "admin",
		},
		nil,
	)

	mock.EXPECT().AddSession("admin").Return(
		&Session{
			"test-token",
			time.Now().Add(time.Hour),
			"admin",
		},
		nil,
	)

	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"admin","password":"admin"}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handleLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `{"token":"test-token"}`) {
		t.Errorf("Expected body %s, got %s", `{"token":"test-token"}`, rec.Body.String())
	}

}

func TestLoginAPI(t *testing.T){
	var err error
	db, err = NewRepository("file::memory:?cache=shared")

	if err != nil{
		t.Errorf("NewRepository() error = %v", err)
	}

	db.Init()

	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"admin","password":"admin"}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handleLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, rec.Code)
	}
}

//E2E need to run in container first
// func TestLoginAndList(t *testing.T) {
// 	baseURL := "http://localhost:8888"
// 	reqBody := strings.NewReader(`{"username":"admin","password":"admin"}`)
// 	resp, err := makeRequest("POST", baseURL+"/login", "", reqBody)

// 	if err != nil {
// 		t.Fatalf("Failed to make login request: %v", err)
// 	}

// 	defer resp.Body.Close()

// 	var response map[string]string
// 	err = json.NewDecoder(resp.Body).Decode(&response)
// 	if err != nil {
// 		t.Fatalf("Failed to decode login response: %v", err)
// 	}

// 	token, ok := response["token"]
// 	if !ok {
// 		t.Fatalf("No token returned in login response")
// 	}

// 	resp1, err := makeRequest("POST", baseURL+"/v1/lists", token, strings.NewReader(`{"name":"Groceries","items":["Milk","Eggs","Bread"]}`))

// 	if err != nil {
// 		t.Fatalf("Failed to make create list request: %v", err)
// 	}

// 	if resp1.StatusCode != http.StatusCreated {
// 		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp1.StatusCode)
// 	}

// 	resp2, err := makeRequest("GET", baseURL+"/v1/lists", token, nil)

// 	if err != nil {
// 		t.Fatalf("Failed to make get lists request: %v", err)
// 	}

// 	if resp2.StatusCode != http.StatusOK {
// 		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp2.StatusCode)
// 	}

// 	defer resp2.Body.Close()

// 	var listsGot []ShoppingList

// 	err = json.NewDecoder(resp2.Body).Decode(&listsGot)

// 	if err != nil {
// 		t.Fatalf("Failed to decode get lists response: %v", err)
// 	}
	
// 	if err != nil {
// 		t.Fatalf("Failed to read get lists response: %v", err)
// 	}

// 	if len(listsGot) == 0 || listsGot[0].Name != "Groceries" {
// 		t.Fatalf("Expected list not found in response")
// 	}

// }

// func makeRequest(method, url, token string, body io.Reader) (*http.Response, error) {
// 	req, err := http.NewRequest(method, url, body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	if token != "" {
// 		req.Header.Set("Authorization", "Bearer "+token)
// 	}

// 	return http.DefaultClient.Do(req)
// }