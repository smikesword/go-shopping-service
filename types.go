package main

import "time"

type Role string

// @Description Session represents a user session with a token, expiration time, and associated username.
type Session struct {
	Token string `json:"token" example:"abc123"`
	Expires time.Time `json:"expires" example:"2023-01-01T00:00:00Z"`
	Username string `json:"username" example:"john_doe"`
}

type User struct {
	ID int `json: "id"`
  	Role Role `json: "role"`
	Username string
	Password string
}

type UserDTO struct {
	Role Role `json: "role"`
	Username string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Description ShoppingList represents a shopping list with an ID, name, and items.
type ShoppingList struct {
	ID int `json:"id" example:"1"`
	Name string `json:"name" example:"Groceries"`
	Items []string `json:"items" example:"[\"Milk\",\"Eggs\",\"Bread\"]"`
}

// @Description ShoppingListDTO represents a shopping list without an ID, so clients can't see
type ShoppingListDTO struct {
	Name string `json:"name" example:"Groceries"`
	Items []string `json:"items" example:"[\"Milk\",\"Eggs\",\"Bread\"]"`
}

type ShoppingListPatch struct {
	Name *string `json:"name"`
	Items []string `json:"items"`
}

type ListPushAction struct {
	Item string `json:"item"`
}

