# Comandos de Prueba para `passh`

Esta es una lista rápida de comandos que puedes ejecutar para probar la aplicación de línea de comandos conectada al servidor local de la API (`http://localhost:3200`).

---

## 1. Servidor de la API (Backend)
Inicia la API si aún no está corriendo:
```bash
cd ../ssh-api && go run cmd/api/main.go
```

---

## 2. Compilar la CLI
Compila el binario `passh` desde la raíz de `ssh-cli`:
```bash
go build -o passh main.go
```

---

## 3. Módulo de Gestión de Cuentas y Seguridad (`passh user ...`)

### A. Registrar un nuevo usuario (Admin JWT requerido)
Registra un nuevo usuario indicando únicamente su email. El backend le creará una contraseña temporal aleatoria y le enviará un correo de bienvenida.
```bash
# Modo Interactivo
API_URL=http://localhost:3200 ./passh user register

# Usando Banderas (Flags)
API_URL=http://localhost:3200 ./passh user register --email nuevo_usuario@ejemplo.com
```

### B. Iniciar Sesión (Login)
```bash
# Modo Interactivo
API_URL=http://localhost:3200 ./passh user login

# Usando Banderas (Flags)
API_URL=http://localhost:3200 ./passh user login --email test@example.com --password mypassword123
```

### C. Solicitar Restablecimiento de Contraseña (Reset Token)
```bash
# Modo Interactivo
API_URL=http://localhost:3200 ./passh user reset-request

# Usando Banderas (Flags)
API_URL=http://localhost:3200 ./passh user reset-request --email test@example.com
```

### D. Confirmar Restablecimiento de Contraseña
```bash
# Modo Interactivo
API_URL=http://localhost:3200 ./passh user reset-confirm

# Usando Banderas (Flags)
API_URL=http://localhost:3200 ./passh user reset-confirm --token "TU_TOKEN_HEX" --new-password "NuevaPass123!" --confirm-password "NuevaPass123!"
```

### E. Cerrar Sesión (Logout)
```bash
API_URL=http://localhost:3200 ./passh user logout
```

---

## 4. Módulo de Conexiones SSH (`passh ssh ...`)

### A. Crear una Conexión SSH
```bash
API_URL=http://localhost:3200 ./passh ssh add
```

### B. Listar Conexiones SSH
```bash
API_URL=http://localhost:3200 ./passh ssh list
```

### C. Editar Conexión SSH (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh ssh edit 1
```

### D. Eliminar Conexión SSH (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh ssh delete 1
```

### E. Conexión Interactiva General (Con paginación)
```bash
API_URL=http://localhost:3200 ./passh ssh connect
```

### F. Conexión Directa por ID
```bash
API_URL=http://localhost:3200 ./passh ssh connect 1
```

### G. Conexión por Proximidad de Nombre (%like%)
```bash
API_URL=http://localhost:3200 ./passh ssh connect local
```

---

## 5. Módulo de Contraseñas Seguras (`passh pass ...`)

### A. Guardar una nueva Contraseña
```bash
API_URL=http://localhost:3200 ./passh pass add
```

### B. Listar Contraseñas Guardadas
```bash
API_URL=http://localhost:3200 ./passh pass list
```

### C. Consultar / Descifrar Contraseña (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh pass view 1
```

### D. Editar una Credencial (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh pass edit 1
```

### E. Eliminar una Credencial (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh pass delete 1
```

---

## 6. Módulo de Gestión de Grupos (`passh group ...`)

### A. Crear un Grupo
```bash
# Modo Interactivo
API_URL=http://localhost:3200 ./passh group create

# Usando banderas
API_URL=http://localhost:3200 ./passh group create --name trabajo
```

### B. Listar mis Grupos
```bash
API_URL=http://localhost:3200 ./passh group list
```

### C. Ver Detalles del Grupo (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh group view 1
```

### D. Agregar Miembro al Grupo (ID Grupo: 1)
```bash
# Formato: ./passh group add-member [id_grupo] [email_miembro]
API_URL=http://localhost:3200 ./passh group add-member 1 miembro@passhapi.local
```

### E. Eliminar Miembro del Grupo (ID Grupo: 1)
```bash
# Formato: ./passh group remove-member [id_grupo] [id_usuario_miembro]
API_URL=http://localhost:3200 ./passh group remove-member 1 3
```

### F. Compartir Conexión SSH (ID Grupo: 1, ID Conexión: 5)
```bash
# Formato: ./passh group share-ssh [id_grupo] [id_conexion]
API_URL=http://localhost:3200 ./passh group share-ssh 1 5
```

### G. Dejar de Compartir Conexión SSH (ID Grupo: 1, ID Conexión: 5)
```bash
# Formato: ./passh group unshare-ssh [id_grupo] [id_conexion]
API_URL=http://localhost:3200 ./passh group unshare-ssh 1 5
```

### H. Compartir Contraseña (ID Grupo: 1, ID Contraseña: 8)
```bash
# Formato: ./passh group share-pass [id_grupo] [id_password]
API_URL=http://localhost:3200 ./passh group share-pass 1 8
```

### I. Dejar de Compartir Contraseña (ID Grupo: 1, ID Contraseña: 8)
```bash
# Formato: ./passh group unshare-pass [id_grupo] [id_password]
API_URL=http://localhost:3200 ./passh group unshare-pass 1 8
```

### J. Eliminar un Grupo (ID: 1)
```bash
API_URL=http://localhost:3200 ./passh group delete 1
```
