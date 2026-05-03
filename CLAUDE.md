# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Visão Geral

**skill-manager** é uma aplicação desktop construída com [Wails v2](https://wails.io/), combinando um backend Go com um frontend React/TypeScript. O frontend é compilado e embutido no binário Go final, gerando um executável standalone.

## Comandos de Desenvolvimento

### Desenvolvimento (hot reload)
```bash
wails dev
```
Inicia o backend Go e o servidor Vite simultaneamente. O frontend fica em `localhost:34115` durante o dev.

### Build para produção
```bash
wails build
```
Gera o executável em `build/bin/skill-manager`.

### Frontend isolado
```bash
cd frontend
npm install       # instalar dependências
npm run dev       # servidor Vite standalone
npm run build     # compilar TypeScript + Vite
```

## Arquitetura

```
skill-manager/
├── main.go          # entry point: cria App, configura janela Wails
├── app.go           # struct App com métodos expostos ao frontend
├── wails.json       # configuração do projeto Wails
└── frontend/
    ├── src/
    │   ├── main.tsx         # ponto de entrada React
    │   └── App.tsx          # componente principal, chama métodos Go
    └── wailsjs/             # bindings TypeScript auto-gerados (não editar)
        └── go/main/         # interfaces TypeScript para métodos Go
```

## Padrão de Comunicação Go ↔ Frontend

Métodos públicos da struct `App` em `app.go` são automaticamente expostos ao frontend pelo Wails.

**Backend (Go)**:
```go
func (a *App) Greet(name string) string { ... }
```

**Bindings gerados** em `frontend/wailsjs/go/main/App.d.ts`:
```typescript
export function Greet(arg1: string): Promise<string>;
```

**Frontend (TypeScript)**:
```typescript
import { Greet } from '../wailsjs/go/main/App';
const result = await Greet(name);
```

Após adicionar ou alterar métodos em `app.go`, execute `wails generate module` para atualizar os bindings em `frontend/wailsjs/`.

## Notas

- **Não há testes configurados** — nem no backend Go nem no frontend.
- A pasta `build/bin/` e `frontend/dist/` são artefatos de build (ignorados pelo git).
- O contexto Wails (`ctx`) é inicializado em `app.startup()` e necessário para usar `runtime.*` (notificações, diálogos de arquivo, etc.).
