# 📁 Estrutura do SvelteKit - YouClips

Esta documentação explica a organização de arquivos e pastas do frontend.

## Arquivos principais (raiz de `web/`)

**`svelte.config.js`** - Configuração do SvelteKit (adapters, preprocessadores)  
**`vite.config.ts`** - Configuração do Vite (proxy para a API, plugins)  
**`package.json`** - Dependências e scripts npm  
**`tsconfig.json`** - Configuração do TypeScript

---

## `src/` - Código fonte

### `src/app.html`
Template HTML base da aplicação. O SvelteKit usa este arquivo como shell:
- `%sveltekit.head%` - Onde o SvelteKit injeta tags `<head>`
- `%sveltekit.body%` - Onde o SvelteKit renderiza sua aplicação

### `src/routes/` - Sistema de rotas (file-based routing)

O SvelteKit usa **roteamento baseado em arquivos**. A estrutura de pastas em `routes/` define automaticamente as URLs da aplicação.

#### Arquivos especiais:

**`+page.svelte`** - Define uma página/rota
- Cada `+page.svelte` representa uma URL
- Exemplo: `src/routes/+page.svelte` → `/` (página inicial)

**`+layout.svelte`** - Layout compartilhado
- Envolve todas as páginas dentro dele (e sub-rotas)
- Usado para elementos comuns como header, footer, navegação
- `{@render children()}` renderiza o conteúdo da página atual

**`+page.ts`** ou **`+page.js`** - Carrega dados antes da página
- Roda no servidor e/ou cliente
- Use para fetch de dados, lógica de carregamento

**`+page.server.ts`** ou **`+page.server.js`** - Carrega dados apenas no servidor
- Para dados sensíveis, acesso a banco de dados
- Não expõe código ao cliente

**`+layout.ts`** / **`+layout.server.ts`** - Carrega dados para o layout
- Dados compartilhados entre páginas

**`+error.svelte`** - Página de erro customizada
- Renderizada quando ocorre erro na rota

**`.css` / `.scss`** - Estilos
- `layout.css` - Estilos globais importados no layout

---

## Exemplos de rotas

```
src/routes/
├── +page.svelte              → /
├── +layout.svelte            → Layout de todas as páginas
├── layout.css                → Estilos globais (TailwindCSS)
│
├── about/
│   └── +page.svelte          → /about
│
├── clips/
│   ├── +page.svelte          → /clips (lista de clips)
│   ├── +page.ts              → Carrega lista de clips
│   │
│   └── [id]/                 → Rota dinâmica
│       ├── +page.svelte      → /clips/123 (detalhe do clip)
│       └── +page.ts          → Carrega clip específico
│
└── api/
    └── clips/
        └── +server.ts        → /api/clips (endpoint de API)
```

### Rotas dinâmicas

Use `[parametro]` para criar rotas dinâmicas:
- `[id]` → aceita qualquer valor (ex: `/clips/123`, `/clips/abc`)
- `[...rest]` → captura múltiplos segmentos (ex: `/docs/a/b/c`)

**Exemplo de uso:**
```svelte
<!-- src/routes/clips/[id]/+page.svelte -->
<script lang="ts">
	import { page } from '$app/stores';
	// $page.params.id contém o valor do ID da URL
</script>

<h1>Clip ID: {$page.params.id}</h1>
```

---

## `src/lib/` - Componentes e utilitários reutilizáveis

Código que você quer reutilizar em várias páginas.

**`lib/index.ts`** - Exporta módulos para importar via `$lib/...`  
**`lib/components/`** - Componentes Svelte reutilizáveis  
**`lib/utils/`** - Funções utilitárias  
**`lib/stores/`** - Stores do Svelte (estado global)  
**`lib/assets/`** - Imagens, ícones, fontes

**Exemplo de importação:**
```typescript
import { myFunction } from '$lib/utils';
import MyComponent from '$lib/components/MyComponent.svelte';
```

O alias `$lib` sempre aponta para `src/lib/`.

---

## `static/` - Arquivos estáticos públicos

Arquivos servidos diretamente sem processamento.

- `robots.txt` - Regras para crawlers de busca
- Imagens, PDFs, fontes que não precisam ser processadas
- Favicon, manifest.json, etc.

**Exemplo:** `static/logo.png` → acessível em `/logo.png`

---

## Estrutura atual do YouClips

