package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"xo-packs/controller"
	"xo-packs/core"
	"xo-packs/db"
	"xo-packs/middleware"
	"xo-packs/repository"
	"xo-packs/service"

	_ "xo-packs/docs"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/celler/httputil"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

// @title 	XOPacks API
// @version	1.0
// @description The official XOPAcks API
// @host 	apidev.xopacks.com
// @BasePath /
func main() {
	// Get the current working directory (project root).
	currDir, err := os.Getwd()
	if err != nil {
		panic("ERROR GETTING CUR DIR")
	}

	if os.Getenv("ENV") == "local" {
		core.LoadLocalEnvironment(currDir)
	}

	// db connection
	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("SSL_MODE")
	password, err := core.GetSecret(os.Getenv("DB_PASSWORD"))
	if err != nil {
		fmt.Println(err)
		panic("ERROR GETTING DB PASSWORD")
	}

	dbConn, err := db.NewDB(user, password, host, port, dbname, sslmode)
	if err != nil {
		fmt.Println("Error connecting to DB: ", err)
	}
	defer dbConn.Close()

	// cache client
	cacheClient := db.NewCache()
	if cacheClient == nil {
		fmt.Println("Error connecting to cache client")
	}
	defer cacheClient.Close()

	// router setup
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ResponseLogger())
	router.Use(middleware.FulfilledRequestLoggingMiddleware())
	router.Use(middleware.RequestLoggingMiddleware())

	// repository instantiation
	userRepo := repository.NewUserRepo(dbConn, cacheClient)
	loggingRepo := repository.NewLoggingRepo(dbConn, cacheClient)
	vendorRepo := repository.NewVendorRepo(dbConn, cacheClient)
	tokenRepo := repository.NewTokenRepository(dbConn, cacheClient)
	packRepo := repository.NewPackRepo(dbConn, cacheClient)
	itemRepo := repository.NewItemRepo(dbConn, cacheClient)
	categoryRepo := repository.NewCategoryRepo(dbConn, cacheClient)
	analyticsRepo := repository.NewAnalyticsRepo(dbConn, cacheClient)
	adminRepo := repository.NewAdminRepo(dbConn, cacheClient)
	applicationRepo := repository.NewApplicationRepo(dbConn, cacheClient)
	referralRepo := repository.NewReferralRepo(dbConn, cacheClient)
	reportRepo := repository.NewReportRepo(dbConn, cacheClient)
	transactionRepo := repository.NewTransactionRepo(dbConn, cacheClient)
	financialRepo := repository.NewFinancialRepo(dbConn, cacheClient)

	// services
	userService := service.NewUserService(userRepo)
	loggingService := service.NewLoggingService(loggingRepo)
	vendorService := service.NewVendorService(vendorRepo)
	tokenService := service.NewTokenService(tokenRepo)
	packService := service.NewPackService(packRepo)
	itemService := service.NewItemService(itemRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	firebaseService := service.NewFirebaseSvc()
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	adminService := service.NewAdminService(adminRepo)
	applicationService := service.NewApplicationService(applicationRepo)
	referralService := service.NewReferralService(referralRepo)
	reportService := service.NewReportService(reportRepo)
	transactionService := service.NewTransactionService(transactionRepo)
	financialService := service.NewFinancialService(financialRepo)

	// controller instantiation
	userContr := controller.NewUserController(userService, vendorService, itemService)
	vendorContr := controller.NewVendorController(vendorService, categoryService, packService, itemService)
	tokenContr := controller.NewTokenController(tokenService)
	packContr := controller.NewPackController(packService, vendorService, itemService, userService, tokenService)
	loggingContr := controller.NewLoggingService(loggingService, userService)
	itemContr := controller.NewItemController(itemService, vendorService, packService)
	firebaseContr := controller.NewFirebaseController(firebaseService, userService)
	categoryContr := controller.NewCategoryController(categoryService)
	analyticsContr := controller.NewAnalyticsController(analyticsService)
	transactionContr := controller.NewTransactionController(transactionService, tokenService)
	adminContr := controller.NewAdminController(userService, applicationService, adminService)
	applicationContr := controller.NewApplicationController(applicationService, referralService)
	referralContr := controller.NewReferralController(referralService, vendorService)
	reportContr := controller.NewReportController(reportService)
	financialContr := controller.NewFinancialController(financialService)

	// controller registration
	userContr.Register(router)
	vendorContr.Register(router)
	tokenContr.Register(router)
	packContr.Register(router)
	loggingContr.Register(router)
	itemContr.Register(router)
	firebaseContr.Register(router)
	categoryContr.Register(router)
	analyticsContr.Register(router)
	transactionContr.Register(router)
	adminContr.Register(router)
	applicationContr.Register(router)
	referralContr.Register(router)
	reportContr.Register(router)
	financialContr.Register(router)

	InitRoutes(router)

	// core routes
	router.GET("/faqs", func(ctx *gin.Context) {
		faqs, err := core.GetFaqs(ctx, dbConn)
		if err != nil {
			httputil.NewError(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.JSON(http.StatusOK, faqs)
		return
	})

	// run app
	err = router.Run(":80")
	if err != nil {
		log.Fatal(err)
	}
}

func InitRoutes(router *gin.Engine) {
	// swagger
	ginSwagger.WrapHandler(swaggerfiles.Handler,
		ginSwagger.URL("https://apidev.xopacks.com/swagger/swagger.json"),
		ginSwagger.DefaultModelsExpandDepth(-1))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// base and greeting API routes
	router.GET("/", func(ctx *gin.Context) {
		ctx.IndentedJSON(http.StatusOK, os.Getenv("title"))
		return
	})
	router.GET("/test", func(ctx *gin.Context) {
		// build request to CCBill CBPT-API
		baseURL := "https://3l68dsk0ql.execute-api.us-east-2.amazonaws.com/dev/transaction/test"
		requestBody := []byte(`{}`)
		req, err := http.NewRequest("GET", baseURL, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		// Set the content type to JSON
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}
		fmt.Println("Response Status: ", resp.StatusCode)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}
		log.Printf("Response: %s", body)

		ctx.IndentedJSON(resp.StatusCode, string(body))
		return
	})
	router.GET("/hello", func(ctx *gin.Context) {
		ctx.IndentedJSON(http.StatusOK, os.Getenv("greeting"))
		return
	})
}
