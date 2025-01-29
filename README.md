
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

## Log del Docker

Ejemplo de log generado por la aplicación cuando se ejecuta en Docker:

```
api-1  | 2025/01/29 02:22:15 Starting server on port 8080
api-1  | 2025/01/29 02:22:27 Received POST request for /tweet
api-1  | 2025/01/29 02:22:27 Tweet posted successfully: 1
api-1  | 2025/01/29 02:22:34 Received POST request for /tweet
api-1  | 2025/01/29 02:22:34 Tweet posted successfully: 2
api-1  | 2025/01/29 02:22:37 Received POST request for /tweet
api-1  | 2025/01/29 02:22:37 Tweet posted successfully: 3
api-1  | 2025/01/29 02:22:40 Received POST request for /tweet
api-1  | 2025/01/29 02:22:40 Tweet posted successfully: 4
api-1  | 2025/01/29 02:22:43 Received POST request for /tweet
api-1  | 2025/01/29 02:22:43 Tweet posted successfully: 5
api-1  | 2025/01/29 02:22:46 Received POST request for /follow
api-1  | 2025/01/29 02:22:46 User 1 followed user 2 successfully
api-1  | 2025/01/29 02:22:50 Received GET request for /timeline/1
api-1  | 2025/01/29 02:22:50 Fetched timeline for user 1 successfully
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

Ejemplo de respuesta

{
    "message": "Tweet posted successfully",
    "tweetID": "5"
}

### Seguir a un Usuario

```sh
curl -X POST http://localhost:8080/follow -d "followerID=1" -d "followeeID=2"
```

Ejemplo de respuesta

{
    "message": "Followed successfully"
}

### Ver Timeline

```sh
curl http://localhost:8080/timeline/1
```

Ejemplo de respuesta

[
    {
        "userID": "1",
        "content": "holaAAA ?",
        "timestamp": 1738115249
    },
    {
        "userID": "2",
        "content": "holaAAA ",
        "timestamp": 1738115241
    },
    {
        "userID": "1",
        "content": "hola ",
        "timestamp": 1738115231
    }
]

### Pruebas Unitarias
Para ejecutar las pruebas unitarias, usa el siguiente comando:

```
go test ./internal/application -v
```

### Middleware de Limitación de Tasa

Se agrego un middleware que limita el número de solicitudes que un cliente puede hacer en un período de tiempo determinado, ayudando a proteger tu aplicación contra abusos y ataques de denegación de servicio (DoS).

### Uso de Redis

Se eligió Redis como base de datos debido a sus características de alto rendimiento y baja latencia, lo que lo hace ideal para aplicaciones que requieren una gran cantidad de lecturas rápidas. Redis almacena los datos en memoria, lo que permite acceder a ellos de manera extremadamente rápida. Esto es crucial para una aplicación que necesita escalar a millones de usuarios y estar optimizada para lecturas, como es el caso de esta aplicación de tweets.

### Arquitectura Hexagonal en Go

La aplicación está desarrollada en Go utilizando una arquitectura hexagonal. Esta arquitectura se centra en la separación de preocupaciones y la independencia de los componentes, lo que facilita el mantenimiento y la escalabilidad de la aplicación. Las características principales de la arquitectura hexagonal son:

Separación de Capas: La aplicación se divide en capas independientes, como la capa de dominio, la capa de aplicación y la capa de infraestructura. Cada capa tiene responsabilidades claras y bien definidas.

Independencia de la Infraestructura: La lógica de negocio no depende de detalles de infraestructura como bases de datos o servicios externos. Esto permite cambiar la infraestructura sin afectar la lógica de negocio.

Facilidad de Pruebas: La separación de capas y la independencia de la infraestructura facilitan la creación de pruebas unitarias y de integración, lo que mejora la calidad del código.

Escalabilidad y Mantenibilidad: La arquitectura hexagonal facilita la escalabilidad y el mantenimiento de la aplicación, ya que los cambios en una capa no afectan a las demás capas.

### Posibles Actualizaciones o Mejoras

1. Autenticación y Autorización: Implementar un sistema de autenticación y autorización para asegurar que solo usuarios autenticados puedan publicar tweets y seguir a otros usuarios.

2. Microservicios: Dividir la aplicación en microservicios independientes para mejorar la escalabilidad y la mantenibilidad.

3. Notificaciones en Tiempo Real: Implementar notificaciones en tiempo real para alertar a los usuarios cuando reciben nuevos seguidores o tweets.

4. Optimización del Rendimiento: Optimizar el rendimiento de la aplicación mediante el uso de técnicas de caching, balanceo de carga y optimización de consultas.

5. Pruebas Automatizadas: Añadir pruebas unitarias y de integración para asegurar la calidad del código y facilitar el desarrollo continuo.
