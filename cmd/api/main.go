package main

import (
	"log"
	"net/http"
	"time"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	inventoryhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/inventory"
	producthttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/product"
	salehttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/sale"
	transferhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/transfer"
	userhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/user"
	postgresinventory "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/inventory"
	postgresproduct "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/product"
	postgressale "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/sale"
	postgrestransfer "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/transfer"
	postgresuser "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/user"
	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
	inventoryapp "github.com/IanStuardo-Dev/backend-crud/internal/application/inventory"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	saleapp "github.com/IanStuardo-Dev/backend-crud/internal/application/sale"
	transferapp "github.com/IanStuardo-Dev/backend-crud/internal/application/transfer"
	userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/config"
	embeddingprovider "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/provider"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/http/router"
	postgresdb "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/persistence/postgres"
	jwtinfra "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/security/jwt"
	passwordinfra "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/security/password"
)

func main() {
	dsn := config.GetDatabaseDSN()
	sqlDB, err := postgresdb.New(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer sqlDB.Close()

	passwordHasher := passwordinfra.NewBcryptHasher(0)
	tokenService := jwtinfra.NewService(
		config.GetJWTSecret(),
		config.GetJWTIssuer(),
		config.GetJWTDuration(),
	)

	repo := postgresuser.NewRepository(sqlDB)
	useCase := userapp.NewUseCase(repo, passwordHasher)
	handler := userhttp.NewHandler(useCase)
	authUseCase := authapp.NewUseCase(repo, passwordHasher, tokenService)
	authHandler := authhttp.NewHandler(authUseCase)
	authMiddleware := authhttp.NewMiddleware(tokenService)
	inventoryRepo := postgresinventory.NewRepository(sqlDB)
	inventoryUseCase := inventoryapp.NewUseCase(inventoryRepo)
	inventoryHandler := inventoryhttp.NewHandler(inventoryUseCase)
	txManager := postgresdb.NewTransactionManager(sqlDB)
	productCatalogStore := postgresproduct.NewCatalogStore(sqlDB)
	productSimilarityStore := postgresproduct.NewSimilarityStore(sqlDB)
	productFeedbackStore := postgresproduct.NewFeedbackStore(sqlDB)
	productEmbedder, embeddingProviderName := embeddingprovider.NewProductEmbedder()
	if closer, ok := productEmbedder.(interface{ Close() error }); ok {
		defer closer.Close()
	}
	log.Printf("product embedding provider: %s", embeddingProviderName)
	productUseCase := productapp.NewUseCase(productCatalogStore, productCatalogStore, productSimilarityStore, productFeedbackStore, productEmbedder)
	productHandler := producthttp.NewHandler(productUseCase)
	saleStore := postgressale.NewSaleStore(sqlDB)
	saleStockStore := postgressale.NewStockStore(sqlDB)
	saleMovementStore := postgressale.NewMovementStore(sqlDB)
	saleReferenceStore := postgressale.NewReferenceStore(sqlDB)
	saleUseCase := saleapp.NewUseCase(txManager, saleStore, saleStore, saleStockStore, saleStockStore, saleMovementStore, saleReferenceStore, saleReferenceStore)
	saleHandler := salehttp.NewHandler(saleUseCase)
	transferStore := postgrestransfer.NewTransferStore(sqlDB)
	transferStockStore := postgrestransfer.NewStockStore(sqlDB)
	transferMovementStore := postgrestransfer.NewMovementStore(sqlDB)
	transferReferenceStore := postgrestransfer.NewReferenceStore(sqlDB)
	transferUseCase := transferapp.NewUseCase(txManager, transferStore, transferStore, transferStockStore, transferStockStore, transferMovementStore, transferReferenceStore, transferReferenceStore)
	transferHandler := transferhttp.NewHandler(transferUseCase)
	r := router.New(handler, authHandler, authMiddleware, inventoryHandler, productHandler, saleHandler, transferHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("listening on :8080")
	log.Fatal(server.ListenAndServe())
}
