# Product Intelligence Roadmap

## Objetivo

Esta hoja de ruta define una linea de evolucion para llevar el backend desde gestion operativa de inventario hacia soporte inteligente de decisiones.

La idea no es agregar IA por moda. El foco es construir capacidades que mejoren:

- calidad de catalogo
- continuidad operacional
- rotacion de inventario
- decisiones de transferencia
- sustitucion de productos
- deteccion de riesgos y comportamientos atipicos

## Principio rector

Cada feature avanzada debe combinar al menos parte de estas señales:

- reglas de negocio
- stock real por sucursal
- historial de ventas
- historial de transferencias
- feedback humano
- embeddings o similitud semantica
- contexto geográfico y operativo

## Principio operativo

La capa inteligente no debe tratarse como un recurso gratuito e ilimitado.

El core transaccional del producto puede permanecer siempre disponible, pero las capacidades de inteligencia deben diseñarse para ser:

- medibles
- limitables
- cacheables
- monetizables
- degradables de forma controlada

Esto permite proteger margen operativo, controlar infraestructura y convertir la inteligencia en una propuesta de valor real por plan o nivel de servicio.

## Fase 0: Gobernanza de Consumo

### 0. Quotas por compania y por feature

Objetivo:
medir y limitar el uso de capacidades intensivas en modelo o ranking.

Valor:

- evita abuso accidental o uso desbalanceado
- protege costo por cliente
- habilita planes comerciales diferenciados

Eventos sugeridos:

- `semantic_neighbor_search`
- `embedding_generation`
- `catalog_dedup_scan`
- `transfer_recommendation_run`

### 0.1 Feature gating por plan

Objetivo:
separar claramente el core del sistema de las capacidades inteligentes premium.

Valor:

- permite empaquetar mejor el producto
- evita regalar costo computacional sin retorno

Ejemplos:

- plan base: inventario, ventas, transferencias
- plan intermedio: vecinos cercanos con limite mensual
- plan avanzado: catalogo inteligente, sustitucion y recomendaciones preventivas

### 0.2 Ventanas horarias para procesos intensivos

Objetivo:
ejecutar tareas pesadas en horarios baratos o de menor presion operativa.

Valor:

- reduce costo de infraestructura
- evita competir con trafico transaccional

Procesos candidatos:

- re-embedding masivo
- deteccion de duplicados
- recalculo de rankings preventivos
- limpieza y enriquecimiento de catalogo

### 0.3 Cache y precalculo de resultados

Objetivo:
reutilizar resultados de inteligencia cuando el catalogo o el contexto no cambiaron.

Valor:

- menor latencia
- menor uso del modelo
- mejor margen por cliente

Casos sugeridos:

- top de vecinos cercanos por producto
- sustitutos mas aceptados
- recomendaciones de transferencia de alta rotacion

### 0.4 Degradacion controlada

Objetivo:
mantener experiencia usable incluso cuando una cuota se agota o un servicio inteligente no esta disponible.

Valor:

- evita que una capacidad premium rompa el flujo operativo base
- mejora resiliencia del producto

Respuestas posibles:

- fallback a busqueda simple
- resultados cacheados
- limite temporal por compania
- mensaje comercial de upgrade o renovacion de cuota

## Fase 1: Base Inteligente

### 1. Deteccion de duplicados o casi duplicados

Objetivo:
detectar productos muy parecidos cargados varias veces con distinto nombre, descripcion o SKU auxiliar.

Valor:

- reduce ruido en catalogo
- evita dispersion de inventario
- mejora precision de reportes y recomendaciones

Datos requeridos:

- `sku`
- `name`
- `description`
- `brand`
- `category`
- `embedding`

### 2. Similitud enriquecida de productos

Objetivo:
dejar de usar solo cercania vectorial y pasar a un score hibrido que combine similitud semantica con señales operativas.

Valor:

- sugerencias mas utiles para el atendedor
- menor riesgo de recomendar productos parecidos pero poco viables

Señales sugeridas:

- distancia vectorial
- misma categoria
- misma marca
- rango de precio
- stock disponible
- historial de aceptacion

### 3. Captura de feedback sobre sugerencias

Objetivo:
registrar si una sugerencia de producto similar fue aceptada, rechazada o ignorada.

Valor:

- permite aprender desde el uso real
- habilita ranking futuro basado en evidencia

Datos sugeridos:

- producto consultado
- producto sugerido
- usuario
- sucursal
- accion tomada
- timestamp

## Fase 2: Soporte a la Decision

### 4. Sustitucion inteligente de productos

Objetivo:
sugerir reemplazos no solo por similitud, sino por probabilidad real de servir en una venta.

Valor:

