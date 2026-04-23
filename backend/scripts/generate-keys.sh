#!/bin/bash
#
# generate-jwt-keys.sh — Генерация JWT-ключей через openssl
#
# Использование:
#   ./generate-jwt-keys.sh [algorithm] [format]
#
# Алгоритмы: hs256, es256, rs256 (по умолчанию: hs256)
# Форматы: plain, env, config (по умолчанию: env)
#

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Параметры по умолчанию
ALGORITHM="${1:-hs256}"
FORMAT="${2:-env}"
OUTPUT_FILE="${3:-}"

# Функция помощи
print_help() {
    cat << EOF
${BLUE}JWT Key Generator (openssl)${NC}

${YELLOW}Использование:${NC}
  $0 [algorithm] [format] [output_file]

${YELLOW}Алгоритмы:${NC}
  hs256   — HMAC-SHA256 (симметричный, простой)
  es256   — ECDSA P-256 (асимметричный, безопасный)
  rs256   — RSA 2048-bit (асимметричный, стандарт OAuth2)

${YELLOW}Форматы:${NC}
  plain   — Только ключи (для просмотра)
  env     — Переменные среды (.env формат)
  config  — YAML/JSON snippet для config.yaml

${YELLOW}Примеры:${NC}
  $0 hs256 env              # Генерация HS256 в .env формате
  $0 es256 env .env.local   # Генерация ES256 и запись в файл
  $0 rs256 config           # Генерация RS256 в YAML формате
  $0 hs256 plain            # Просто вывести секрет

EOF
}

# Проверка наличия openssl
check_openssl() {
    if ! command -v openssl &> /dev/null; then
        echo -e "${RED}❌ Error: openssl not found${NC}"
        echo "Please install openssl:"
        echo "  macOS: brew install openssl"
        echo "  Ubuntu: apt-get install openssl"
        echo "  Windows: choco install openssl"
        exit 1
    fi
}

# Генерация HS256 секрета (32 байта = 256 бит)
generate_hs256() {
    SECRET=$(openssl rand -base64 32 | tr -d '\n')
    echo "$SECRET"
}

# Генерация ES256 ключей (ECDSA P-256)
generate_es256() {
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Генерация приватного ключа
    openssl ecparam -genkey -name prime256v1 -noout -out "$TEMP_DIR/private.key" 2>/dev/null
    
    # Генерация публичного ключа
    openssl ec -in "$TEMP_DIR/private.key" -pubout -out "$TEMP_DIR/public.key" 2>/dev/null
    
    # Вывод
    cat "$TEMP_DIR/private.key"
    echo "---"
    cat "$TEMP_DIR/public.key"
}

# Генерация RS256 ключей (RSA 2048-bit)
generate_rs256() {
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Генерация приватного ключа
    openssl genrsa -out "$TEMP_DIR/private.key" 2048 2>/dev/null
    
    # Генерация публичного ключа
    openssl rsa -in "$TEMP_DIR/private.key" -pubout -out "$TEMP_DIR/public.key" 2>/dev/null
    
    # Вывод
    cat "$TEMP_DIR/private.key"
    echo "---"
    cat "$TEMP_DIR/public.key"
}

# Форматирование вывода
format_output() {
    local algo="$1"
    local format="$2"
    local private_key="$3"
    local public_key="$4"
    
    case "$format" in
        plain)
            if [ "$algo" = "hs256" ]; then
                echo -e "${GREEN}JWT Secret (HS256):${NC}"
                echo "$private_key"
            else
                echo -e "${GREEN}Private Key:${NC}"
                echo "$private_key"
                echo ""
                echo -e "${GREEN}Public Key:${NC}"
                echo "$public_key"
            fi
            ;;
        
        env)
            if [ "$algo" = "hs256" ]; then
                cat << EOF
