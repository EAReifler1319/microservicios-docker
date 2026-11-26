# Arquitectura de Microservicios con APIs Multilenguaje, SQLite y Docker

Cuatro APIs REST independientes, cada una escrita en un lenguaje distinto, con su propia base de datos SQLite y su propio contenedor Docker. Docker Compose las levanta juntas, y cada base de datos se guarda en un volumen persistente.

```
                         Docker Compose
   ┌──────────────┬──────────────┬──────────────┬──────────────┐
   │  Users API   │   Auth API   │  Orders API  │ Products API │
   │ Python :8000 │  Java :8001  │   Go :8002   │   JS :8003   │
   │   FastAPI    │   Javalin    │   net/http   │   Express    │
   │   users.db   │   auth.db    │  orders.db   │ products.db  │
   │  users_data  │  auth_data   │ orders_data  │products_data │
   └──────────────┴──────────────┴──────────────┴──────────────┘
          4 APIs · 4 contenedores · 4 bases de datos · 4 volúmenes
```

## Estructura del proyecto

```
MICROSERVICIOS/
├── users-api/        Python
│   ├── main.py
│   ├── requirements.txt
│   └── Dockerfile
├── auth-api/         Java
│   ├── src/main/java/Main.java
│   ├── pom.xml
│   └── Dockerfile
├── orders-api/       Go
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── products-api/     JavaScript
│   ├── main.js
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## Tecnologías

| API | Lenguaje | Framework | Driver SQLite | Puerto | Base de datos | Volumen |
|---|---|---|---|---|---|---|
| Users | Python 3.12 | FastAPI | `sqlite3` (incluido en Python) | 8000 | `users.db` | `users_data` |
| Auth | Java 21 | Javalin | `sqlite-jdbc` | 8001 | `auth.db` | `auth_data` |
| Orders | Go 1.25 | `net/http` (estándar) | `modernc.org/sqlite` | 8002 | `orders.db` | `orders_data` |
| Products | Node.js 24 | Express | `node:sqlite` (incluido en Node) | 8003 | `products.db` | `products_data` |

### Cómo se inicializan las bases de datos

No hay archivos `.db` en el repositorio. Cada API crea sus tablas al arrancar con `CREATE TABLE IF NOT EXISTS`. La ruta del archivo llega por la variable de entorno `DB_PATH` y siempre apunta a `/app/data/`, que es la carpeta donde se monta el volumen de esa API.

## Requisitos

- Docker Desktop (Windows o Mac) o Docker Engine con el plugin Compose (Linux).
- No hace falta instalar Python, Java, Go ni Node: todo se compila dentro de los contenedores.

## Ejecución

```bash
# Construir las imágenes y levantar los 4 servicios en segundo plano
docker compose up -d --build

# Ver el estado de los contenedores
docker compose ps

# Ver los logs (de todos o de uno)
docker compose logs
docker compose logs -f auth-api

# Detener los servicios (los contenedores y los datos se conservan)
docker compose stop

# Volver a iniciarlos
docker compose start

# Detener y ELIMINAR los contenedores (los volúmenes y los datos se conservan)
docker compose down

# ⚠️ Eliminar contenedores Y volúmenes (se borran todos los datos)
docker compose down -v
```

Al levantar los servicios, estas direcciones deberían responder:

| Servicio | URL |
|---|---|
| Users API | http://localhost:8000 (Swagger en http://localhost:8000/docs) |
| Auth API | http://localhost:8001/health |
| Orders API | http://localhost:8002 |
| Products API | http://localhost:8003 |

## Endpoints

Todas las respuestas son JSON. Los errores devuelven `{"detail": "mensaje"}` con su código HTTP (400, 401 o 404).

### Users API (8000)

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/` | Estado de la API |
| POST | `/users` | Registrar usuario `{name, email, password}` |
| GET | `/users` | Listar usuarios |
| GET | `/users/{id}` | Consultar usuario por ID |
| PUT | `/users/{id}` | Actualizar usuario (solo los campos enviados) |
| DELETE | `/users/{id}` | Eliminar usuario |

La contraseña se guarda como hash SHA-256 y nunca aparece en las respuestas.

### Auth API (8001)

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/health` | Estado de la API |
| POST | `/register` | Registrar credenciales `{username, password}` |
| POST | `/login` | Validar credenciales. Responde 200 `{valid: true}` o 401 `{valid: false}` |
| GET | `/logins` | Historial de intentos de login |

Usuario de prueba creado al iniciar: **admin / admin123**.

### Orders API (8002)

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/` | Estado de la API |
| POST | `/orders` | Crear pedido `{customer, product, quantity, total}` (inicia en `pending`) |
| GET | `/orders` | Listar pedidos |
| GET | `/orders/{id}` | Consultar pedido por ID |
| PUT | `/orders/{id}/status` | Cambiar estado `{status}`: `pending`, `shipped`, `delivered`, `cancelled` |
| DELETE | `/orders/{id}` | Eliminar pedido |

