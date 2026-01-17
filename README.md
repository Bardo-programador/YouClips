# YouClips API

API para criar clips de vídeos do YouTube feito em Golang.

## Pré-requisitos

- Go 1.25+
- yt-dlp instalado no sistema
- SQLite3

## Instalação

```bash
# Instalar dependências
go mod tidy

# Compilar (opcional)
go build -o bin/api ./cmd/api
```

## Executar

```bash
# Usar configurações padrão
./bin/api

# se quiser rodar sem compilar  
go run cmd/api/main.go

# Ou configurar via variáveis de ambiente
DB_PATH=./data/app.db STORAGE_DIR=./data/clips ./bin/api
```

## Endpoints

### POST /clips
Cria um novo clip.

**Request:**

Video:
```json
{
  "url": "https://youtube.com/watch?v=...",
  "start_time": 10,
  "end_time": 30,
  "format": "video"
}
```

Audio
```json
{
  "url": "https://youtube.com/watch?v=...",
  "start_time": 10,
  "end_time": 30,
  "format": "audio"
}
```


**Response:**
```json
{
  "id": 1,
  "status": "processing"
}
```

### GET /clips/{id}
Obtém informações de um clip.

**Response:**
```json
{
  "id": 1,
  "status": "completed",
  "title": "Video Title",
  "duration": 20,
  "size": 1234567,
  "download_url": "/clips/1/download"
}
```

### GET /clips/{id}/download
Baixa o arquivo do clip.

### GET /clips?page=1&limit=20
Lista clips com paginação.

**Response:**
```json
{
  "clips": [...],
  "total": 50,
  "page": 1
}
```

## Estrutura do Projeto

```
.
├── cmd/api/              # Aplicação principal
├── internal/
│   ├── controller/       # Handlers HTTP
│   ├── entities/         # Modelos de dados
│   ├── repository/       # Acesso ao banco
│   └── service/          # Lógica de negócio
├── storage/clips/        # Clips salvos
└── youclips.db           # Banco SQLite
```
