# Imagen base de Python (versión liviana)
FROM python:3.12-slim

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

# fastapi run = modo producción (fastapi dev es solo para desarrollo local)
CMD ["fastapi", "run", "main.py", "--port", "8000"]
