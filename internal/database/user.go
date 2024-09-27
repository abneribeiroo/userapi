package database

// User represents a user in the database.
type User struct {
	ID        int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

//Create a new user
//CreateUser creates a new user in the database.

