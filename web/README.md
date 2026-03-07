# Go Mock Server — Admin UI

This is the built-in React admin interface for [Go Mock Server](https://github.com/Caik/go-mock-server).

## Features

- Browse and manage mock responses
- Configure hosts (latency, error simulation)
- Inspect real-time traffic logs

## Development

Requires Node.js v20+ and npm. The backend must be running on port 9090.

**Install dependencies:**

```bash
npm install
```

**Start dev server** (Vite on port 5173, proxies API calls to the backend on 9090):

```bash
npm run dev
# Open http://localhost:5173/ui/
```

## Production Build

```bash
npm run build
```

Output goes to `build/client/`. Pass `--ui-dir /path/to/build/client` to Go Mock Server to serve it.

The Docker image builds this automatically and places the output at `/app/ui`, so `--ui-dir /app/ui` works out of the box.
