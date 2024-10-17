package main

import (
	"InternshipProject/controllers"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"github.com/go-michi/michi"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("DOMAIN environment variable is not set")
	}
	fmt.Println("Loaded DOMAIN:", domain) // Debugging print statement

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("file:///" + GetRootpath("database/migrations"))
	mig, err := migrate.New(
		"file://"+GetRootpath("database/migrations"),
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := mig.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		log.Printf("migrations: %s", err.Error())
	}

	defer db.Close()

	controllers.SetDB(db)

	r := michi.NewRouter()

	r.Route("/", func(sub *michi.Router) {
		
		sub.Handle("GET", "/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	
		
		sub.Handle("GET", "/", http.HandlerFunc(controllers.IndexUserHandler)) 
		sub.Handle("GET", "/{id}", http.HandlerFunc(controllers.ShowUserHandler)) 
		sub.Handle("PUT", "/{id}", http.HandlerFunc(controllers.UpdateUserHandler)) 
		sub.Handle("DELETE", "/{id}", http.HandlerFunc(controllers.DeleteUserHandler))
	
		sub.Route("/vendors", func(router *michi.Router) {
			router.Use(controllers.JWTMiddleware) 
			router.Handle("GET", "/", http.HandlerFunc(controllers.IndexVendorHandler)) 
			router.Handle("GET", "/{id}", http.HandlerFunc(controllers.ShowVendorHandler)) 
			router.Handle("POST", "/", http.HandlerFunc(controllers.StoreVendorHandler)) 
			router.Handle("PUT", "/{id}", http.HandlerFunc(controllers.UpdateVendorHandler)) 
			router.Handle("DELETE", "/{id}", http.HandlerFunc(controllers.DeleteVendorHandler)) 
		})
	})	
}
	fmt.Println("Starting server on :8000")
	http.ListenAndServe(":8000", r)

func GetRootpath(dir string) string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(path.Dir(ex), dir)
}
