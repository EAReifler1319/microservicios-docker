# ---------- Etapa 1: compilar ----------
# Imagen base de Go con todas las herramientas para compilar
FROM golang:1.25-alpine AS build

WORKDIR /app

# Copiar dependencias y código
COPY go.mod main.go ./

# go mod tidy descarga las dependencias y genera go.sum
# CGO_ENABLED=0 crea un binario independiente (el driver SQLite es Go puro)
RUN go mod tidy && CGO_ENABLED=0 go build -o orders-api .

# ---------- Etapa 2: ejecutar ----------
# Imagen mínima: solo lleva el binario ya compilado (~15 MB)
FROM alpine:3.22

WORKDIR /app
COPY --from=build /app/orders-api .

# Carpeta donde vive la base de datos (aquí se monta el volumen)
ENV DB_PATH=/app/data/orders.db
RUN mkdir -p /app/data

EXPOSE 8002

CMD ["./orders-api"]