# JWT Configuration (HS256)
JWT_SECRET=$private_key
JWT_ALGORITHM=HS256
JWT_EXPIRY=24h
EOF
            else
                # Для асимметричных ключей кодируем в base64 для удобства
                local priv_b64=$(echo "$private_key" | base64 -w0)
                local pub_b64=$(echo "$public_key" | base64 -w0)
                cat << EOF
# JWT Configuration ($algo)
JWT_PRIVATE_KEY=$priv_b64
JWT_PUBLIC_KEY=$pub_b64
JWT_ALGORITHM=$(echo "$algo" | tr '[:lower:]' '[:upper:]')
JWT_EXPIRY=24h
EOF
            fi
            ;;
        
        config)
            if [ "$algo" = "hs256" ]; then
                cat << EOF
auth:
  jwt_secret: $private_key
  jwt_algorithm: HS256
  jwt_expiry: 24h
EOF
            else
                # Индентация для YAML
                local priv_indented=$(echo "$private_key" | sed 's/^/    /')
                local pub_indented=$(echo "$public_key" | sed 's/^/    /')
                cat << EOF
auth:
  jwt_private_key: |
$priv_indented
  jwt_public_key: |
$pub_indented
  jwt_algorithm: $(echo "$algo" | tr '[:lower:]' '[:upper:]')
  jwt_expiry: 24h
EOF
            fi
            ;;
        
        *)
            echo -e "${RED}❌ Error: Unknown format '$format'${NC}"
            exit 1
            ;;
    esac
}

# Основная логика
main() {
    # Проверка флагов помощи
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        print_help
        exit 0
    fi
    
    # Проверка openssl
    check_openssl
    
    # Валидация алгоритма
    case "$ALGORITHM" in
        hs256|es256|rs256)
            ;;
        *)
            echo -e "${RED}❌ Error: Unknown algorithm '$ALGORITHM'${NC}"
            echo "Supported: hs256, es256, rs256"
            exit 1
            ;;
    esac
    
    # Валидация формата
    case "$FORMAT" in
        plain|env|config)
            ;;
        *)
            echo -e "${RED}❌ Error: Unknown format '$FORMAT'${NC}"
            echo "Supported: plain, env, config"
            exit 1
            ;;
    esac
    
    echo -e "${BLUE}🔑 Generating JWT keys...${NC}"
    echo -e "Algorithm: ${YELLOW}$ALGORITHM${NC}"
    echo -e "Format: ${YELLOW}$FORMAT${NC}"
    echo ""
    
    # Генерация ключей
    case "$ALGORITHM" in
        hs256)
            KEYS=$(generate_hs256)
            PRIVATE_KEY="$KEYS"
            PUBLIC_KEY=""
            ;;
        es256)
            KEYS=$(generate_es256)
            PRIVATE_KEY=$(echo "$KEYS" | sed -n '1,/---/p' | head -n -1)
            PUBLIC_KEY=$(echo "$KEYS" | sed -n '/---/,$p' | tail -n +2)
            ;;
        rs256)
            KEYS=$(generate_rs256)
            PRIVATE_KEY=$(echo "$KEYS" | sed -n '1,/---/p' | head -n -1)
            PUBLIC_KEY=$(echo "$KEYS" | sed -n '/---/,$p' | tail -n +2)
            ;;
    esac
    
    # Форматирование и вывод
    OUTPUT=$(format_output "$ALGORITHM" "$FORMAT" "$PRIVATE_KEY" "$PUBLIC_KEY")
    
    if [ -n "$OUTPUT_FILE" ]; then
        echo "$OUTPUT" > "$OUTPUT_FILE"
        chmod 600 "$OUTPUT_FILE"  # Только владелец может читать
        echo -e "${GREEN}✅ Keys written to $OUTPUT_FILE${NC}"
        echo -e "${YELLOW}⚠️  Remember: never commit this file to version control!${NC}"
    else
        echo "$OUTPUT"
    fi
}

# Запуск
main "$@"