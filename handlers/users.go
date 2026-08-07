package handlers

import (
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *fiber.Ctx) error {

	var newUser models.User
	err := c.BodyParser(&newUser)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid JSON format", err.Error())
	}

	validate := validator.New()
	err = validate.Struct(newUser)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Please fill all fields", err.Error())

	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to encrypt password", err.Error())

	}

	query := `INSERT INTO users (first_name, last_name, email, password) VALUES($1, $2, $3, $4) RETURNING id`

	var newID int
	err = storage.DB.QueryRow(c.Context(), query,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		string(hashedBytes),
	).Scan(&newID)
	if err != nil {
		return utils.SendError(c, fiber.StatusConflict, "User with this email already exist (or other error in db)", err.Error())

	}

	response := map[string]int{
		"user_id": newID,
	}

	return utils.SendSuccess(c, fiber.StatusCreated, "User created successfully", response)
}

func LoginUser(c *fiber.Ctx) error {
	var userRequest models.LoginRequest
	err := c.BodyParser(&userRequest)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid JSON", err.Error())

	}

	var foundUser models.User
	query := `
	SELECT id, first_name, last_name, email, password 
	FROM users 
	WHERE email =$1 AND deleted_at IS NULL`
	err = storage.DB.QueryRow(c.Context(), query, userRequest.Email).Scan(
		&foundUser.ID,
		&foundUser.FirstName,
		&foundUser.LastName,
		&foundUser.Email,
		&foundUser.Password,
	)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid email or password", err.Error())

	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(userRequest.Password))
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid email or password", nil)

	}

	claims := jwt.MapClaims{
		"author_id": foundUser.ID,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())

	}

	response := map[string]string{
		"token": tokenString,
	}

	return utils.SendSuccess(c, fiber.StatusOK, "Login successful!", response)
}

func GetMe(c *fiber.Ctx) error {
	tokenAuthorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "Failed to get ID from context", nil)

	}

	query := `
		SELECT id, first_name, last_name, email FROM users WHERE id = $1 AND deleted_at IS NULL
	`

	var getMe models.Author
	err := storage.DB.QueryRow(c.Context(), query, tokenAuthorID).Scan(
		&getMe.ID,
		&getMe.FirstName,
		&getMe.LastName,
		&getMe.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "User not found", nil)
		} else {
			return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())
		}
	}
	return utils.SendSuccess(c, fiber.StatusOK, "Current user fetched successfully", getMe)
}

func GetUsers(c *fiber.Ctx) error {

	page, limit := utils.GetPaginationParams(c)
	offset := (page - 1) * limit

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT $1 OFFSET $2
	`
	rows, err := storage.DB.Query(c.Context(), query, limit, offset)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())

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

	return utils.SendSuccess(c, fiber.StatusOK, "Users loaded successfully", activeUsers)
}

func GetUsersByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())

	}

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE deleted_at IS NULL AND id = $1
	`
	var foundedUser models.Author

	err = storage.DB.QueryRow(c.Context(), query, id).Scan(
		&foundedUser.ID,
		&foundedUser.FirstName,
		&foundedUser.LastName,
		&foundedUser.Email,
	)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "There are no users with this id", err.Error())

	}

	return utils.SendSuccess(c, fiber.StatusOK, "User loaded successfully", foundedUser)
}

func UpdateUser(c *fiber.Ctx) error {

	var updateUser models.UpdateUser
	err := c.BodyParser(&updateUser)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid JSON format", err.Error())

	}

	tokenAuthorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "Failed to get ID from token", nil)

	}

	validate := validator.New()
	err = validate.Struct(updateUser)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Please fill all fields", err.Error())

	}

	if updateUser.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(updateUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to encrypt password", err.Error())

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
		c.Context(),
		query,
		updateUser.FirstName,
		updateUser.LastName,
		updateUser.Email,
		updateUser.Password,
		tokenAuthorID,
	).Scan(
		&updatedUser.ID,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.Email,
	)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())

	}

	return utils.SendSuccess(c, fiber.StatusOK, "User updated succesfully", updatedUser)
}

func DeleteUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid ID type", err.Error())

	}

	tokenAuthorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusUnauthorized, "Failed to get ID from token", nil)

	}

	if tokenAuthorID != id {
		return utils.SendError(c, fiber.StatusForbidden, "You do not have permission to delete this user", nil)

	}

	query := `
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL AND id = $1
	`

	_, err = storage.DB.Exec(c.Context(), query, id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())

	}

	return utils.SendSuccess(c, fiber.StatusOK, "User deleted successfully", nil)
}
