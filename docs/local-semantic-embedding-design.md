# Local Semantic Embedding Design

## Objetivo

Este documento describe la implementacion actual del flujo de embeddings semanticos locales y las decisiones tecnicas que permiten usar similitud semantica sin sacar el proyecto de su arquitectura actual.

> [!IMPORTANT]
> La meta no es solo encontrar nombres parecidos. La meta es aproximar relaciones utiles entre productos con variantes de escritura, tildes, unidades y terminos cercanos como `cafe`, `café` y `coffee`.

## Estado actual

El backend ya no genera embeddings dentro del binario principal. La inferencia semantica vive en un servicio hermano:

- `services/embedding-service`

La API de Go sigue dependiendo del puerto:

- `internal/application/product.Embedder`

Y la seleccion del proveedor queda encapsulada en infraestructura:

- `internal/infrastructure/embedding/provider`

Esto mantiene limpio el dominio y permite cambiar la estrategia de embeddings sin tocar handlers ni casos de uso.

## Arquitectura

```text
HTTP Product Handler
-> Product Use Case
-> Embedder port
-> Infrastructure provider selection
-> HTTP semantic embedder adapter
-> Local embedding-service
-> FastEmbed / ONNX
-> Local multilingual model
```

> [!NOTE]
> El backend principal no conoce detalles del runtime del modelo. Solo conoce el contrato del `Embedder`.

## Servicio actual

La implementacion local usa:

- runtime: `FastEmbed`
- modelo: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`
- inferencia: CPU local
- despliegue: `docker compose`

Archivo principal:

- `services/embedding-service/app.py`

### Endpoint interno

```http
POST /embed
```

Body:

```json
{
  "text": "Cafe Dolca instantaneo 170g"
}
```

Respuesta:

```json
{
  "embedding": [0.0123, -0.0171, 0.0044],
  "dimensions": 1536,
  "base_dimensions": 384,
  "model": "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
}
```

## Normalizacion previa

Antes de inferir, el texto pasa por una capa local de normalizacion en:

- `services/embedding-service/text_normalization.py`

Reglas actuales:

- lowercase
- eliminacion de tildes
- separacion de unidades pegadas a numeros
- limpieza de caracteres no utiles
- colapso de espacios
- alias locales de negocio

Ejemplos:

- `café` -> `cafe`
- `coffee` -> `cafe`
- `coffe` -> `cafe`
- `170gr` -> `170 g`
- `instant` -> `instantaneo`

> [!TIP]
> Esta capa vale tanto como el modelo. En catalogos pequenos y medianos, una buena normalizacion suele mover mucho la calidad del ranking.

## Compatibilidad con pgvector

La base actual usa:

- `embedding vector(1536)`

El modelo local produce embeddings base de:

- `384` dimensiones

Para no romper el esquema actual, el servicio hace zero-padding hasta `1536`.

Eso permite:

- mantener compatibilidad inmediata con la columna actual
- evitar una migracion de esquema urgente
- probar semantica real sin redisenar toda la persistencia

> [!WARNING]
> El zero-padding es una estrategia de compatibilidad, no el estado final ideal. Si esta capacidad se consolida, conviene migrar a un esquema versionado como `embedding_v2`.

## Seleccion de proveedor

Variables relevantes:

- `EMBEDDING_PROVIDER`
- `EMBEDDING_SERVICE_URL`
- `EMBEDDING_REQUEST_TIMEOUT`

Valores esperados hoy:

- `local-hash`
- `local-semantic-service`
- `disabled`

Comportamiento:

- `local-hash`: fallback ligero para uso simple
- `local-semantic-service`: proveedor recomendado para validacion semantica real
- `disabled`: sin generacion automatica

## Seeds y catalogo de prueba

El catalogo seed ya incluye:

- un set pequeno para similitud obvia:
  - `Wireless Mouse`
  - `Wireless Ergonomic Mouse`
- un set de cafe:
  - `Cafe Dolca Instantaneo 170g`
  - `Cafe Nestle Clasico 170g`
  - `Coffee Marley Instant Blend 170g`
- un set amplio de:
  - `100` abarrotes (`SEED-GRC-*`)

Esto permite validar:

- similitud lexical
- similitud semantica con variantes de idioma
- comportamiento en un catalogo mas grande y heterogeneo

## Validaciones utiles

### Salud del servicio

```sh
curl http://localhost:8000/health
```

### Regenerar seeds con proveedor semantico

```sh
EMBEDDING_PROVIDER=local-semantic-service \
EMBEDDING_SERVICE_URL=http://localhost:8000 \
EMBEDDING_REQUEST_TIMEOUT=45s \
go run ./cmd/seed
```

### Probar vecinos cercanos

```sh
curl 'http://localhost:8080/products/44/neighbors?limit=5&min_similarity=0.20' \
  -H "Authorization: Bearer TU_TOKEN"
```

## Costos y tradeoffs

Lo mas caro de esta capacidad no es la API en Go ni PostgreSQL, sino el `embedding-service`.

Tradeoffs actuales:

- mas RAM fija por tener el modelo cargado
- mas complejidad operativa que el hash local
- mucha mejor calidad semantica para catalogos reales

Este costo se compensa mejor cuando:

- la feature se limita por cuota o plan
- se cachean resultados
- se programan tareas pesadas fuera de horario punta

## Siguientes pasos tecnicos

- versionar embeddings en base de datos
- agregar score hibrido mas alla del vector puro
- capturar feedback humano sobre sugerencias
- precalcular vecinos de productos de alta rotacion
- instrumentar uso por compania para monetizar la capa inteligente
