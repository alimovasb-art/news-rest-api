package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {

	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(newUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Please fill all fields", err.Error())
		return
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to encrypt password", err.Error())
		return
	}

	query := `INSERT INTO users (first_name, last_name, email, password) VALUES($1, $2, $3, $4) RETURNING id`

	var newID int
	err = storage.DB.QueryRow(context.Background(), query,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		string(hashedBytes),
	).Scan(&newID)
	if err != nil {
		utils.SendError(w, http.StatusConflict, "User with this email already exist (or other error in db)", err.Error())
		return
	}

	response := map[string]int{
		"user_id": newID,
	}

	utils.SendSuccess(w, http.StatusCreated, "User created successfully", response)
}

var jwtSecretKey = []byte("hello-everyone")

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var userRequest models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	var foundUser models.User
	query := `
	SELECT id, first_name, last_name, email, password 
	FROM users 
	WHERE email =$1 AND deleted_at IS NULL`
	err = storage.DB.QueryRow(context.Background(), query, userRequest.Email).Scan(
		&foundUser.ID,
		&foundUser.FirstName,
		&foundUser.LastName,
		&foundUser.Email,
		&foundUser.Password,
	)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid email or password", err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userRequest.Password))
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	claims := jwt.MapClaims{
		"author_id":  foundUser.ID,
		"first_name": foundUser.FirstName,
		"last_name":  foundUser.LastName,
		"email":      foundUser.Email,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	response := map[string]string{
		"token": tokenString,
	}

	utils.SendSuccess(w, http.StatusOK, "Login successful!", response)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {

	page, limit := utils.GetPaginationParams(r)
	offset := (page - 1) * limit

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT $1 OFFSET $2
	`
	rows, err := storage.DB.Query(context.Background(), query, limit, offset)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}
	defer rows.Close()

	var activeUsers []models.Author

	for rows.Next() {
		var user models.Author

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
		)

		if err != nil {
			continue
		}

		activeUsers = append(activeUsers, user)
	}

	utils.SendSuccess(w, http.StatusOK, "Users loaded successfully", activeUsers)
}

func GetUsersByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE deleted_at IS NULL AND id = $1
	`
	var foundedUser models.Author

	err = storage.DB.QueryRow(context.Background(), query, id).Scan(
		&foundedUser.ID,
		&foundedUser.FirstName,
		&foundedUser.LastName,
		&foundedUser.Email,
	)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "There are no users with this id", err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User loaded successfully", foundedUser)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid id type", err.Error())
		return
	}

	var updateUser models.UpdateUser
	err = json.NewDecoder(r.Body).Decode(&updateUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	tokenAuthorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Failed to get ID from token", nil)
		return
	}

	if tokenAuthorID != id {
		utils.SendError(w, http.StatusForbidden, "You do not have permission to modify this user's data", nil)
		return
	}

	validate := validator.New()
	err = validate.Struct(updateUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Please fill all fields", err.Error())
		return
	}

	if updateUser.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(updateUser.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Failed to encrypt password", err.Error())
			return
		}
		updateUser.Password = string(hashedBytes)
	}

	query := `
		UPDATE users 
		SET 
    		first_name = COALESCE(NULLIF($1, ''), first_name),
    		last_name = COALESCE(NULLIF($2, ''), last_name),
    		email = COALESCE(NULLIF($3, ''), email),
    		password = COALESCE(NULLIF($4, ''), password),
    		updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, first_name, last_name, email
	`

	var updatedUser models.Author
	err = storage.DB.QueryRow(
		context.Background(),
		query,
		updateUser.FirstName,
		updateUser.LastName,
		updateUser.Email,
		updateUser.Password,
		id,
	).Scan(
		&updatedUser.ID,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.Email,
	)

	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to update user", err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User updated succesfully", updateUser)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID type", err.Error())
		return
	}

	tokenAuthorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusUnauthorized, "Failed to get ID from token", nil)
		return
	}

	if tokenAuthorID != id {
		utils.SendError(w, http.StatusForbidden, "You do not have permission to delete this user", nil)
		return
	}

	query := `
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL AND id = $1
	`

	_, err = storage.DB.Exec(context.Background(), query, id)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to delete user", err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User deleted successfully", nil)
}
