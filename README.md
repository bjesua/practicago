# README

## Instalando el ambiente

1. Este proyecto usa **Docker** para funcionar, por lo que necesita estar instalado. Se puede descargar e instalar desde [Docker Desktop](https://www.docker.com/products/docker-desktop/) seleccionando el sistema operativo que se está usando en la computadora donde se va a levantar el proyecto.

2. Después de instalar Docker, se debe descargar e instalar **Git** desde [git-scm.com](https://git-scm.com/downloads), seleccionando el sistema operativo en la computadora a usar.

3. Crear una carpeta en "Mis documentos" y abrir una terminal en **Mac** o **Linux**, y una consola en **Windows**. Clonar el proyecto con el siguiente comando:
   ```bash
   git clone https://github.com/bjesua/practicago.git
   ```
   Esto creará una carpeta llamada "practicago".

4. Abrir el programa Docker ya instalado y después entrar en la carpeta "practicago" con el comando:
   ```bash
   cd practicago
   ```
   Luego, correr el siguiente comando para levantar el proyecto:
   ```bash
   docker-compose up --build
   ```

## Probando la solución

5. Descargar e instalar el programa **Postman** para realizar peticiones a endpoints API desde [postman.com](https://www.postman.com/downloads/).

6. Abriendo el programa Postman, pegar en el apartado donde dice "Enter URL", seleccionar petición **GET** y pegar la siguiente URL:
   ```
   http://localhost:8081/api?song=mujeres&artist=ricardo arjona&album=historias nuevas
   ```
   Adicionalmente, en la ventana de headers, colocar como parámetro **Key** = `"API-Key"` (sin comillas) y **Value** = `"Tribal_Token"` (sin comillas).

7. Dar clic en **Send** y verificar la información que está retornando. Un ejemplo de respuesta podría ser:

```json
[
    {
        "id": 1,
        "name": "Mujeres",
        "artist": "Ricardo Arjona",
        "album": "",
        "artwork": "http://www.chartlyrics.com/mXJiQHkJukWJJCr1_cNuqA.aspx",
        "price": 0,
        "origin": "ChartLyrics",
        "duration": "0:00"
    },
    {
        "id": 2,
        "name": "Las Tres Mujeres",
        "artist": "Ricardo Cerda El Gavilan",
        "album": "",
        "artwork": "http://www.chartlyrics.com/pBahQwFNVUmxLlRgyYn7Zg.aspx",
        "price": 0,
        "origin": "ChartLyrics",
        "duration": "0:00"
    }
]

```
