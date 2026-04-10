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
- busqueda de productos similares mediante embeddings semanticos locales y `pgvector`

La base prioriza consistencia, permisos y trazabilidad desde el inicio, para que las decisiones de negocio no queden diluidas entre handlers, queries y validaciones dispersas.

## Principios del proyecto

> [!NOTE]
> Documentos complementarios:
> - [Product Intelligence Roadmap](docs/product-intelligence-roadmap.md)
> - [Local Semantic Embedding Design](docs/local-semantic-embedding-design.md)
> - [QA Validation Guide](docs/qa-validation-guide.md)

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

Regla operativa para esta base:

- `application/product`, `application/sale` y `application/transfer` se organizan en subpackages por responsabilidad como `command`, `query`, `dto`, `ports` y `errors`
- los adapters PostgreSQL se segmentan en stores pequenos por capacidad, no en repositories monoliticos
- un repository es un adapter de persistencia, no un workflow engine
- la coordinacion transaccional vive en application layer
- `make guard-architecture` verifica que los archivos Go de estos modulos no excedan `200` lineas

### Enfoque pragmatico

El proyecto privilegia decisiones utiles para un backend real:

- JWT para autenticacion simple y portable
- `database/sql` para mantener control explicito sobre queries y transacciones
- `pgvector` para enriquecer productos con similitud semantica sin salir de PostgreSQL
- un servicio local de embeddings semanticos para mantener la inferencia fuera del backend principal

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
- generacion automatica de embedding semantico local
- endpoint de vecinos cercanos:
  - `GET /products/{id}/neighbors`
- captura de feedback operativo sobre sugerencias semanticas:
  - `POST /products/{id}/neighbors/{neighbor_id}/feedback`

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
- supervisor obligatorio para aprobar transferencias
- estados operativos para transferencias:
  - `pending_approval`
  - `approved`
  - `in_transit`
  - `received`
  - `cancelled`

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
- sin dependencia de APIs externas
- posibilidad de correr toda la pila en Docker Compose
- separacion clara entre backend transaccional e inferencia semantica

La implementacion actual usa un servicio local basado en `FastEmbed` y el modelo `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`, expuesto en la misma red del proyecto. La inferencia se consume desde la API Go via `gRPC`, mientras el servicio mantiene un endpoint HTTP de salud para operacion y Docker healthchecks. El servicio genera embeddings semanticos y los adapta a `1536` dimensiones para mantener compatibilidad con el esquema actual de `pgvector`.

Antes de generar el embedding, el texto pasa por una normalizacion local orientada al dominio: lowercase, eliminacion de tildes, unificacion de unidades y alias utiles como `coffee -> cafe`.

## Modelo de datos

La base ya contempla piezas propias de un sistema transaccional:

- `users`
- `companies`
- `branches`
- `products`
- `product_neighbor_feedback`
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
- `POST /products/{id}/neighbors/{neighbor_id}/feedback`
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
- `GET /inventory/transfers/branches/{branch_id}`
- `GET /inventory/transfers/{id}`
- `POST /inventory/transfers/{id}/approve`
- `POST /inventory/transfers/{id}/dispatch`
- `POST /inventory/transfers/{id}/receive`
- `POST /inventory/transfers/{id}/cancel`

## Levantar el proyecto

### 1. Stack local

```sh
docker compose up -d --build
```

Esto levanta:

- PostgreSQL con `pgvector`
- API en Go
- servicio local de embeddings semanticos

> [!IMPORTANT]
> El `embedding-service` forma parte de la pila local. Si quieres validar vecinos cercanos con el proveedor semantico real, no basta con levantar solo `db` y `api`.

