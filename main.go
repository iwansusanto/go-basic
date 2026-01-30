package main

import (
	"fmt"
	"log"
	"net/http"

	"kasir-api/database"
	"kasir-api/docs"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
	"kasir-api/utils"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Kasir API
// @version         1.0
// @description     This is a sample server for a Cashier System.
// @BasePath        /api

func main() {
	// load .env using viper
	viper.SetConfigFile(".env")
	viper.AutomaticEnv() // read value from system env too

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	portStr := viper.GetString("PORT")
	if portStr == "" {
		portStr = "8080"
	}

	// Dynamic Swagger Host
	docs.SwaggerInfo.Host = "localhost:" + portStr

	// connect to DB
	dbConnStr := viper.GetString("DATABASE_URL")
	db, err := database.Connect(dbConnStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	fmt.Println("Successfully connected to database!")

	// {{host}}/health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, utils.Response{
			Status:  "success",
			Message: "API Running",
		})
	})

	// Swagger
	http.HandleFunc("/", httpSwagger.WrapHandler)

	// Routes
	http.HandleFunc("/api/category/", func(w http.ResponseWriter, r *http.Request) {
		categoryRepo := repositories.NewCategoryRepository(db)
		categoryService := services.NewCategoryService(categoryRepo)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		switch r.Method {
		case "GET":
			categoryHandler.GetCategoryByID(w, r)
		case "PUT":
			categoryHandler.UpdateCategory(w, r)
		case "DELETE":
			categoryHandler.DeleteCategory(w, r)
		default:
			utils.WriteJSON(w, http.StatusMethodNotAllowed, utils.Response{
				Status:  "failed",
				Message: "Method not allowed",
			})
		}
	})

	http.HandleFunc("/api/category", func(w http.ResponseWriter, r *http.Request) {
		categoryRepo := repositories.NewCategoryRepository(db)
		categoryService := services.NewCategoryService(categoryRepo)
		categoryHandler := handlers.NewCategoryHandler(categoryService)

		switch r.Method {
		case "GET":
			categoryHandler.GetCategories(w, r)
		case "POST":
			categoryHandler.CreateCategory(w, r)
		default:
			utils.WriteJSON(w, http.StatusMethodNotAllowed, utils.Response{
				Status:  "failed",
				Message: "Method not allowed",
			})
		}
	})

	fmt.Println("Server running on http://localhost:" + portStr)
	err = http.ListenAndServe(":"+portStr, nil)
	if err != nil {
		fmt.Println("Error running server:", err)
	}
}
