# Frontend Web & Dashboard

O frontend da aplicação é extremamente limpo e projetado para não depender de Node.js, Webpack ou Vite. 

Toda a interface está localizada na pasta `web/` e é gerada via `html/template` e diretivas `go:embed` nativas do Go.

## 1. Estrutura de Arquivos
```
web/
├── static/
│   ├── css/
│   │   └── tailwind.css (Baixado via CDN ou arquivo estático pré-compilado v3.4)
│   ├── js/
│   │   └── dashboard.js (Vanilla JS)
├── index.html (Landing Page)
└── dashboard.html (Painel do Usuário logado)
```

## 2. Landing Page (`index.html`)
Apresenta o serviço. É a página inicial servida ao acessar a raiz do subdomínio do frontend (ex: `app.seudominio.com/`). Contém um formulário simples de Login para acessar o painel.

## 3. Painel do Usuário (`dashboard.html`)
Página privada. O backend de Go utiliza Sessões com Cookies nativos para proteger esta página.

**Funcionalidades:**
- Visualização do Saldo (Créditos disponíveis).
- Visualização de estatísticas simples (Total de Tokens consumidos).
- Lista de API Keys pertencentes ao usuário.
- Histórico básico de requisições.

A página faz consultas Ajax via `fetch()` (Vanilla JS) para a rota `/v1/admin/me` que retorna os dados atualizados do usuário em formato JSON, renderizando-os dinamicamente na DOM com manipulação padrão.
