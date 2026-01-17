# Requisitos Funcionais (v1 - MVP)

- Baixar vídeos e áudios do Youtube como clips
- Criar clips de qualquer duração
- Baixar clips criados




# MVP - Criador de clips do Youtube

## Fluxo principal 

1. Usuário cola o URL válido do vídeo do YouTube
2. Usuário determina o intervalo de tempo (início e fim)
3. Usuário escolhe o formato (vídeo ou áudio)
4. Usuário clica em "Criar Clip"
5. Sistema processa o clip
6. Clip fica disponível para download

## Tecnologias adotadas

- Backend: Go por motivos de aprendizados e devido a sua capacidade de concorrência.
- Frontend: a decidir
- Banco de dados: SQLite  
- Processamento de vídeo yt-dlp

## Modelagem relacional (Simplificada para MVP)

Table clip {
  id integer [primary key]
  created_at timestamp
  title string
  start_time integer  // em segundos
  end_time integer    // em segundos
  duration_seconds integer
  format string       // video ou audio
  size bigint
  file_path string
  original_url string
  status clip_status
}

enum clip_status {
  processing
  completed
  failed
} 


## Arquitetura 

┌───────────────┐
│   Frontend    │
└───────▲───────┘
        │ HTTP (REST)
┌───────┴───────┐
│Controller     │  ← Controllers / Handlers
└───────▲───────┘
        │
┌───────┴───────┐
│  Service      │  ← Use cases / Services / Processors 
│   Layer       │
└───────▲───────┘
        │
┌───────┴───────┐
│  Repository   │  ← Persistency
│   Layer       │
└───────▲───────┘
        │
┌───────┴───────┐
│ Entities      │  ← Constants and other defitinions (like Models)
└───────────────┘

Processamento assíncrono: API → Queue → Workers → yt-dlp 

## Armazenamento

- Clips salvos localmente 
- Política de limpeza: sem remoção automática

## Endpoints (MVP)

### Clips 

**POST /clips**
- Cria um novo clip
- Body: `{ "url": string, "start_time": int, "end_time": int, "format": "video" | "audio" }`
- Response: `{ "id": int, "status": "processing" }`

**GET /clips/{id}**
- Obtém informações de um clip específico
- Response: `{ "id": int, "status": string, "title": string, "duration": int, "size": int, "download_url": string }`

**GET /clips/{id}/download**
- Faz download do arquivo do clip
- Response: arquivo binário (video/mp4 ou audio/mp3)

**GET /clips**
- Lista todos os clips (paginado)
- Query params: `?page=1&limit=20`
- Response: `{ "clips": [], "total": int, "page": int }`  

