
# Requisitos Funcionais

- Baixar vídeos e áudios do Youtube como clips
- Clips podem ser compartilhados . Parecido com o MyInstants
- Usuários podem se cadastrar no aplicativo
- A ferramenta deverá limitar automaticamente o intervalo do clip 


# Requisitos não funcionais 

- Clips privados ou publicados devem ter no máximo 30 segundos e poderá ser escolhido qualquer tempo do vídeo.
- Clips que ficaram salvos apenas no dispositivo do usuário, não teram limite de tempo
- Somente usuários cadastrados podem compartilhar e salvar clips  
- Usuário não cadastrados podem criar e baixar clips próprios localmente e de outros compartilhados
- Clips de usuários não cadastrados não ficam salvos no sistema
- Taxa máxima de download deverá ser de 1MB




# 1 Parte do projeto
Desenvoler o criador de clips de videos do Youtube


- Deve criar clips de vídeos do Youtube 

## Fluxo principal 

- Usuário está cadastrado

1. Usuário copia o URL válido do vídeo no input da ferramenta
2. Usuário determina o intervalo de tempo do vídeo, limitado sempre a 30s (0:00-0:30, 0:00-0:15, 2:10-2:25 ).
3. Usuário escolhe o formato do clip (como vídeo ou áudio)
4. Usuário clica em Criar Clip 
5. Clip é salvo e pode ser baixado pelo usuário 

## Fluxo alternativo 1 
1. Após passo 1 se o usuário marcar uma checkbox "Salvar no meu computador", o clip não terá limite de tempo.
2. Após passo 4 o clip não fica salvo e só podera ser baixado.

## Tecnologias adotadas

- Backend: Go por motivos de aprendizados e devido a sua capacidade de concorrência.
- Frontend: a decidir
- Banco de dados: PostgresSQL em container 
- Processamento de vídeo yt-dlp e FFmpeg

## Modelagem relacional


Table user {
  id integer [primary key]
  videos_downloaded integer
  password hash
  email string 

}

Table clip {
  id integer [primary key]
  created_at timestamp
  name string
  duration_seconds string
  size bigint
  path string
  user_id integer
  original_url string
  status clip_status
  is_public bool 
}

enum clip_status {
  failure
  processing
  completed
}

Ref: clip.user_id > user.id 


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
│  Service      │  ← Use cases / Services
│   Layer       │
└───────▲───────┘
        │
┌───────┴───────┐
│  Entities     │  ← Entidades + Regras
│   Layer       │
└───────▲───────┘
        │
┌───────┴───────┐
│ Infrastructure│  ← DB, FFmpeg, yt-dlp
└───────────────┘

Processamento assíncrono: API → Queue → Workers → FFmpeg / yt-dlp

## Questões Críticas:

     - Limite de 30s: Aplicado apenas em clips públicos, mas falta validação clara no fluxo
     - Storage: Não há estratégia de armazenamento (local, S3, CDN) nem gerenciamento de espaço/limpeza
     - Segurança: Faltam endpoints de autenticação (/login, /auth) e não menciona tokens/sessions

## Sugestões Técnicas:

     - Definir política de retenção/exclusão de clips não-cadastrados
     - Considerar Redis para fila de processamento assíncrono

## Endpoints 

## User 
- /user -> lista de usuarios 
  - GET obtém a lista de usuários 
  - POST cria um novo usuário 
- /user/{id}/clips -> lista de clips do usuairio
  - GET -> obtém a lista de clips
  - POST -> cria um novo clip do usuario 
- /user/{id}/clips/{id} -> clip específico de um usuario 
  - GET obtém o clip do usuário 

## Clips 

- /clips -> lista de clips publicados
  - GET pega a lista 
  - POST Salva um novo clip
- /clips/{id} -> obtém um clip específico 
  - GET pega o clip  

