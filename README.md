# YouClips API

API para criar clips de vídeos do YouTube feito em Golang.

## Pré-requisitos

- Go 1.25+
- yt-dlp instalado no sistema
- ffmpeg instalado no sistema (usado pelo yt-dlp e para conversão de áudio)
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

### POST /metadata
Obtém metadados de um vídeo do YouTube (título e duração).

**Request:**
```json
{
  "url": "https://youtube.com/watch?v=..."
}
```

**Response:**
```json
{
  "title": "Rick Astley - Never Gonna Give You Up (Official Video) (4K Remaster)",
  "duration": 213
}
```

**Notas:**
- `duration` está em segundos
- Usado pelo frontend para exibir informações antes de criar o clip
- Valida se a URL é válida e acessível

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

**Parâmetros query:**
- `page` (opcional): Número da página (padrão: 1)
- `limit` (opcional): Itens por página (padrão: 20, máx: 100)

**Response:**
```json
{
  "clips": [
    {
      "id": 1,
      "status": "completed",
      "title": "Video Title",
      "duration": 20,
      "size": 1234567,
      "download_url": "/clips/1/download"
    }
  ],
  "total": 50,
  "page": 1
}
```

### DELETE /clips/{id}
Apaga um clip permanentemente (remove arquivo do disco e registro do banco).

**Response:**
```json
{
  "deleted": true
}
```

**Notas:**
- Remove o arquivo físico do disco (storage/clips/)
- Remove o registro do banco de dados
- Operação irreversível
- Retorna 404 se o clip não existir

## Status dos Clips

Os clips podem ter os seguintes status:
- `processing` - Clip sendo processado pelo yt-dlp
- `completed` - Clip pronto para download
- `failed` - Erro no processamento

## Fluxo de criação de clip

1. **Buscar metadados** (opcional, mas recomendado):
   ```bash
   POST /metadata
   ```
   Retorna título e duração do vídeo

2. **Criar clip**:
   ```bash
   POST /clips
   ```
   Inicia processamento assíncrono, retorna ID e status

3. **Monitorar status**:
   ```bash
   GET /clips/{id}
   ```
   Verificar se status mudou para `completed`

4. **Baixar arquivo**:
   ```bash
   GET /clips/{id}/download
   ```
   Download direto do arquivo MP4 ou MP3

## Frontend

O frontend foi desenvolvido com SvelteKit e está na pasta `web/`.

### Executar o frontend

```bash
cd web
npm run dev
```

O frontend estará disponível em `http://localhost:5173` e fará proxy das requisições `/api/*` para a API em `localhost:8080`.

**Rotas disponíveis:**
- `/` - Criar novo clip
- `/clips` - Listar todos os clips
- `/clips/[id]` - Detalhes de um clip específico

**Funcionalidades:**
- Buscar metadados do vídeo antes de criar clip
- Criar clips de vídeo ou áudio
- Proteção contra duplicação de clips (botão desabilitado durante criação)
- Listar clips com auto-refresh para clips em processamento
- Download de clips concluídos
- Apagar clips com confirmação

Ver `web/ESTRUTURA.md` para documentação completa do frontend.

## Exemplos de uso da API

### Buscar metadados do vídeo
```bash
curl -X POST http://localhost:8080/metadata \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
```

### Criar um clip de vídeo
```bash
curl -X POST http://localhost:8080/clips \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "start_time": 0,
    "end_time": 30,
    "format": "video"
  }'
```

### Verificar status do clip
```bash
curl http://localhost:8080/clips/1
```

### Listar todos os clips
```bash
curl http://localhost:8080/clips?page=1&limit=10
```

### Apagar um clip
```bash
curl -X DELETE http://localhost:8080/clips/1
```

### Baixar clip
```bash
curl -o meu-clip.mp4 http://localhost:8080/clips/1/download
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
├── web/                  # Frontend SvelteKit
└── youclips.db           # Banco SQLite
```