### Products API (8003)

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/` | Estado de la API |
| POST | `/products` | Registrar producto `{name, description, price, stock}` |
| GET | `/products` | Listar productos |
| GET | `/products/{id}` | Consultar producto por ID |
| PUT | `/products/{id}` | Actualizar producto (solo los campos enviados) |
| DELETE | `/products/{id}` | Eliminar producto |

## Pruebas

### Desde PowerShell (Windows)

```powershell
# ---------- Users ----------
Invoke-RestMethod -Method Post -Uri http://localhost:8000/users -ContentType "application/json" -Body '{"name":"Erick","email":"erick@test.com","password":"1234"}'
Invoke-RestMethod http://localhost:8000/users
Invoke-RestMethod http://localhost:8000/users/1
Invoke-RestMethod -Method Put -Uri http://localhost:8000/users/1 -ContentType "application/json" -Body '{"name":"Erick Alvaro"}'
Invoke-RestMethod -Method Delete -Uri http://localhost:8000/users/1

# ---------- Auth ----------
Invoke-RestMethod http://localhost:8001/health
Invoke-RestMethod -Method Post -Uri http://localhost:8001/login -ContentType "application/json" -Body '{"username":"admin","password":"admin123"}'
Invoke-RestMethod -Method Post -Uri http://localhost:8001/login -ContentType "application/json" -Body '{"username":"admin","password":"mala"}'
Invoke-RestMethod -Method Post -Uri http://localhost:8001/register -ContentType "application/json" -Body '{"username":"erick","password":"1234"}'
Invoke-RestMethod http://localhost:8001/logins

# ---------- Orders ----------
Invoke-RestMethod -Method Post -Uri http://localhost:8002/orders -ContentType "application/json" -Body '{"customer":"Erick","product":"Laptop","quantity":1,"total":2500}'
Invoke-RestMethod http://localhost:8002/orders
Invoke-RestMethod http://localhost:8002/orders/1
Invoke-RestMethod -Method Put -Uri http://localhost:8002/orders/1/status -ContentType "application/json" -Body '{"status":"shipped"}'
Invoke-RestMethod -Method Delete -Uri http://localhost:8002/orders/1

# ---------- Products ----------
Invoke-RestMethod -Method Post -Uri http://localhost:8003/products -ContentType "application/json" -Body '{"name":"Laptop","description":"Laptop gamer","price":2500,"stock":10}'
Invoke-RestMethod http://localhost:8003/products
Invoke-RestMethod http://localhost:8003/products/1
Invoke-RestMethod -Method Put -Uri http://localhost:8003/products/1 -ContentType "application/json" -Body '{"price":2300,"stock":8}'
Invoke-RestMethod -Method Delete -Uri http://localhost:8003/products/1
```

> En PowerShell, un login inválido aparece en rojo. Es lo esperado, porque la API responde `401 Unauthorized`.

### Desde Linux, Mac o Git Bash (curl)

```bash
# Users
curl -X POST localhost:8000/users -H "Content-Type: application/json" -d '{"name":"Erick","email":"erick@test.com","password":"1234"}'
curl localhost:8000/users

# Auth
curl -X POST localhost:8001/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'
curl localhost:8001/logins

# Orders
curl -X POST localhost:8002/orders -H "Content-Type: application/json" -d '{"customer":"Erick","product":"Laptop","quantity":1,"total":2500}'
curl -X PUT localhost:8002/orders/1/status -H "Content-Type: application/json" -d '{"status":"shipped"}'
curl localhost:8002/orders

# Products
curl -X POST localhost:8003/products -H "Content-Type: application/json" -d '{"name":"Laptop","price":2500,"stock":10}'
curl localhost:8003/products
```

## Prueba de persistencia

1. Levanta los servicios y crea datos con los comandos de la sección anterior.
2. Elimina los contenedores:
   ```bash
   docker compose down
   ```
3. Confirma que los volúmenes siguen existiendo:
   ```bash
   docker volume ls
   ```
   Deberían aparecer `microservicios_users_data`, `microservicios_auth_data`, `microservicios_orders_data` y `microservicios_products_data`.
4. Levanta todo de nuevo:
   ```bash
   docker compose up -d
   ```
5. Consulta otra vez (por ejemplo `GET /users`, `/logins`, `/orders`, `/products`). Los datos siguen ahí.

### Ver el contenido de una base de datos

Para comprobar que cada API usa solo su propio archivo:

```bash
docker compose exec users-api ls -l /app/data      # solo users.db
docker compose exec products-api ls -l /app/data   # solo products.db
```
