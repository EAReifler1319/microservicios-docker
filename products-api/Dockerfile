# Imagen base de Node.js (alpine = versión mínima). Node 24 trae SQLite incluido (node:sqlite)
FROM node:24-alpine

WORKDIR /app

# Instalar dependencias
COPY package.json .
RUN npm install --omit=dev

# Copiar el código
COPY main.js .

# Carpeta donde vive la base de datos (aquí se monta el volumen)
ENV DB_PATH=/app/data/products.db
RUN mkdir -p /app/data

EXPOSE 8003

CMD ["node", "main.js"]
