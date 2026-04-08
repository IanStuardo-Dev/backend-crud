# Inventory Backend

Backend en Go para operaciones de inventario, ventas y gestion multi-sucursal. El proyecto esta diseñado para equipos que necesitan trazabilidad operativa, control de acceso por compania y una base tecnica preparada para crecer sin comprometer orden ni mantenibilidad.

La solucion combina PostgreSQL, JWT, `pgvector` y una arquitectura por capas para separar con claridad el dominio, los casos de uso, la entrega HTTP y la infraestructura.

## Que resuelve

Este backend cubre necesidades habituales en sistemas comerciales y de inventario:

- usuarios autenticados con alcance por compania y rol
- productos asociados a una compania
- inventario por sucursal
- ventas que descuentan stock y registran que usuario ejecuto la accion
- transferencias internas entre sucursales
- busqueda de productos similares mediante embeddings locales y `pgvector`

La base prioriza consistencia, permisos y trazabilidad desde el inicio, para que las decisiones de negocio no queden diluidas entre handlers, queries y validaciones dispersas.

## Principios del proyecto

### Arquitectura clara

La organizacion sigue una separacion por capas:

- `cmd/`
  - entrypoints de ejecucion
- `internal/domain/`
  - entidades y reglas de negocio
- `internal/application/`
  - casos de uso, puertos y DTOs
- `internal/adapters/`
  - HTTP y repositorios concretos
- `internal/infrastructure/`
  - configuracion, seguridad, embeddings, router, seeds y acceso a PostgreSQL
- `migrations/`
  - cambios versionados de esquema

La razon de esta estructura no es cosmetica. Busca que el dominio no dependa de HTTP, SQL, JWT ni detalles externos, y que cada cambio tecnico tenga el menor impacto posible sobre la logica central.

### Enfoque pragmatico

El proyecto privilegia decisiones utiles para un backend real:

- JWT para autenticacion simple y portable
- `database/sql` para mantener control explicito sobre queries y transacciones
- `pgvector` para enriquecer productos con similitud semantica sin salir de PostgreSQL
- embeddings locales deterministas para no depender de proveedores externos en esta etapa

### Escalabilidad de dominio

La estructura ya soporta crecimiento funcional sin rehacer la base:

- nuevos modulos por feature
- nuevas fuentes de infraestructura
- evolucion de permisos
- estrategias de busqueda y recomendacion
- operaciones de inventario mas sofisticadas

## Capacidades actuales

### Seguridad y acceso

- login con JWT
- usuarios asociados a compania
- roles:
  - `super_admin`
  - `company_admin`
  - `inventory_manager`
  - `sales_user`
- proteccion de rutas por rol y alcance

### Productos

- gestion de catalogo de productos
- catalogo con `sku`, `name`, `description`, `category`, `brand`, precio y moneda
- generacion automatica de embedding local
- endpoint de vecinos cercanos:
  - `GET /products/{id}/neighbors`

### Inventario

- stock por sucursal en `branch_inventory`
- consulta de inventario por sucursal
- sugerencia de sucursales origen
- movimientos de inventario auditables

### Ventas y transferencias

- ventas autenticadas
- descuento de stock al vender
- registro del usuario que realiza la operacion
- transferencias internas entre sucursales de la misma compania

## Decisiones tecnicas importantes

### Por que `internal/`

El proyecto esta pensado como una aplicacion, no como una libreria reutilizable. Por eso el codigo principal vive en `internal/`, lo que ayuda a proteger los limites del sistema y evita exponer paquetes internos como API publica accidental.

### Por que arquitectura por capas

Porque aqui conviven varias preocupaciones que suelen mezclarse muy rapido:

- reglas de negocio
- autenticacion y autorizacion
- acceso a base de datos
- transporte HTTP
- embeddings y busqueda

Separarlas desde el inicio hace mas facil:

- testear casos de uso sin framework
- cambiar infraestructura sin romper el dominio
- crecer por modulos sin convertir handlers en centros de logica

