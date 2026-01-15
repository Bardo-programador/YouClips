Video Downloader - Parte 1

Servidor Go com endpoint POST /clips para criar pedidos de clip.

Uso:
- go run ./cmd/server
- POST http://localhost:8080/clips com JSON: {"url":"...","start":0,"end":10,"format":"video"}
- Resposta 202 com status processing.

Baseado em ideias.md, arquitetura em camadas e processamento assíncrono a ser implementado.
