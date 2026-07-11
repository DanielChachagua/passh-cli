# passh CLI 🚀

`passh` es una interfaz de línea de comandos (CLI) moderna y segura desarrollada en Go para centralizar y administrar tus conexiones SSH, credenciales cifradas y accesos compartidos en equipos de trabajo mediante grupos.

## Características Clave
* **Arquitectura Hexagonal (Puertos y Adaptadores)**: Código altamente desacoplado, modular y fácil de mantener.
* **Seguridad de Vanguardia**: Hashing de contraseñas de usuario con **Argon2id** (especificación OWASP) y almacenamiento de credenciales cifradas simétricamente con **AES-256-GCM** en el backend.
* **Conexión SSH Interactiva Nativa**: Interactúa y abre túneles SSH nativos directamente con Go (`golang.org/x/crypto/ssh`) restaurando la terminal en modo Raw para capturar redimensionamientos de ventana, tabulaciones y atajos del sistema.
* **Almacén de Contraseñas Independiente**: Guarda contraseñas de otros sitios de forma encriptada y descífralas interactiva y puntualmente sobre canales seguros.
* **Gestión de Grupos y Credenciales Compartidas**: Crea grupos de trabajo, añade miembros y comparte accesos SSH o contraseñas cifradas para que tus colaboradores puedan usarlos sin necesidad de conocer las credenciales en texto plano.

---

## 💻 Instalación

### Método 1: Instalación Automatizada (Recomendado)
Puedes descargar e instalar la última versión compilada directamente desde los Releases de GitHub indicando la URL de tu API como argumento:

```bash
curl -sSL https://raw.githubusercontent.com/DanielChachagua/passh/main/install.sh | bash -s -- "http://localhost:3200"
```
*Este instalador detecta automáticamente tu arquitectura (amd64, arm64), mueve el binario a `/usr/bin/passh` y configura la variable de entorno `API_URL` en tu perfil de consola (`.bashrc` o `.zshrc`).*

*Nota: Al finalizar la instalación, recuerda recargar la sesión actual de tu consola con:*
```bash
source ~/.bashrc # o source ~/.zshrc
```

---

### Método 2: Instalación Local (Usando Makefile)
Si deseas compilar la CLI localmente desde el código fuente, asegúrate de tener instalado **Go (1.21 o superior)**:

```bash
# Compilar el binario localmente
make build

# Instalar el binario globalmente en /usr/bin
sudo make install
```

---

### Método 3: Ejecución con Docker
Si prefieres no instalar binarios directamente en tu sistema host, puedes compilar la imagen local e interactuar a través de un contenedor:

```bash
# Compilar la imagen
docker build -t passh .

# Ejecutar pasándole la variable de entorno de tu API
docker run -it -e API_URL=http://host.docker.internal:3200 passh user login
```

---

## 🛠️ Guía de Uso

Todos los comandos están organizados en subcomandos modulares:

### 1. Gestión de Cuentas y Seguridad (`passh user ...`)
Gestión del registro, inicio de sesión y restablecimiento de credenciales:
* **Registrar un Usuario** *(Admin JWT requerido)*:
  ```bash
  passh user register --email usuario_nuevo@correo.com
  ```
* **Iniciar Sesión**:
  ```bash
  passh user login
  ```
* **Cerrar Sesión** *(Limpia el token local)*:
  ```bash
  passh user logout
  ```
* **Solicitar Token de Restablecimiento** *(Envía código por email)*:
  ```bash
  passh user reset-request --email tu_cuenta@correo.com
  ```
* **Confirmar Restablecimiento**:
  ```bash
  passh user reset-confirm --token "TOKEN_HEX"
  ```

---

### 2. Conexiones SSH (`passh ssh ...`)
Registra y conéctate de forma nativa e interactiva a tus servidores remotos:
* **Añadir una Conexión**:
  ```bash
  passh ssh add
  ```
  *(Puedes pasar parámetros mediante banderas o interactuar con el asistente si las omites).*
* **Listar Conexiones Disponibles**:
  ```bash
  passh ssh list
  ```
* **Editar una Conexión**:
  ```bash
  passh ssh edit [connection_id]
  ```
* **Eliminar una Conexión**:
  ```bash
  passh ssh delete [connection_id]
  ```
* **Conectarse a un Servidor (Interactiva Paginada)**:
  Muestra una lista interactiva de servidores con soporte de paginación (grupos de 5) e inyección virtual de opciones anterior/siguiente:
  ```bash
  passh ssh connect
  ```
* **Conexión Directa por ID**:
  ```bash
  passh ssh connect 1
  ```
* **Conexión por Búsqueda de Nombre (Filtro Proximidad %like%)**:
  Filtra la lista por coincidencias en el nombre, IP o detalles. Si encuentra una única coincidencia se conectará directamente de forma automática:
  ```bash
  passh ssh connect local
  ```

---

### 3. Almacén de Contraseñas (`passh pass ...`)
Guarda contraseñas o llaves de servicios de forma encriptada:
* **Guardar una Contraseña**:
  ```bash
  passh pass add
  ```
* **Listar Contraseñas Guardadas** *(Contraseñas enmascaradas)*:
  ```bash
  passh pass list
  ```
* **Ver / Descifrar Contraseña**:
  Muestra los detalles del registro descifrando la contraseña en texto plano:
  ```bash
  passh pass view [password_id]
  ```
* **Editar Registro**:
  ```bash
  passh pass edit [password_id]
  ```
* **Eliminar Registro**:
  ```bash
  passh pass delete [password_id]
  ```

---

### 4. Módulo de Grupos y Uso Compartido (`passh group ...`)
Administra accesos y comparte credenciales en equipos de desarrollo de forma granular:
* **Crear un Grupo**:
  ```bash
  passh group create --name "trabajo"
  ```
* **Listar mis Grupos**:
  ```bash
  passh group list
  ```
* **Ver Detalles del Grupo**:
  Visualiza miembros del grupo con sus roles, además de las conexiones y passwords compartidos en el mismo:
  ```bash
  passh group view [group_id]
  ```
* **Agregar Miembro al Grupo** *(Solo creador)*:
  ```bash
  passh group add-member [group_id] [email_miembro]
  ```
* **Eliminar Miembro del Grupo** *(Solo creador)*:
  ```bash
  passh group remove-member [group_id] [user_id_miembro]
  ```
* **Compartir Conexión SSH**:
  Permite que cualquier miembro del grupo liste e inicie sesión en el servidor remoto:
  ```bash
  passh group share-ssh [group_id] [connection_id]
  ```
* **Retirar Conexión SSH del Grupo**:
  ```bash
  passh group unshare-ssh [group_id] [connection_id]
  ```
* **Compartir Contraseña**:
  Permite que cualquier miembro del grupo visualice y descifre la contraseña compartida:
  ```bash
  passh group share-pass [group_id] [password_id]
  ```
* **Retirar Contraseña del Grupo**:
  ```bash
  passh group unshare-pass [group_id] [password_id]
  ```
* **Eliminar un Grupo** *(Solo creador - borra membresías y compartidos en cascada)*:
  ```bash
  passh group delete [group_id]
  ```

---

## 🛠️ Desarrollo

Si quieres contribuir o realizar cambios:
1. Instala dependencias del proyecto:
   ```bash
   go mod download
   ```
2. Ejecuta las pruebas:
   ```bash
   make test
   ```
3. Compila el ejecutable con soporte de pruebas locales:
   ```bash
   go build -o passh main.go
   ```
