# binoc

Observability playground — a minimal Go service paired with different monitoring stacks.

## Quickstart

```bash
make up                       # starts the default stack (grafana-lgtm)
make up STACK=grafana-lgtm    # or pick one explicitly
make list                     # show available stacks
```

Open http://localhost to see the navigation page, or go directly:

- **Grafana** http://localhost:3000 (admin / admin)
- **Prometheus** http://localhost:9090

```bash
curl localhost/echo?msg=hello
curl -X POST -d 'ping' localhost/echo
```

## Stacks

| Stack          | Description                         |
|----------------|-------------------------------------|
| `grafana-lgtm` | Loki, Grafana, Tempo, Prometheus   |

## Make Targets

| Target  | Description                        |
|---------|------------------------------------|
| `up`    | Build and start the stack          |
| `down`  | Stop the stack and remove volumes  |
| `logs`  | Tail logs from all services        |
| `build` | Build the echo service image       |
| `list`  | List available stacks              |
