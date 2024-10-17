package controllers

import (
	"InternshipProject/models"
	"InternshipProject/utills"
	"fmt"
	"net/http"
	"time"

	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

var (
	vendor_columns = []string{
		"id",
		"name",
		"description",
		"img",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", utills.Domain),
	}
)

func IndexVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendors []models.Vendor
	err := db.Select(&vendors, "SELECT * FROM vendors")
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error fetching vendors")
		return
	}

	utills.SendJSONRespone(w, http.StatusOK, vendors)
}

func ShowVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendor
	id := r.PathValue("id")
	query, args, err := QB.Select(strings.Join(vendor_columns, ", ")).
		From("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&vendor, query, args...); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utills.SendJSONRespone(w, http.StatusOK, vendor)
}

func StoreVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendor
	vendor.ID = uuid.New()
	vendor.Name = r.FormValue("name")
	vendor.Description = r.FormValue("description")

	// Handle image upload
	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		utills.HandleError(w, http.StatusBadRequest, "Invalid file")
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utills.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			utills.HandleError(w, http.StatusInternalServerError, "Error saving image")
			return
		}
		vendor.Img = &imageName
	}

	vendor.CreatedAt = time.Now()
	vendor.UpdatedAt = time.Now()

	// Insert vendor into the database
	query, args, err := QB.Insert("vendors").
		Columns("id", "name", "description", "img", "created_at", "updated_at").
		Values(vendor.ID, vendor.Name, vendor.Description, vendor.Img, vendor.CreatedAt, vendor.UpdatedAt).
		ToSql()

	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error generating query")
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error storing vendor")
		return
	}

	utills.SendJSONRespone(w, http.StatusCreated, vendor)
}

func UpdateVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.Vendor
	id := r.PathValue("id")

	query, args, err := QB.Select(vendor_columns...).
		From("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := db.Get(&vendor, query, args...); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if r.FormValue("name") != "" {
		vendor.Name = r.FormValue("name")
	}
	if r.FormValue("description") != "" {
		vendor.Description = r.FormValue("description")
	}

	// Handle image upload
	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		utills.HandleError(w, http.StatusBadRequest, "Error retrieving file: "+err.Error())
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utills.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			utills.HandleError(w, http.StatusInternalServerError, "Error saving image")
			return
		}
		vendor.Img = &imageName
	}

	vendor.UpdatedAt = time.Now()

	query, args, err = QB.Update("vendors").
		Set("name", vendor.Name).
		Set("description", vendor.Description).
		Set("img", vendor.Img).
		Set("updated_at", vendor.UpdatedAt).
		Where(squirrel.Eq{"id": vendor.ID}).
		ToSql()
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error generating update query")
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error updating vendor")
		return
	}

	utills.SendJSONRespone(w, http.StatusOK, vendor)
}

func DeleteVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	query, args, err := QB.Delete("vendors").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error generating delete query")
		return
	}

	if _, err := db.Exec(query, args...); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error deleting vendor")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


