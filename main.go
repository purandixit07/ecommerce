package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/purandixit07/ecommerce/controllers"
	"github.com/purandixit07/ecommerce/database"
	"github.com/purandixit07/ecommerce/middleware"
	"github.com/purandixit07/ecommerce/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuyer())

	log.Fatal(router.Run(":" + port))
}
