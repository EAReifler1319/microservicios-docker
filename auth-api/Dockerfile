# ---------- Etapa 1: compilar ----------
# Imagen con Maven + Java 21 para compilar el proyecto
FROM maven:3.9-eclipse-temurin-21 AS build

WORKDIR /app

# Primero el pom.xml y descargar dependencias (Docker lo guarda en caché)
COPY pom.xml .
RUN mvn -q dependency:go-offline

# Luego el código y compilar -> genera target/auth-api.jar
COPY src ./src
RUN mvn -q package

# ---------- Etapa 2: ejecutar ----------
# Imagen solo con Java (JRE), sin Maven -> más liviana
FROM eclipse-temurin:21-jre

WORKDIR /app
COPY --from=build /app/target/auth-api.jar .

# Carpeta donde vive la base de datos (aquí se monta el volumen)
ENV DB_PATH=/app/data/auth.db
RUN mkdir -p /app/data

EXPOSE 8001

# -Xmx128m limita la memoria de Java (importante para el plan gratuito de AWS)
CMD ["java", "-Xmx128m", "-jar", "auth-api.jar"]
