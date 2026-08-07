package handlers

import (
	"fmt"
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func CreateNews(c *fiber.Ctx) error {
	authorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get user from context", nil)
	}

	title := c.FormValue("title")
	shortDesc := c.FormValue("short_description")
	description := c.FormValue("description")

	newNews := models.News{
		Title:            title,
		ShortDescription: shortDesc,
		Description:      description,
	}

	validate := validator.New()
	err := validate.Struct(newNews)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
	}

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil {
		_ = os.MkdirAll("./uploads", 0755)

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		filePath := filepath.Join("./uploads", fileName)

		err := c.SaveFile(file, filePath)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save file on disk", err.Error())
		}

		imageURL = "/uploads/" + fileName
	}

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE id = $1
	`

	var authorObject models.Author
	err = storage.DB.QueryRow(c.Context(), query, authorID).Scan(
		&authorObject.ID,
		&authorObject.FirstName,
		&authorObject.LastName,
		&authorObject.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "Author not found in database", nil)
		} else {
			return utils.SendError(c, fiber.StatusInternalServerError, "Database error while fetching author", err.Error())
		}
	}

	insertQuery := `
		INSERT INTO news (author_id, title, short_description, description, image)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, short_description, description, views, author_id, COALESCE(image, ''), created_at
	`
	var newsResponse models.NewsResponse
	err = storage.DB.QueryRow(
		c.Context(), insertQuery,
		authorID, title, shortDesc, description, imageURL,
	).Scan(
		&newsResponse.ID,
		&newsResponse.Title,
		&newsResponse.ShortDescription,
		&newsResponse.Description,
		&newsResponse.Views,
		&newsResponse.AuthorID,
		&newsResponse.ImageURL,
		&newsResponse.CreatedAt,
	)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create news in database", err.Error())
	}

	newsResponse.Author = &authorObject
	return utils.SendSuccess(c, fiber.StatusCreated, "News created successfully", newsResponse)
}

func GetNews(c *fiber.Ctx) error {
	page, limit := utils.GetPaginationParams(c)
	offset := (page - 1) * limit

	search := c.Query("search")
	authorIDStr := c.Query("author_id")

	query := `
		SELECT 
			n.id, n.title, n.short_description, n.description, COALESCE(n.image, '') AS image, n.views, n.author_id,
			u.id, u.first_name, u.last_name, u.email,
			n.created_at, n.updated_at
		FROM news n
		JOIN users u ON n.author_id = u.id
		WHERE n.deleted_at IS NULL
		  AND ($1 = '' OR n.title ILIKE '%' || $1 || '%' OR n.description ILIKE '%' || $1 || '%')
		  AND ($2 = 0 OR n.author_id = $2)
		ORDER BY n.created_at DESC
		LIMIT $3 OFFSET $4
	`
	var filterAuthorID int
	if authorIDStr != "" {
		filterAuthorID, _ = strconv.Atoi(authorIDStr)
	}

	rows, err := storage.DB.Query(c.Context(), query, search, filterAuthorID, limit, offset)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())
	}
	defer rows.Close()

	var newsList []models.NewsResponse
	for rows.Next() {
		var item models.NewsResponse
		var author models.Author
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.ShortDescription,
			&item.Description,
			&item.ImageURL,
			&item.Views,
			&item.AuthorID,
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.Email,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			continue
		}
		item.Author = &author
		newsList = append(newsList, item)
	}
	return utils.SendSuccess(c, fiber.StatusOK, "News list loaded successfully", newsList)
}

func GetNewsByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	query := `
		SELECT 
			n.id,
			n.title,
			n.short_description,
			n.description,
			COALESCE(n.image, '') AS image,
			n.views,
			n.author_id,
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			n.created_at,
			n.updated_at
		FROM news n
		JOIN users u ON n.author_id = u.id 
		WHERE n.id = $1 AND n.deleted_at IS NULL
	`
	var responseNews models.NewsResponse
	var author models.Author

	err = storage.DB.QueryRow(c.Context(), query, id).Scan(
		&responseNews.ID,
		&responseNews.Title,
		&responseNews.ShortDescription,
		&responseNews.Description,
		&responseNews.ImageURL,
		&responseNews.Views,
		&responseNews.AuthorID,

		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.Email,

		&responseNews.CreatedAt,
		&responseNews.UpdatedAt,
	)
	responseNews.Author = &author

	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "News not found", nil)
		} else {
			return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())
		}
	}

	_, _ = storage.DB.Exec(c.Context(), "UPDATE news SET views = views + 1 WHERE id = $1", id)
	responseNews.Views++

	return utils.SendSuccess(c, fiber.StatusOK, "News fetched successfully", responseNews)
}

func PatchNews(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	authorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get user from context", nil)
	}

	title := c.FormValue("title")
	shortDescription := c.FormValue("short_description")
	description := c.FormValue("description")

	updateNews := models.NewsResponse{
		Title:            title,
		ShortDescription: shortDescription,
		Description:      description,
	}

	validate := validator.New()
	err = validate.Struct(updateNews)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
	}

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil {
		_ = os.MkdirAll("./uploads", 0755)

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		filePath := filepath.Join("./uploads/", fileName)

		err := c.SaveFile(file, filePath)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to save file on disk", err.Error())
		}
		imageURL = "/uploads/" + fileName
	}

	query := `
		UPDATE news n
		SET
			title = COALESCE(NULLIF($1, ''), n.title),
			short_description = COALESCE(NULLIF($2, ''), n.short_description),
			description = COALESCE(NULLIF($3, ''), n.description),
			image = COALESCE(NULLIF($4, ''), n.image),
			updated_at = CURRENT_TIMESTAMP
		FROM users u
		WHERE n.author_id = u.id
		AND n.id = $5
		AND n.author_id = $6
		AND n.deleted_at IS NULL
		RETURNING 
			n.id, n.title, n.short_description, n.description, COALESCE(n.image, '') AS image, n.views, n.author_id,
			u.id, u.first_name, u.last_name, u.email,
			n.created_at, n.updated_at	
	`

	var responseNews models.NewsResponse
	var author models.Author
	err = storage.DB.QueryRow(c.Context(), query, updateNews.Title, updateNews.ShortDescription, updateNews.Description, imageURL, id, authorID).Scan(
		&responseNews.ID,
		&responseNews.Title,
		&responseNews.ShortDescription,
		&responseNews.Description,
		&responseNews.ImageURL,
		&responseNews.Views,
		&responseNews.AuthorID,

		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.Email,

		&responseNews.CreatedAt,
		&responseNews.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "News not found or you are not the author", nil)
		} else {
			return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())
		}
	}

	responseNews.Author = &author
	return utils.SendSuccess(c, fiber.StatusOK, "News updated successfully", responseNews)
}

func DeleteNews(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	autorID, ok := c.Locals("author_id").(int)
	if !ok {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get user from context", nil)
	}

	query := `
		UPDATE news SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL
	`

	cmdTag, err := storage.DB.Exec(c.Context(), query, id, autorID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Database error", err.Error())
	}
	if cmdTag.RowsAffected() == 0 {
		return utils.SendError(c, fiber.StatusNotFound, "News not found or you are not the author", nil)
	}

	return utils.SendSuccess(c, fiber.StatusOK, "News deleted successfully", nil)
}