Configuracion principal:

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
export EMBEDDING_PROVIDER=local-semantic-service
export EMBEDDING_GRPC_TARGET=localhost:50051
export EMBEDDING_REQUEST_TIMEOUT=45s
```

Alternativamente:

```sh
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/go-crud?sslmode=disable'
```

Si existe un `.env`, la aplicacion intenta cargarlo automaticamente.

Si ejecutas la API fuera de Compose, el proveedor por defecto sigue siendo `local-hash`. Para usar el servicio semantico local debes definir `EMBEDDING_PROVIDER=local-semantic-service`.

### 2.1 Contrato gRPC

El contrato entre la API Go y el `embedding-service` vive en:

- `proto/embedding/v1/embedding.proto`

Los stubs generados se versionan en el repo para Go y Python. Si cambias el `.proto`, regeneralos con:

```sh
make proto
```

Ese comando instala generadores versionados en una carpeta temporal y vuelve a crear los archivos generados de ambos servicios.

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

Si quieres que los productos seed usen el proveedor semantico local, ejecuta el seeder con estas variables:

```sh
EMBEDDING_PROVIDER=local-semantic-service \
EMBEDDING_GRPC_TARGET=localhost:50051 \
EMBEDDING_REQUEST_TIMEOUT=45s \
go run ./cmd/seed
```

> [!IMPORTANT]
> Si corres los seeds sin esas variables, el catalogo se puede poblar con el proveedor local por hash en vez del proveedor semantico. Para pruebas reales de vecinos cercanos, conviene regenerarlos con `local-semantic-service`.

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

Tambien se cargan productos base en la compania `1`, sucursal `1`, con embeddings ya generados.

Set de validacion semantica:

- `Wireless Mouse`
- `Wireless Ergonomic Mouse`
- `Cafe Dolca Instantaneo 170g`
- `Cafe Nestle Clasico 170g`
- `Coffee Marley Instant Blend 170g`

Catalogo adicional para pruebas masivas:

- `100` abarrotes distintos (`SEED-GRC-001` a `SEED-GRC-100`)
- categorias como pantry, breakfast, snacks, baking, canned-goods, desserts y hot-beverages
- familias cercanas como arroz, fideos, legumbres, cafe, te, galletas, aceites y conservas

> [!TIP]
> Este set de abarrotes sirve para probar el comportamiento del ranking semantico con un catalogo mas realista, sin depender de solo dos o tres productos artificialmente parecidos.

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

curl -X GET 'http://localhost:8080/products/ID_DEL_PRODUCTO/neighbors?limit=5&min_similarity=0.20' \
  -H "Authorization: Bearer ${TOKEN}"
```

La coleccion Postman incluye una carpeta `Semantic Search` que:

- obtiene el catalogo
- captura automaticamente los IDs de seeds relevantes
- prueba vecinos cercanos para cafe y abarrotes sin depender de IDs fijos

> [!TIP]
> En Postman, el orden recomendado es:
> 1. login
> 2. `Refresh Semantic Seed IDs`
> 3. `Get Coffee Seed Neighbors` o `Get Grocery Seed Neighbors`

## Testing

```sh
make test
```

Opcionalmente:

```sh
make test-pretty
PATH="$(pwd)/bin:$PATH" make test-pretty
```

## Flujo de ramas

El repositorio sigue un flujo simple de promocion entre ramas:

- `dev`
  - integracion de trabajo en curso mediante Pull Request
- `qa`
  - validacion de cambios que provienen solo desde `dev`
- `main`
  - rama principal, alimentada solo desde `qa`

Las ramas `dev`, `qa` y `main` estan pensadas para operar con protecciones de GitHub:

- merge solo por Pull Request
- checks requeridos antes de mergear
- bloqueo de force-push
- bloqueo de borrado de ramas

## Proyeccion

La base actual permite extender el sistema de forma ordenada en varias direcciones:

- reglas de negocio mas ricas por compania
- ajustes y reservas de inventario
- flujos de aprobacion
- stock por sucursal del mismo producto
- estrategias de recomendacion mas avanzadas
- integraciones externas cuando realmente aporten valor
