package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

func CreateNews(w http.ResponseWriter, r *http.Request) {
	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var newNews models.News
	err := json.NewDecoder(r.Body).Decode(&newNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	newNews.AuthorID = authorID

	validate := validator.New()
	err = validate.Struct(newNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	query := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE id = $1
	`

	var authorObject models.Author
	err = storage.DB.QueryRow(context.Background(), query, authorID).Scan(
		&authorObject.ID,
		&authorObject.FirstName,
		&authorObject.LastName,
		&authorObject.Email,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			utils.SendError(w, http.StatusNotFound, "Author not found in database", nil)
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Database error while fetching author", err.Error())
		}
		return
	}

	query = `
		INSERT INTO news (author_id, title, short_description, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, short_description, description, author_id, created_at
	`

	var userResponse models.NewsResponse
	err = storage.DB.QueryRow(
		context.Background(), query,
		newNews.AuthorID,
		newNews.Title,
		newNews.ShortDescription,
		newNews.Description,
	).Scan(
		&userResponse.ID,
		&userResponse.Title,
		&userResponse.ShortDescription,
		&userResponse.Description,
		&userResponse.AuthorID,
		&userResponse.CreatedAt,
	)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to create news in database", err.Error())
		return
	}

	userResponse.Author = &authorObject
	utils.SendSuccess(w, http.StatusCreated, "News created successfully", userResponse)
}

func GetNews(w http.ResponseWriter, r *http.Request) {
	authorIdStr := r.URL.Query().Get("author_id")
	title := r.URL.Query().Get("title")
	page, limit := utils.GetPaginationParams(r)
	offset := (page - 1) * limit

	query := `
		SELECT
			n.id,
			n.title,
			n.short_description,
			n.description,
			n.author_id,
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			n.views,
			n.created_at
		FROM news n
		JOIN users u ON n.author_id = u.id
		WHERE n.deleted_at IS NULL
	`

	var arguments []interface{}
	argumentsID := 1

	if authorIdStr != "" {
		authorID, err := strconv.Atoi(authorIdStr)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
			return
		}
		query += fmt.Sprintf(" AND n.author_id = $%d", argumentsID)
		arguments = append(arguments, authorID)
		argumentsID++
	}

	if title != "" {
		query += fmt.Sprintf(" AND n.title ILIKE $%d", argumentsID)
		arguments = append(arguments, "%"+title+"%")
		argumentsID++
	}

	query += fmt.Sprintf(" ORDER BY n.id DESC LIMIT $%d OFFSET $%d", argumentsID, argumentsID+1)
	arguments = append(arguments, limit, offset)

	rows, err := storage.DB.Query(context.Background(), query, arguments...)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
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
			&item.AuthorID,
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.Email,
			&item.Views,
			&item.CreatedAt,
		)
		if err != nil {
			continue
		}

		item.Author = &author

		newsList = append(newsList, item)
	}

	if newsList == nil {
		newsList = []models.NewsResponse{}
	}

	utils.SendSuccess(w, http.StatusOK, "News list retrieved successfully", newsList)

}

func GetNewsByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	query := `
		SELECT 
			n.id,
			n.title,
			n.short_description,
			n.description,
			n.views,
			n.author_id,
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			n.created_at
		FROM news n
		JOIN users u ON n.author_id = u.id 
		WHERE n.id = $1 AND n.deleted_at IS NULL
	`
	var responseNews models.NewsResponse
	var author models.Author

	err = storage.DB.QueryRow(context.Background(), query, id).Scan(
		&responseNews.ID,
		&responseNews.Title,
		&responseNews.ShortDescription,
		&responseNews.Description,
		&responseNews.Views,
		&responseNews.AuthorID,

		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.Email,

		&responseNews.CreatedAt,
	)
	responseNews.Author = &author

	if err != nil {
		if err == pgx.ErrNoRows {
			utils.SendError(w, http.StatusNotFound, "News not found", nil)
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		}
		return
	}

	storage.DB.Exec(context.Background(), "UPDATE news SET views = views + 1 WHERE id = $1", id)

	responseNews.Views++

	utils.SendSuccess(w, http.StatusOK, "News fetched successfully", responseNews)

}

func UpdateNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var updateNews models.News
	err = json.NewDecoder(r.Body).Decode(&updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	query := `
		UPDATE news n
		SET 
			title = $1,
			short_description = $2,
			description = $3,
			updated_at = CURRENT_TIMESTAMP
		FROM users u
		WHERE n.author_id = u.id 
		  AND n.id = $4 
		  AND n.author_id = $5 
		  AND n.deleted_at IS NULL
		RETURNING 
			n.id, n.title, n.short_description, n.description, n.views, n.author_id,
			u.id, u.first_name, u.last_name, u.email,
			n.created_at, n.updated_at
	`

	var responseNews models.NewsResponse
	var author models.Author
	err = storage.DB.QueryRow(context.Background(), query, updateNews.Title, updateNews.ShortDescription, updateNews.Description, id, authorID).Scan(
		&responseNews.ID,
		&responseNews.Title,
		&responseNews.ShortDescription,
		&responseNews.Description,
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
			utils.SendError(w, http.StatusNotFound, "News not found", nil)
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		}
		return
	}

	utils.SendSuccess(w, http.StatusOK, "News updated successfully", responseNews)
}

func PatchNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var updateNews models.UpdateNews
	err = json.NewDecoder(r.Body).Decode(&updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	query := `
		UPDATE news n
		SET
			title = COALESCE(NULLIF($1, ''), n.title),
			short_description = COALESCE(NULLIF($2, ''), n.short_description),
			description = COALESCE(NULLIF($3, ''), n.description),
			updated_at = CURRENT_TIMESTAMP
		FROM users u
		WHERE n.author_id = u.id
		AND n.id = $4
		AND n.author_id = $5
		AND n.deleted_at IS NULL
		RETURNING 
			n.id, n.title, n.short_description, n.description, n.views, n.author_id,
			u.id, u.first_name, u.last_name, u.email,
			n.created_at, n.updated_at	
	`

	var responseNews models.NewsResponse
	var author models.Author
	err = storage.DB.QueryRow(context.Background(), query, updateNews.Title, updateNews.ShortDescription, updateNews.Description, id, authorID).Scan(
		&responseNews.ID,
		&responseNews.Title,
		&responseNews.ShortDescription,
		&responseNews.Description,
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
			utils.SendError(w, http.StatusNotFound, "News not found", nil)
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		}
		return
	}

	responseNews.Author = &author

	utils.SendSuccess(w, http.StatusOK, "News updated successfully", responseNews)
}

func DeleteNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	autorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	query := `
		UPDATE news SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL
	`

	cmdTag, err := storage.DB.Exec(context.Background(), query, id, autorID)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}
	if cmdTag.RowsAffected() == 0 {
		utils.SendError(w, http.StatusNotFound, "News not found or you are not the author", nil)
		return
	}

	utils.SendSuccess(w, http.StatusOK, "News deleted successfully", nil)
}