```
web/
├── src/
│   ├── app.html                    # Template HTML base
│   ├── routes/
│   │   ├── +layout.svelte          # Layout global
│   │   ├── layout.css              # Estilos globais (Tailwind)
│   │   ├── +page.svelte            # Página inicial (criar clip)
│   │   └── clips/
│   │       ├── +page.svelte        # Lista de clips (/clips)
│   │       └── [id]/
│   │           └── +page.svelte    # Detalhes do clip (/clips/123)
│   └── lib/
│       ├── index.ts
│       └── assets/
│           └── favicon.svg
│
├── static/
│   └── robots.txt
│
├── svelte.config.js
├── vite.config.ts                  # Proxy para API Go (/clips, /metadata)
├── package.json
└── tsconfig.json
```

---

## Rotas implementadas

### `/` - Página inicial (criar novo clip)
**Arquivo:** `src/routes/+page.svelte`

Permite ao usuário criar um novo clip de vídeo do YouTube.

**Fluxo:**
1. Usuário insere URL do YouTube
2. Clica em "Buscar" para obter metadados (título e duração)
3. Usa **dual range sliders** para selecionar intervalo de tempo
4. Escolhe formato (vídeo ou áudio)
5. Cria o clip

**Funcionalidades:**
- Busca metadados via `POST /api/metadata`
- Exibe título e duração do vídeo em formato HH:MM:SS
- Dual sliders com validação automática (início sempre < fim)
- Mostra duração do clip selecionado em tempo real
- **Proteção contra duplicação**: Botão desabilitado durante processamento
- Feedback visual de loading e erros
- Link para "Ver Clips" no header

### `/clips` - Lista de clips
**Arquivo:** `src/routes/clips/+page.svelte`

Exibe todos os clips criados pelo usuário.

**Funcionalidades:**
- Lista todos os clips com paginação
- Mostra status com cores (verde/amarelo/vermelho)
- Exibe título, duração e tamanho
- Botão de download para clips concluídos
- **Botão de apagar clip** com confirmação e loading state
- Link para página de detalhes de cada clip
- **Auto-refresh a cada 5 segundos** se houver clips processando
- Atualização automática da lista após apagar clip
- Mensagem amigável quando não há clips
- Contador total de clips

### `/clips/[id]` - Detalhes do clip
**Arquivo:** `src/routes/clips/[id]/+page.svelte`

Página de detalhes de um clip específico.

**Funcionalidades:**
- Mostra informações completas do clip
- Status visual com indicador de progresso
- Botão de download grande quando concluído
- **Botão de apagar clip** no header com confirmação
- Redireciona para lista após apagar
- **Auto-refresh a cada 3 segundos** enquanto processa
- Mensagens contextuais por status (processando/falhou/pronto)
- Link para voltar à lista

---

## Funções utilitárias

### `formatTime(seconds: number): string`
Converte segundos em formato de tempo legível (HH:MM:SS ou MM:SS).

**Exemplo:**
```typescript
formatTime(213) // "03:33"
formatTime(3661) // "01:01:01"
```

### `formatSize(bytes: number): string`
Converte bytes em formato legível (B, KB, MB, GB).

**Exemplo:**
```typescript
formatSize(1024) // "1 KB"
formatSize(1048576) // "1 MB"
```

### `getStatusColor(status: string): string`
Retorna classes CSS do Tailwind para colorir status.

**Mapeamento:**
- `completed` → verde (`bg-green-100 text-green-800`)
- `processing` → amarelo (`bg-yellow-100 text-yellow-800`)
- `failed` → vermelho (`bg-red-100 text-red-800`)

### `getStatusText(status: string): string`
Traduz status para português.

**Mapeamento:**
- `completed` → "Concluído"
- `processing` → "Processando"
- `failed` → "Falhou"

---

## Integração com API

### Endpoints utilizados

**`POST /api/metadata`** - Buscar metadados do vídeo
```typescript
fetch('/api/metadata', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ url: 'https://youtube.com/...' })
});
// Resposta: { title: "...", duration: 213 }
```

**`POST /api/clips`** - Criar novo clip
```typescript
fetch('/api/clips', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://youtube.com/...',
    start_time: 10,
    end_time: 30,
    format: 'video'
  })
});
// Resposta: { id: 1, status: "processing" }
```

