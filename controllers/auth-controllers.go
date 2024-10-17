package controllers

import (
	"InternshipProject/models"
	"InternshipProject/utills"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "encoding/json"

	_ "github.com/Masterminds/squirrel"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

// Load JWT secret from environment variable
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// GenerateJWT creates a new JWT token for authenticated users
func GenerateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	user := models.User{
		ID:       uuid.New(),
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	if user.Password == "" {
		utills.HandleError(w, http.StatusBadRequest, "Password is required")
		return
	}

	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		utills.HandleError(w, http.StatusBadRequest, "Invalid file")
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utills.SaveImageFile(file, "users", fileHeader.Filename)
		if err != nil {
			utills.HandleError(w, http.StatusInternalServerError, "Error saving image")
			return
		}
		user.Img = &imageName
	}

	hashedPassword, err := utills.HashPassword(user.Password)
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user.Password = hashedPassword

	query, args, err := QB.
		Insert("users").
		Columns("id", "img", "name", "phone", "email", "password").
		Values(user.ID, user.Img, user.Name, user.Phone, user.Email, user.Password).
		Suffix(fmt.Sprintf("RETURNING  %s", strings.Join(user_columns, ", "))).
		ToSql()

	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error generate query")
		return
	}

	if err := db.QueryRowx(query, args...).StructScan(&user); err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Error creating user"+err.Error())
		return
	}
	utills.SendJSONRespone(w, http.StatusCreated, user)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	err := db.Get(&user, "SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		utills.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		utills.HandleError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := GenerateJWT(user)
	if err != nil {
		utills.HandleError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	utills.SendJSONRespone(w, http.StatusOK, map[string]string{"token": token})
}


func JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            utills.HandleError(w, http.StatusUnauthorized, "Authorization header is missing")
            return
        }

        // Remove "Bearer " from the token
        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

        claims := jwt.MapClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil // Ensure jwtSecret is defined and accessible
        })

        if err != nil || !token.Valid {
            utills.HandleError(w, http.StatusUnauthorized, "Invalid token")
            return
        }

        // Check if "id" claim exists and is of type string
        id, ok := claims["id"].(string)
        if !ok {
            utills.HandleError(w, http.StatusUnauthorized, "Invalid token claims")
            return
        }

        // Attach the user ID to the request context
        ctx := context.WithValue(r.Context(), "userID", id)

        // Call the next handler with the new context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}