- disminuye ventas perdidas
- ayuda a operar mejor ante quiebres de stock

Salida esperada:

- top de sustitutos
- porcentaje de parecido
- score operativo
- explicacion breve de por que aparece sugerido

### 5. Score de riesgo de quiebre por producto-sucursal

Objetivo:
anticipar que combinaciones producto-sucursal tienen mayor probabilidad de quedarse sin stock pronto.

Valor:

- prioriza reposicion
- reduce quiebres
- ayuda a ordenar transferencias internas

Señales sugeridas:

- stock actual
- velocidad de salida
- estacionalidad
- retraso promedio de reposicion
- presencia de sustitutos

### 6. Forecast de demanda por sucursal

Objetivo:
predecir demanda futura con señales mas ricas que un promedio historico simple.

Valor:

- mejora decisiones de stock
- soporta planeamiento y transferencia preventiva

Señales sugeridas:

- ventas historicas
- dia de semana
- estacionalidad
- patrones locales por sucursal
- comportamiento de productos similares

## Fase 3: Optimizacion Operativa

### 7. Recomendador de transferencias preventivas

Objetivo:
sugerir movimientos de stock antes de que una sucursal llegue al quiebre.

Valor:

- convierte la operacion en proactiva
- reduce urgencias y transferencias tardias

Salidas sugeridas:

- producto
- sucursal origen
- sucursal destino
- cantidad sugerida
- prioridad
- impacto esperado

### 8. Ranking de prioridad de transferencias

Objetivo:
ordenar solicitudes de transferencia segun impacto esperado.

Valor:

- ayuda a operar cuando hay varias solicitudes abiertas
- favorece decisiones con mayor retorno operativo

Señales sugeridas:

- riesgo de quiebre
- margen del producto
- urgencia
- distancia
- stock disponible en origen
- probabilidad de concretar la recepcion

### 9. Stock objetivo por sucursal

Objetivo:
definir un rango recomendado de stock por producto y sucursal.

Valor:

- ayuda a evitar sobrestock y quiebre
- soporta mejores decisiones de compra y traslado

## Fase 4: Inteligencia Comercial y de Riesgo

### 10. Deteccion de anomalías operativas

Objetivo:
identificar ventas, transferencias o movimientos fuera del patron esperado.

Valor:

- apoyo a auditoria
- deteccion temprana de errores o comportamientos sospechosos

Casos sugeridos:

- transferencias inusuales
- mermas extrañas
- caidas abruptas de rotacion
- ventas fuera de patron

### 11. Clustering de productos por comportamiento real

Objetivo:
agrupar productos segun como se venden, se transfieren y se sustituyen, no solo por categoria declarada.

Valor:

- descubre familias operativas reales
- mejora politicas de stock
- mejora sugerencias futuras

### 12. Clustering de sucursales

Objetivo:
identificar sucursales con comportamiento similar para transferir aprendizajes y estrategias.

Valor:

- facilita planeamiento regional
- permite comparaciones mas utiles que una segmentacion manual

## Fase 5: Aprendizaje Continuo

### 13. Learning-to-rank para sugerencias

Objetivo:
ordenar recomendaciones de productos usando feedback historico real.

Valor:

- mejora tasa de aceptacion de sustituciones
- reduce ruido en las sugerencias

### 14. Experimentacion sobre recomendaciones

Objetivo:
medir si una estrategia de sugerencias funciona mejor que otra.

Valor:

- evita decisiones basadas solo en intuicion
- permite evolucionar con evidencia

### 15. Algoritmos de exploracion-explotacion

Objetivo:
probar variantes de sugerencia sin dejar de priorizar las que mejor funcionan.

Valor:

- aprendizaje continuo
- adaptacion dinamica a cambios de catalogo y demanda

## Prioridad sugerida

Orden recomendado de implementacion:

1. deteccion de duplicados de catalogo
2. similitud enriquecida de productos
3. captura de feedback humano
4. sustitucion inteligente de productos
5. score de riesgo de quiebre
6. recomendador de transferencias preventivas
7. deteccion de anomalías operativas

## Dependencias tecnicas utiles

Antes de avanzar mucho en estas fases, conviene fortalecer:

- historico de sugerencias y decisiones humanas
- trazabilidad de transferencias y motivos
- stock por sucursal mas rico
- datos limpios de catalogo
- metrica de aceptacion de sustituciones

## Criterio de adopcion

Cada feature nueva deberia justificar claramente:

- que decision mejora
- que datos necesita
- como se valida
- como se mide su impacto

La meta final no es tener un sistema “inteligente” en abstracto, sino un backend capaz de mejorar decisiones comerciales y operativas con senales que un software de inventario tradicional normalmente no aprovecha.
