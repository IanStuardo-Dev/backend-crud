package router

import (
	"net/http"

	authhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/auth"
	inventoryhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/inventory"
	producthttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/product"
	salehttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/sale"
	transferhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/transfer"
	userhttp "github.com/IanStuardo-Dev/backend-crud/internal/adapters/http/user"
	"github.com/gorilla/mux"
)

func New(
	userHandler *userhttp.Handler,
	authHandler *authhttp.Handler,
	authMiddleware *authhttp.Middleware,
	inventoryHandler *inventoryhttp.Handler,
	productHandler *producthttp.Handler,
	saleHandler *salehttp.Handler,
	transferHandler *transferhttp.Handler,
) http.Handler {
	root := mux.NewRouter()
	configureRootHandlers(root)

	public := root
	salesAccess := root
	inventoryManagerAccess := root
	adminAccess := root
	if authMiddleware != nil {
		salesAccess = root.PathPrefix("").Subrouter()
		salesAccess.Use(authMiddleware.RequireAuthentication)
		salesAccess.Use(authMiddleware.RequireRoles("company_admin", "inventory_manager", "sales_user"))

		inventoryManagerAccess = root.PathPrefix("").Subrouter()
		inventoryManagerAccess.Use(authMiddleware.RequireAuthentication)
		inventoryManagerAccess.Use(authMiddleware.RequireRoles("company_admin", "inventory_manager"))

		adminAccess = root.PathPrefix("").Subrouter()
		adminAccess.Use(authMiddleware.RequireAuthentication)
		adminAccess.Use(authMiddleware.RequireRoles("company_admin"))
	}

	authhttp.RegisterPublicRoutes(public, authHandler)
	userhttp.RegisterAdminRoutes(adminAccess, userHandler)
	inventoryhttp.RegisterInventoryRoutes(salesAccess, inventoryHandler)
	producthttp.RegisterReadRoutes(salesAccess, productHandler)
	producthttp.RegisterWriteRoutes(inventoryManagerAccess, productHandler)
	salehttp.RegisterSalesRoutes(salesAccess, saleHandler)
	transferhttp.RegisterTransferRoutes(inventoryManagerAccess, transferHandler)

	return root
}
