# Imagen base de Python (alpine = versión mínima)
FROM python:3.12-alpine

WORKDIR /app

# Instalar dependencias
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar el código
COPY main.py .

# Carpeta donde vive la base de datos (aquí se monta el volumen)
ENV DB_PATH=/app/data/users.db
RUN mkdir -p /app/data

EXPOSE 8000

# uvicorn es el servidor que ejecuta la app FastAPI ("main:app" = archivo main.py, variable app)
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
