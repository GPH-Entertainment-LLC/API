package middleware

import (
	"fmt"
	"time"
	_ "xo-packs/db"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
		}
		c.Next()
	}
}

func DBMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

// func DBMiddleware2() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// db connection
// 		attempts := 0

// 		for attempts < 3 {
// 			dbConn, err := db.NewDB()
// 			if err != nil {
// 				fmt.Println("Issue connecting to DB, trying again..")
// 			} else {
// 				c.Set("DB", dbConn)
// 				c.Next()
// 				fmt.Println("Closing DB Connection")
// 				dbConn.Close()
// 			}
// 		}
// 	}
// }

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		latency := time.Since(t)

		fmt.Printf("%s %s %s %s\n",
			c.Request.Method,
			c.Request.RequestURI,
			c.Request.Proto,
			latency,
		)
	}
}

func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		c.Next()

		fmt.Printf("%d %s %s\n",
			c.Writer.Status(),
			c.Request.Method,
			c.Request.RequestURI,
		)
	}
}
