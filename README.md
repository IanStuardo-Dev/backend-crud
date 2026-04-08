# Backend CRUD

Backend en Go con PostgreSQL, `pgvector`, JWT y una estructura inspirada en Clean Architecture. El proyecto ya no es solo un CRUD de usuarios: hoy incluye autenticacion, autorizacion por roles, productos, ventas, inventario por sucursal y transferencias internas.

## Stack

- Go 1.23
- `net/http` + `gorilla/mux`
- PostgreSQL 15
- `pgvector`
- JWT
- `database/sql` + `lib/pq`

## Arquitectura

La organizacion sigue una separacion por capas y responsabilidades:

- `cmd/`
  - puntos de entrada de la aplicacion (`api`, `migrate`, `seed`)
- `internal/domain/`
  - entidades y reglas de negocio puras
- `internal/application/`
  - casos de uso, DTOs y puertos
- `internal/adapters/`
  - HTTP y repositorios concretos
- `internal/infrastructure/`
  - config, JWT, password hashing, router, migraciones, seeds, conexion PostgreSQL
- `migrations/`
  - migraciones SQL versionadas

El router central solo compone subrouters y middlewares; cada feature registra sus propias rutas junto a su handler HTTP.

## Modulos actuales

- `auth`
  - login con JWT
- `users`
  - usuarios del sistema, asociados a compania y rol
- `products`
  - catalogo de productos con campo `embedding` para futuras busquedas semanticas
- `sales`
  - registro de ventas y descuento de stock
- `inventory`
  - consulta de inventario por sucursal y sugerencia de sucursales origen
- `transfers`
  - transferencias internas entre sucursales

## Modelo de autorizacion

Los usuarios tienen:

- `company_id`
- `role`
- `is_active`
- `default_branch_id`

Roles implementados:

- `super_admin`
  - acceso global
- `company_admin`
  - administra recursos de su compania
- `inventory_manager`
  - productos, inventario y transferencias de su compania
- `sales_user`
  - ventas e inventario de su compania

Reglas principales:

- solo `super_admin` puede operar sin restriccion por compania
- el resto de usuarios solo puede operar dentro de su `company_id`
- `sales` registra que usuario ejecuta el descuento de stock
- las rutas se protegen con JWT y chequeo de rol

## Base de datos

El modelo ya incluye, entre otras, estas piezas:

- `users`
- `companies`
- `branches`
- `products`
- `branch_inventory`
- `sales`
- `sale_items`
- `inventory_movements`
- `inventory_transfers`
- `inventory_transfer_items`

### Products y pgvector

`products` incluye un campo:

- `embedding vector(1536)`

El backend ahora genera ese embedding automaticamente al crear un producto si el request no lo trae. La implementacion actual es local y deterministica: tokeniza el texto del producto, aplica hashing sobre tokens y bigramas, y normaliza el vector final a `1536` dimensiones.

`pgvector` aqui se usa para productos, no para cercania geografica entre sucursales.

## Levantar el proyecto

### 1. Levantar PostgreSQL con Compose

```sh
docker compose up -d
```

El contenedor usa estos valores:

- `POSTGRES_DB=go-crud`
- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`

### 2. Configurar variables de entorno

Puedes usar `DATABASE_URL` o variables separadas. Con el `compose.yml` actual, una configuracion valida es:

```sh
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=go-crud
export DB_SSLMODE=disable
export JWT_SECRET=dev-secret-change-me
export JWT_ISSUER=crud-api
export JWT_TTL=24h
```

Tambien puedes usar:

```sh
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/go-crud?sslmode=disable'
```

Si quieres manejarlo desde archivo, puedes crear un `.env`; la app intenta cargarlo automaticamente con `godotenv`.

### 3. Aplicar migraciones

```sh
go run ./cmd/migrate up
```

O con `make`:

```sh
make migrate-up
```

### 4. Cargar seeds

```sh
go run ./cmd/seed
```

### 5. Ejecutar la API

```sh
go run ./cmd/api
```

La API escucha en:

```text
http://localhost:8080
```

## Seeds

El proyecto incluye seeders idempotentes para usuarios iniciales:

- `superadmin@example.com` / `Password123`
- `admin@default-company.local` / `Password123`
- `inventory@default-company.local` / `Password123`
- `sales@default-company.local` / `Password123`

Tambien carga 10 productos base en la compania `1`, sucursal `1`, con embeddings locales generados automaticamente. Entre ellos se incluyen dos nombres deliberadamente parecidos para pruebas de similitud:

- `Wireless Mouse`
- `Wireless Ergonomic Mouse`

Puedes sobreescribirlos con variables de entorno:

- `SEED_SUPER_ADMIN_EMAIL`
- `SEED_SUPER_ADMIN_PASSWORD`
- `SEED_COMPANY_ADMIN_EMAIL`
- `SEED_COMPANY_ADMIN_PASSWORD`
- `SEED_INVENTORY_MANAGER_EMAIL`
- `SEED_INVENTORY_MANAGER_PASSWORD`
- `SEED_SALES_USER_EMAIL`
- `SEED_SALES_USER_PASSWORD`

## Endpoints

### Publicos

- `POST /auth/login`

### Solo `company_admin` y `super_admin`

- `POST /users`
- `GET /users`
- `GET /users/{id}`
- `PUT /users/{id}`
- `DELETE /users/{id}`

### `company_admin`, `inventory_manager`, `sales_user` y `super_admin`

- `GET /products`
- `GET /products/{id}`
- `GET /products/{id}/neighbors`
- `POST /sales`
- `GET /sales`
- `GET /sales/{id}`
- `GET /inventory/branch-items`
- `GET /inventory/source-candidates`

### `company_admin`, `inventory_manager` y `super_admin`

- `POST /products`
- `PUT /products/{id}`
- `DELETE /products/{id}`
- `POST /inventory/transfers`
- `GET /inventory/transfers`
- `GET /inventory/transfers/{id}`

## Ejemplo rapido de login

```sh
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "superadmin@example.com",
    "password": "Password123"
  }'
```

## Ejemplo rapido de venta autenticada

```sh
TOKEN='pega_aqui_el_jwt'

curl -X POST http://localhost:8080/sales \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{
    "company_id": 1,
    "branch_id": 1,
    "items": [
      {
        "product_id": 1,
        "quantity": 1
      }
    ]
  }'
```

## Ejemplo rapido de vecinos cercanos

```sh
TOKEN='pega_aqui_el_jwt'

curl -X GET 'http://localhost:8080/products/4/neighbors?limit=5&min_similarity=0.20' \
  -H "Authorization: Bearer ${TOKEN}"
```

## Migraciones

CLI disponible en `cmd/migrate`:

- `go run ./cmd/migrate up`
- `go run ./cmd/migrate down`
- `go run ./cmd/migrate down 2`
- `go run ./cmd/migrate version`

Atajos en `Makefile`:

- `make migrate-up`
- `make migrate-down`
- `make migrate-version`

## Testing

```sh
make test
```

Opcionalmente:

```sh
make test-pretty
PATH="$(pwd)/bin:$PATH" make test-pretty
```

## Notas

- el modulo actual en `go.mod` es `github.com/IanStuardo-Dev/backend-crud`
- la cercania entre sucursales hoy se resuelve con `latitude` y `longitude`, no con `pgvector`
- no existe un endpoint `/health` por ahora
