package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Create the upload route
	r.POST("/upload", func(c *gin.Context) {
		// Parse multipart form
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image from form"})
			return
		}

		// Ensure the `images` directory exists
		dir := "./images"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create images directory"})
				return
			}
		}

		// Create the destination file path
		filePath := filepath.Join(dir, file.Filename)

		// Save the file
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"message": "Image uploaded successfully",
			"path":    filePath,
		})
	})

	// Start the server
	r.Run(":8080") // Listen and serve on localhost:8080
}

// package main

// import (
// 	"acwj/db"
// 	"acwj/routes"

// 	"github.com/gin-gonic/gin"
// )

// func main() {
// 	// Connect to database
// 	db.Connect()
// 	db.Migrate()

// 	// Initialize Gin
// 	r := gin.Default()

// 	// Setup routes
// 	routes.UserRoutes(r)

// 	// Start the server
// 	r.Run(":8080")
// }