### Por que `pgvector`

Porque la similitud de productos es una capacidad de negocio util dentro del mismo backend:

- sugerir reemplazos
- detectar productos cercanos por nombre y contexto
- preparar el terreno para catalogos mas inteligentes

En este proyecto `pgvector` se usa para similitud semantica de productos, no para geolocalizacion.

### Por que embeddings locales

En esta etapa se privilegio independencia operativa:

- sin costo por request a terceros
- sin dependencia de red
- resultados deterministas
- entorno local y CI mas simples

La implementacion actual toma informacion textual del producto, aplica hashing sobre tokens y bigramas y normaliza un vector de `1536` dimensiones.

## Modelo de datos

La base ya contempla piezas propias de un sistema transaccional:

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

Las sucursales incluyen ademas informacion operativa y geografica como:

- `latitude`
- `longitude`
- `address`
- `city`
- `region`
- `is_active`
- `opening_hours`

## Endpoints principales

### Publicos

- `POST /auth/login`

### Lectura y operacion comercial

- `GET /products`
- `GET /products/{id}`
- `GET /products/{id}/neighbors`
- `POST /sales`
- `GET /sales`
- `GET /sales/{id}`
- `GET /inventory/branch-items`
- `GET /inventory/source-candidates`

### Gestion administrativa y operativa

- `POST /users`
- `GET /users`
- `GET /users/{id}`
- `PUT /users/{id}`
- `DELETE /users/{id}`
- `POST /products`
- `PUT /products/{id}`
- `DELETE /products/{id}`
- `POST /inventory/transfers`
- `GET /inventory/transfers`
- `GET /inventory/transfers/{id}`

## Levantar el proyecto

### 1. Base de datos

```sh
docker compose up -d
```

Configuracion del contenedor:

- `POSTGRES_DB=go-crud`
- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`

### 2. Variables de entorno

Puedes usar `DATABASE_URL` o variables separadas.

Ejemplo:

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

Alternativamente:

```sh
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/go-crud?sslmode=disable'
```

Si existe un `.env`, la aplicacion intenta cargarlo automaticamente.

### 3. Migraciones

```sh
go run ./cmd/migrate up
```

O:

```sh
make migrate-up
```

### 4. Seeds

```sh
go run ./cmd/seed
```

### 5. API

```sh
go run ./cmd/api
```

La API queda disponible en:

```text
http://localhost:8080
```

## Datos seed

El proyecto incluye seeders idempotentes para acelerar desarrollo, demostraciones y pruebas funcionales.

Usuarios iniciales:

- `superadmin@example.com` / `Password123`
- `admin@default-company.local` / `Password123`
- `inventory@default-company.local` / `Password123`
- `sales@default-company.local` / `Password123`

Tambien se cargan 10 productos base en la compania `1`, sucursal `1`, con embeddings ya generados. Entre ellos se incluyen productos deliberadamente cercanos para validar la busqueda por similitud:

- `Wireless Mouse`
- `Wireless Ergonomic Mouse`

## Ejemplos rapidos

### Login

```sh
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "superadmin@example.com",
    "password": "Password123"
  }'
```

### Venta autenticada

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

### Vecinos cercanos

```sh
TOKEN='pega_aqui_el_jwt'

curl -X GET 'http://localhost:8080/products/4/neighbors?limit=5&min_similarity=0.20' \
  -H "Authorization: Bearer ${TOKEN}"
```

## Testing

```sh
make test
```

Opcionalmente:

```sh
make test-pretty
PATH="$(pwd)/bin:$PATH" make test-pretty
```

## Proyeccion

La base actual permite extender el sistema de forma ordenada en varias direcciones:

- reglas de negocio mas ricas por compania
- ajustes y reservas de inventario
- flujos de aprobacion
- stock por sucursal del mismo producto
- estrategias de recomendacion mas avanzadas
- integraciones externas cuando realmente aporten valor
