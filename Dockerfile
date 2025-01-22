# Usar la imagen oficial de Golang
FROM golang:1.20

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos al contenedor
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Construir la aplicación
RUN go build -o main .

# Exponer el puerto de la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
