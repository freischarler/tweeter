
## Instalación

1. Clona el repositorio:

```sh
git clone https://github.com/freischarler/tweeter.git
cd tweeter
```

2. Configura las variables de entorno para Redis:

```sh
export REDIS_HOST=localhost:6379
 export REDIS_PASSWORD=yourpassword
export PORT=8080
```

3. Ejecuta la aplicación:

```sh
go run cmd/server/main.go
go run cmd/server/main.go
```

## Uso con Docker

1. Asegúrate de tener Docker y Docker Compose instalados en tu máquina.

2. Construye y ejecuta los contenedores con Docker Compose:

```sh
docker-compose up --build
```

3. Para detener y eliminar los contenedores, volúmenes e imágenes de Docker:
```sh
docker-compose down --volumes --rmi all
```

## Endpoints

### Publicar un Tweet

- **URL**: `/tweet`
- **Método**: `POST`
- **Parámetros**:
  - [userID](http://_vscodecontentref_/2): ID del usuario que publica el tweet.
  - [tweet](http://_vscodecontentref_/3): Contenido del tweet.
- **Respuesta**:
  - `200 OK`: Tweet publicado exitosamente.
  - `400 Bad Request`: El tweet excede la longitud máxima.

### Seguir a un Usuario

- **URL**: `/follow`
- **Método**: `POST`
- **Parámetros**:
  - [followerID](http://_vscodecontentref_/4): ID del usuario que sigue.
  - [followeeID](http://_vscodecontentref_/5): ID del usuario a seguir.
- **Respuesta**:
  - `200 OK`: Seguido exitosamente.
  - `400 Bad Request`: No puedes seguirte a ti mismo.

### Ver Timeline

- **URL**: `/timeline/:userID`
- **Método**: `GET`
- **Parámetros**:
  - [userID](http://_vscodecontentref_/6): ID del usuario cuyo timeline se quiere ver.
- **Respuesta**:
  - `200 OK`: Lista de tweets en el timeline.
  - `500 Internal Server Error`: Error al obtener la lista de seguidos.

## Ejemplo de Uso

### Publicar un Tweet

```sh
curl -X POST http://localhost:8080/tweet -d "userID=1" -d "tweet=Hola Mundo"
```

### Seguir a un Usuario

```sh
curl -X POST http://localhost:8080/follow -d "followerID=1" -d "followeeID=2"
```

### Ver Timeline

```sh
curl http://localhost:8080/timeline/1
```
