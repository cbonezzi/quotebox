package mysql

import (
	"database/sql"
	"strings"

	"cb.net/snippetbox/pkg/models"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// UserModel model
type UserModel struct {
	DB *sql.DB
}

// Get method to fetch details for a specific user based
// on their user ID.
func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}
	
	stmt := `SELECT id, name, email, created, active FROM users WHERE id = ?`
	
	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	return u, nil
}

// Insert method to add a new record to the users table.
func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
			VALUES(?, ?, ?, UTC_TIMESTAMP())`
	// Use the Exec() method to insert the user details and hashed password
	// into the users table. If this returns an error, we try to type assert
	// it to a *mysql.MySQLError object so we can check if the error number is
	// 1062 and, if it is, we also check whether or not the error relates to
	// our users_uc_email key by checking the contents of the message string.
	// If it does, we return an ErrDuplicateEmail error. Otherwise, we just
	// return the original error (or nil if everything worked).
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
	}
	return err
}

// Authenticate method to verify whether a user exists with
// the provided email address and password. This will return the relevant
// user ID if they do.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the id and hashed password associated with the given email. If no
	// matching email exists, or the user is not active, we return the
	// ErrInvalidCredentials error.
	var id int
	var hashedPassword []byte
	
	stmt := "SELECT id, hashed_password FROM users WHERE email = ? AND active = TRUE"
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// Otherwise, the password is correct. Return the user ID.
	return id, nil
}

func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	row := m.DB.QueryRow("SELECT hashed_password FROM users WHERE id = ?", id)
	err := row.Scan(&currentHashedPassword)
	if err != nil {
		return err
	}
	
	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return models.ErrInvalidCredentials
	} else if err != nil {
		return err
	}
	
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	
	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
		return err
}
