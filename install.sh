#!/usr/bin/env bash

# Evitar continuar si ocurre un error
set -e

# Colores para la consola
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # Sin color

# Configuración del repositorio
GITHUB_REPO="DanielChachagua/passh"
BINARY_NAME="passh"
DEST_DIR="/usr/bin"

echo -e "${BLUE}===============================================${NC}"
echo -e "${BLUE}        Instalador de passh CLI (Linux)        ${NC}"
echo -e "${BLUE}===============================================${NC}"

# 1. Validar que se haya provisto la URL de la API como argumento obligatorio
if [ -z "$1" ]; then
    echo -e "${RED}Error: La URL de la API (API_URL) es obligatoria para la instalación.${NC}"
    echo -e "Debes indicar a qué servidor API se conectará passh."
    echo -e "\n${YELLOW}Ejemplo de instalación remota:${NC}"
    echo -e "  ${BLUE}curl -sSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/install.sh | bash -s -- \"http://localhost:3200\"${NC}"
    echo -e "\n${YELLOW}Ejemplo de instalación local:${NC}"
    echo -e "  ${BLUE}./install.sh \"http://localhost:3200\"${NC}\n"
    exit 1
fi

CUSTOM_API_URL="$1"

# 2. Validar Sistema Operativo
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ]; then
    echo -e "${RED}Error: Este instalador solo es compatible con Linux.${NC}"
    exit 1
fi

# 3. Detectar Arquitectura
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        BIN_ARCH="amd64"
        ;;
    aarch64|arm64)
        BIN_ARCH="arm64"
        ;;
    i386|i686)
        BIN_ARCH="386"
        ;;
    *)
        echo -e "${RED}Error: Arquitectura de CPU no soportada (${ARCH}).${NC}"
        exit 1
        ;;
esac

echo -e "Sistema: ${GREEN}Linux${NC} (${ARCH})"

# 4. Obtener la última versión desde GitHub Releases
echo -e "${YELLOW}Buscando la última versión en GitHub...${NC}"

# Obtener tag de la última versión
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${YELLOW}Advertencia: No se pudo determinar la última versión desde GitHub (¿repositorio privado o sin releases?).${NC}"
    echo -e "Usando versión por defecto: ${BLUE}v1.0.0${NC}"
    LATEST_RELEASE="v1.0.0"
else
    echo -e "Versión encontrada: ${GREEN}${LATEST_RELEASE}${NC}"
fi

# Construir URL de descarga del binario precompilado
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}-linux-${BIN_ARCH}"

# Si el repositorio no está configurado, avisamos al usuario
if [[ "$GITHUB_REPO" == "usuario/passh" ]]; then
    echo -e "\n${YELLOW}[!] NOTA: El script está usando el repositorio de ejemplo 'usuario/passh'.${NC}"
    echo -e "Para usarlo de forma real, asegúrate de reemplazar GITHUB_REPO con tu repositorio en este script."
    echo -e "Presiona Ctrl+C para cancelar, o Enter para continuar intentando descargar...${NC}"
    read -r
fi

# 5. Descargar el binario
TEMP_FILE="/tmp/${BINARY_NAME}"
echo -e "${YELLOW}Descargando binario desde:${NC}"
echo -e "   ${BLUE}${DOWNLOAD_URL}${NC}"

# Descargar usando curl o wget
if command -v curl &> /dev/null; then
    curl -L -f -o "$TEMP_FILE" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -q -O "$TEMP_FILE" "$DOWNLOAD_URL"
else
    echo -e "${RED}Error: Se requiere 'curl' o 'wget' para descargar el binario.${NC}"
    exit 1
fi

echo -e "${GREEN}Descarga completada con éxito.${NC}"

# 6. Instalar en /usr/bin
echo -e "${YELLOW}Instalando el ejecutable en ${DEST_DIR}/${BINARY_NAME}...${NC}"

# Verificar si se necesitan privilegios de administrador
if [ -w "$DEST_DIR" ]; then
    mv "$TEMP_FILE" "${DEST_DIR}/${BINARY_NAME}"
    chmod +x "${DEST_DIR}/${BINARY_NAME}"
else
    echo -e "${YELLOW}Se requieren permisos de administrador (sudo) para escribir en ${DEST_DIR}.${NC}"
    sudo mv "$TEMP_FILE" "${DEST_DIR}/${BINARY_NAME}"
    sudo chmod +x "${DEST_DIR}/${BINARY_NAME}"
fi

# 7. Configurar la variable de entorno API_URL en el sistema
echo -e "Usando API_URL: ${GREEN}${CUSTOM_API_URL}${NC}"

SHELL_RC=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_RC="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_RC="$HOME/.zshrc"
fi

if [ -n "$SHELL_RC" ]; then
    if grep -q "API_URL=" "$SHELL_RC"; then
        echo -e "${YELLOW}Actualizando la variable API_URL existente en $(basename "$SHELL_RC")...${NC}"
        # Borrar configuración vieja usando sed de forma portable
        sed -i.bak "/API_URL=/d" "$SHELL_RC" && rm -f "${SHELL_RC}.bak"
        echo "export API_URL=\"$CUSTOM_API_URL\"" >> "$SHELL_RC"
        echo -e "${GREEN}¡API_URL actualizada con éxito en $(basename "$SHELL_RC")!${NC}"
    else
        echo -e "${YELLOW}Configurando API_URL (${CUSTOM_API_URL}) en $(basename "$SHELL_RC")...${NC}"
        echo "export API_URL=\"$CUSTOM_API_URL\"" >> "$SHELL_RC"
        echo -e "${GREEN}¡API_URL configurada con éxito en $(basename "$SHELL_RC")!${NC}"
    fi
fi

# 8. Validar instalación
if command -v "$BINARY_NAME" &> /dev/null; then
    echo -e "\n${GREEN}===============================================${NC}"
    echo -e "${GREEN}      ¡passh CLI se ha instalado con éxito!    ${NC}"
    echo -e "${GREEN}===============================================${NC}"
    echo -e "Puedes ejecutar la herramienta desde cualquier parte usando:"
    echo -e "  ${BLUE}passh user login${NC}"
    echo -e "  ${BLUE}passh --help${NC}"
    
    if [ -n "$SHELL_RC" ]; then
        echo -e "\nPor favor, recarga tu terminal actual para aplicar la variable de entorno:"
        echo -e "  ${BLUE}source $(basename "$SHELL_RC")${NC}"
    fi
else
    echo -e "\n${RED}Error: El binario fue copiado pero no se pudo encontrar en el PATH.${NC}"
    echo -e "Por favor, asegúrate de que '${DEST_DIR}' está en tu variable de entorno PATH.${NC}"
    exit 1
fi
