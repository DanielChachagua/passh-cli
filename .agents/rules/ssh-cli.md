---
trigger: always_on
---

Reglas de Programación y Arquitectura: CLI App (Go + Cobra)
1. Arquitectura y Estructura de Archivos
Patrón: Arquitectura Hexagonal adaptada a herramientas de consola.

Separación: El backend de la CLI procesa datos (internal/app), el puerto define la comunicación con servicios externos (internal/ports), y el adaptador (internal/adapters/cli/cobra) gestiona la interfaz gráfica y de comandos de la terminal.

2. Gestión de Variables de Entorno y Configuración Local
Inicialización: La aplicación lee un archivo local de configuración (ej. en ~/.ssh-manager/config.json) o variables de entorno del sistema al arrancar.

Valores por Defecto: Si la variable API_URL no está definida, se asume por defecto http://localhost:8080.

3. Manejo Estandarizado de Errores de la API
Validación de Respuesta: La CLI siempre debe inspeccionar el campo success del JSON de la API antes de mapear la estructura de datos.

Manejo del Token: Si la API devuelve un error con código AUTH_FAILED o un código 401 Unauthorized, la CLI debe limpiar el archivo de configuración local, notificar al usuario que su sesión expiró y solicitar la ejecución de ssh-manager login de nuevo de forma clara.

4. Interacción y Flujo SSH Automático
Interfaz de Usuario: Uso de manifoldco/promptui para las listas interactivas navegables con teclado.

Conexión SSH: No se permite el uso de dependencias del sistema operativo como sshpass. La CLI debe establecer el túnel e interactuar utilizando de forma nativa la librería oficial golang.org/x/crypto/ssh.

Sesión Interactiva: Al concretarse la conexión, la CLI debe ligar directamente el flujo hacia los descriptores del sistema para ceder la terminal:

Go
session.Stdout = os.Stdout
session.Stdin = os.Stdin
session.Stderr = os.Stderr