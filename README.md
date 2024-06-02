# Burn it

Self-destruct messaging app. Send notes and files that will self-destruct after being read.

## How to

### Quickstart

With this method, you don't need all tools for development but the minimum to run dependencies and app.

Requirements:

- Docker
- Make

And run all with this command:

```bash
make docker-run-all
```

Access app from http://localhost:8000.

And be happy.

### For development

Requirements:

- Golang 1.18+
- Docker
- Make

Run dependencies with docker:

```bash
make docker-dependencies
```

Run app:

```bash
make run
```

And access app from http://localhost:8000 and API from http://localhost:8000/api.
