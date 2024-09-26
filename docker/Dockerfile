FROM golang:1.21

WORKDIR /api

# Copiar go.mod y go.sum y descargar dependencias
COPY ./api/go.mod ./api/go.sum ./
RUN go mod download

# Copiar el resto de los archivos
COPY ./api .

# Compilar el binario
RUN go build -o search-api ./main.go

# Verificar si el archivo binario existe
RUN ls -la

# Otorgar permisos de ejecución al binario
RUN chmod +x search-api

# Exponer el puerto 8081
EXPOSE 8081

# Ejecutar el binario compilado
CMD ["./search-api"]
