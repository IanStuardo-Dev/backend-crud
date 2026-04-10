# QA Validation Guide

## Objetivo

Esta guia resume un flujo practico para validar el backend sin depender de conocimiento previo del proyecto ni de IDs fijos en base de datos.

> [!NOTE]
> El foco esta puesto en las dos capacidades mas sensibles del estado actual:
> - transferencias con aprobacion
> - similitud semantica de productos

## Stack recomendado

Levantar la pila completa:

```sh
docker compose up -d --build
```

Verificar:

```sh
docker compose ps
curl http://localhost:8000/health
```

## Datos base

### Usuarios seed

- `superadmin@example.com` / `Password123`
- `admin@default-company.local` / `Password123`
- `inventory@default-company.local` / `Password123`
- `sales@default-company.local` / `Password123`

### Catalogo seed

Incluye:

- productos base
- set de cafe para validacion semantica
- 100 abarrotes para pruebas masivas de vecinos

> [!IMPORTANT]
> Si necesitas que los productos seed usen embeddings semanticos reales, corre de nuevo los seeds con `EMBEDDING_PROVIDER=local-semantic-service`.

Si ejecutas la API o el seeder fuera de `docker compose`, recuerda agregar tambien:

- `EMBEDDING_GRPC_TARGET=localhost:50051`

## Ruta rapida con Postman

La coleccion ya incluye una carpeta:

- `Semantic Search`

Orden recomendado:

1. `Login Sales User`
2. `Refresh Semantic Seed IDs`
3. `Get Coffee Seed Neighbors`
4. `Get Grocery Seed Neighbors`

La request `Refresh Semantic Seed IDs` evita depender de IDs fijos y captura automaticamente:

- `coffeeProductId`
- `coffeeSimilarProductId`
- `groceryProductId`

## Validaciones funcionales sugeridas

### 1. Login y permisos

Validar que:

- `sales_user` pueda consultar productos e inventario
- `inventory_manager` pueda operar transferencias
- `company_admin` pueda aprobar transferencias

### 2. Workflow de transferencias

Validar que:

- la transferencia exige `supervisor_user_id`
- no se autoaprueba al crearla
- `approve` cambia estado pero no mueve stock
- `dispatch` descuenta stock en origen
- `receive` acredita stock en destino
- `cancel` respeta las reglas del estado actual

Endpoints clave:

- `POST /inventory/transfers`
- `GET /inventory/transfers`
- `GET /inventory/transfers/branches/{branch_id}`
- `POST /inventory/transfers/{id}/approve`
- `POST /inventory/transfers/{id}/dispatch`
- `POST /inventory/transfers/{id}/receive`
- `POST /inventory/transfers/{id}/cancel`

### 3. Vecinos cercanos de cafe

Caso esperado:

- producto fuente:
  - `Cafe Dolca Instantaneo 170g`

Resultados esperables arriba del ranking:

- `Cafe Nestle Clasico 170g`
- `Coffee Marley Instant Blend 170g`

### 4. Vecinos cercanos de abarrotes

Caso sugerido:

- producto fuente:
  - `Arroz Grano Largo Gallo 1kg`

Resultados esperables cerca del top:

- `Arroz Integral CampoLindo 1kg`
- `Arroz Jasmine Oriente 1kg`

> [!TIP]
> No esperes que el top 5 sea “perfecto” en todos los casos. Lo importante en QA es comprobar que los resultados sean razonables, estables y mejores que una cercania puramente literal.

## Verificaciones tecnicas

### Coleccion Postman

```sh
jq empty postman/backend-crud.postman_collection.json
```

### Test suite Go

```sh
go test ./...
```

### Tests del servicio de embeddings local

```sh
python3 -m unittest services/embedding-service/test_app.py services/embedding-service/test_runtime.py
```

## Riesgos conocidos

> [!WARNING]
> La columna actual sigue en `1536` dimensiones y el servicio hace zero-padding sobre un embedding base de `384`. Eso es intencional por compatibilidad, pero no es la version final ideal del modelo de datos.

> [!WARNING]
> Si los seeds se cargan con el proveedor hash en vez del proveedor semantico, los resultados de vecinos cercanos pueden verse correctos solo en casos muy obvios y engañar la validacion.
