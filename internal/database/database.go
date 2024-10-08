package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// GetAllUsers returns a list of all users in the database.
    GetAllUsers() ([]User, error)
    GetUser(id int) (User, error)
	DeleteUser(id int) (string, error)
	UpdateUser(id int, newName string, newEmail string) (string, error)
	CreateUser(username, email string) (User, error)
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// GetAllUsers returns a list of all users in the database.

func(s *service) GetAllUsers() ([]User, error){
	rows, err := s.db.Query("SELECT id, username, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func(s *service) GetUser(id int) (User, error){

	var user User

	err := s.db.QueryRow("SELECT id, username, email FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.Email)

	if err == sql.ErrNoRows {
		return User{}, fmt.Errorf("user not found")
	} else if err != nil{
		return User{}, err
	}
	return user, nil
}

func(s *service) DeleteUser(id int) (string, error) {
	exists, err := s.userExists(id)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("user does not exist")
	}

	query := "DELETE FROM users WHERE id = $1"
	_, err = s.db.Exec(query, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User with ID %d successfully deleted", id), nil
}

func (s *service) CreateUser(username, email string) (User, error) {
	// Verificar se o username já existe
	exists, err := s.usernameExists(username)
	if err != nil {
		return User{}, err
	}
	if exists {
		return User{}, fmt.Errorf("user already exists")
	}

	// Inserir o novo usuário
	var newUser User
	query := "INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id, username, email"
	err = s.db.QueryRow(query, username, email).Scan(&newUser.ID, &newUser.Username, &newUser.Email)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}



func (s *service) UpdateUser(id int, newName string, newEmail string) (string, error) {
	// Verificar se o usuário existe
	exists, err := s.userExists(id)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("user does not exist")
	}

	// Query para atualizar o usuário
	query := "UPDATE users SET username = $1, email = $2 WHERE id = $3"
	_, err = s.db.Exec(query, newName, newEmail, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User with ID %d successfully updated", id), nil
}

func (s *service) userExists(id int) (bool, error) {
	var userID int
	err := s.db.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&userID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (s *service) usernameExists(username string) (bool, error) {
	var id int
	err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
