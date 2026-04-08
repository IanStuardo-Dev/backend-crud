package main

import (
	"log"
	"net/http"
	"time"

	authhttp "github.com/example/crud/internal/adapters/http/auth"
	inventoryhttp "github.com/example/crud/internal/adapters/http/inventory"
	producthttp "github.com/example/crud/internal/adapters/http/product"
	salehttp "github.com/example/crud/internal/adapters/http/sale"
	transferhttp "github.com/example/crud/internal/adapters/http/transfer"
	userhttp "github.com/example/crud/internal/adapters/http/user"
	postgresinventory "github.com/example/crud/internal/adapters/repository/postgres/inventory"
	postgresproduct "github.com/example/crud/internal/adapters/repository/postgres/product"
	postgressale "github.com/example/crud/internal/adapters/repository/postgres/sale"
	postgrestransfer "github.com/example/crud/internal/adapters/repository/postgres/transfer"
	postgresuser "github.com/example/crud/internal/adapters/repository/postgres/user"
	authapp "github.com/example/crud/internal/application/auth"
	inventoryapp "github.com/example/crud/internal/application/inventory"
	productapp "github.com/example/crud/internal/application/product"
	saleapp "github.com/example/crud/internal/application/sale"
	transferapp "github.com/example/crud/internal/application/transfer"
	userapp "github.com/example/crud/internal/application/user"
	"github.com/example/crud/internal/infrastructure/config"
	localembedding "github.com/example/crud/internal/infrastructure/embedding/local"
	"github.com/example/crud/internal/infrastructure/http/router"
	postgresdb "github.com/example/crud/internal/infrastructure/persistence/postgres"
	jwtinfra "github.com/example/crud/internal/infrastructure/security/jwt"
	passwordinfra "github.com/example/crud/internal/infrastructure/security/password"
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
	productRepo := postgresproduct.NewRepository(sqlDB)
	productEmbedder := localembedding.NewService()
	productUseCase := productapp.NewUseCase(productRepo, productEmbedder)
	productHandler := producthttp.NewHandler(productUseCase)
	saleRepo := postgressale.NewRepository(sqlDB)
	saleUseCase := saleapp.NewUseCase(saleRepo)
	saleHandler := salehttp.NewHandler(saleUseCase)
	transferRepo := postgrestransfer.NewRepository(sqlDB)
	transferUseCase := transferapp.NewUseCase(transferRepo)
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