**`GET /api/clips?page=1&limit=20`** - Listar clips
```typescript
fetch('/api/clips?page=1&limit=20');
// Resposta: {
//   clips: [...],
//   total: 50,
//   page: 1
// }
```

**`GET /api/clips/{id}`** - Obter detalhes do clip
```typescript
fetch('/api/clips/1');
// Resposta: {
//   id: 1,
//   status: "completed",
//   title: "...",
//   duration: 20,
//   size: 1234567,
//   download_url: "/clips/1/download"
// }
```

**`DELETE /api/clips/{id}`** - Apagar clip
```typescript
fetch('/api/clips/1', {
  method: 'DELETE'
});
// Resposta: { deleted: true }
```

**`GET /clips/{id}/download`** - Baixar arquivo do clip
```html
<a href="/clips/1/download" download>Download</a>
```

### Proxy configurado

O Vite está configurado para fazer proxy das requisições `/api/*`:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, '')
    }
  }
}
```

Isso permite que:
- O frontend em `localhost:5173` faça requisições sem CORS
- As rotas do SvelteKit (`/clips`) não conflitem com a API
- Requisições `/api/clips` sejam reescritas para `/clips` no backend

---

## ✅ Funcionalidades implementadas

### 1. ✅ Página de listagem de clips
Implementado em `src/routes/clips/+page.svelte`
- Lista todos os clips com paginação
- Auto-refresh a cada 5 segundos para clips processando
- Botões de download, detalhes e apagar
- Confirmação antes de apagar

### 2. ✅ Página de detalhes do clip
Implementado em `src/routes/clips/[id]/+page.svelte`
- Exibe informações completas do clip
- Auto-refresh a cada 3 segundos durante processamento
- Botão de download grande quando concluído
- Botão de apagar no header com redirecionamento

### 3. ✅ Busca de metadados de vídeo
Implementado na página inicial (`src/routes/+page.svelte`)
- Endpoint `POST /api/metadata` integrado
- Dual range sliders para seleção de intervalo
- Formatação de tempo HH:MM:SS

### 4. ✅ Proteção contra duplicação de clips
Implementado na página inicial
- Botão de criar clip desabilitado durante processamento
- Previne múltiplos cliques acidentais

### 5. ✅ Funcionalidade de apagar clips
Implementado em ambas páginas de clips
- DELETE em `src/routes/clips/+page.svelte` (lista)
- DELETE em `src/routes/clips/[id]/+page.svelte` (detalhes)
- Confirmação antes da ação
- Feedback visual durante operação
- Atualização automática da interface

---

## 🚀 Próximos passos sugeridos

### 1. Criar componentes reutilizáveis

Extrair lógica repetida para componentes:

```
src/lib/components/ClipCard.svelte    # Card para exibir clip
src/lib/components/Header.svelte      # Cabeçalho da aplicação
src/lib/components/TimeSlider.svelte  # Slider de tempo reutilizável
src/lib/components/StatusBadge.svelte # Badge de status
```

**Benefícios:**
- Reduz duplicação de código
- Facilita manutenção
- Permite reutilização entre páginas

### 2. Adicionar stores para estado global

```typescript
// src/lib/stores/clips.ts
import { writable } from 'svelte/store';

export const clipsStore = writable<Clip[]>([]);
export const currentClip = writable<Clip | null>(null);
```

**Benefícios:**
- Compartilhar dados entre páginas
- Evitar fetches redundantes
- Cache de clips já carregados

### 3. Melhorar UX com loading states

- Skeletons durante carregamento
- Animações de transição entre páginas
- Toast notifications para feedback de ações

### 4. Adicionar validações e tratamento de erros

- Validar formato de URL do YouTube
- Limite mínimo/máximo de duração do clip
- Mensagens de erro mais descritivas
- Retry automático em caso de falha

### 5. Features adicionais

- **Histórico de URLs**: Salvar URLs recentes no localStorage
- **Preview do vídeo**: Embed do YouTube com preview
- **Compartilhamento**: Gerar links compartilháveis de clips
- **Temas**: Dark mode / Light mode
- **Busca e filtros**: Buscar clips por título, ordenar por data

---

## Links úteis

- [Documentação SvelteKit](https://svelte.dev/docs/kit)
- [Roteamento](https://svelte.dev/docs/kit/routing)
- [Loading data](https://svelte.dev/docs/kit/load)
- [TailwindCSS](https://tailwindcss.com/docs)
