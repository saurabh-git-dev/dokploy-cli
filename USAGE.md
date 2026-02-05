# Dokploy CLI Usage and Examples

This document covers how to use the Dokploy CLI, its global flags, subcommands, and some end-to-end examples.

## Global flags

All commands need a Dokploy URL and API key. You can provide them either as flags or via environment variables:

- `--url` (`-url`) or `DOKPLOY_URL`: Base URL of your Dokploy instance (e.g. `https://your-dokploy-instance.com`).
- `--key` (`-key`) or `DOKPLOY_API_KEY`: Dokploy API key (sent as `x-api-key`).

Example prefix you can reuse (no flags needed if env vars are set):

```bash
export DOKPLOY_URL="https://your-dokploy-instance.com"
export DOKPLOY_API_KEY="YOUR-GENERATED-API-KEY"

dokploy project create --id my-project-id --name "My Project"
```

> All `create` / `create-or-update` commands print the resource **ID** returned by Dokploy (when available) on stdout so you can capture it in scripts and feed it into the next command.

---

## Project commands

### Create project

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  project create \
  --id my-project-id \
  --name "My Project"
```

- Creates a project with the given `name` (Dokploy assigns the real project ID).
- Prints the project ID on stdout if Dokploy includes it in the response.

### Delete project

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  project delete \
  --id my-project-id
```

---

## Environment commands

### Create environment

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  environment create \
  --projectId my-project-id \
  --name "staging"
```

- Either `--id` or `--name` is required by the CLI (typically you use `--name`).
- Prints the environment ID on stdout if Dokploy includes it in the response.

### Delete environment

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  environment delete \
  --id my-environment-id
```

---

## Compose commands

### Get compose (by ID)

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  compose get \
  --id my-compose-id
```

- Fetches a compose app using the official `compose.one` endpoint.
- Prints the full JSON response.

> Note: The current implementation requires `--id`; lookup by `--name` is not supported by the Dokploy `compose.one` API.

### Create or update compose

```bash
# Create a new compose app

dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  compose create \
  --name "my-compose-app" \
  --environmentId my-environment-id \
  --compose-file ./docker-compose.yml \
  --env-vars APP_ENV=staging \
  --env-vars LOG_LEVEL=debug

# Update an existing compose app by ID

dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  compose create \
  --id my-compose-id \
  --environmentId my-environment-id \
  --compose-file ./docker-compose.yml \
  --env-vars APP_ENV=staging
```

- On create (no `--id`): calls Dokploy `compose.create` and prints the created compose ID.
- On update (with `--id`): calls Dokploy `compose.update` and prints the compose ID.

### Delete compose

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  compose delete \
  --id my-compose-id \
  --delete-volumes=true
```

- Calls Dokploy `compose.delete`.
- `--delete-volumes` controls whether volumes are also deleted (defaults to `true`).

### Deploy compose

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  compose deploy \
  --id my-compose-id
```

- Calls Dokploy `compose.deploy` for the given compose ID.

---

## Domain commands

### Create or update domain

```bash
# Create a new domain for a compose app

dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  domain create \
  --host example.com \
  --path / \
  --port 80 \
  --serviceName web \
  --composeId my-compose-id \
  --certificateType letsencrypt \
  --https

# Update an existing domain by ID

dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  domain create \
  --id my-domain-id \
  --host example.com \
  --path /app \
  --port 8080 \
  --serviceName web \
  --composeId my-compose-id
```

- `certificateType` is validated as an enum: one of `none`, `letsencrypt`, or `custom` (default is `none`).
- Creates or updates a domain for a given compose app.
- If no `--id` is provided, the CLI first lists domains for the given `composeId` and will **update** an existing domain (preferring a matching `host` + `path`) instead of creating duplicates; if none exist, it creates a new domain.
- Prints the domain ID on stdout if Dokploy includes it in the response.

### Delete domain

```bash
dokploy \
  --url "$DOKPLOY_URL" \
  --key "$DOKPLOY_KEY" \
  domain delete \
  --id my-domain-id
```

---

## End-to-end example (project → environment → compose → domain)

Below is a simple shell flow that wires everything together. Each step prints an ID that is captured into a variable and passed to the next step.

```bash
export DOKPLOY_URL="https://your-dokploy-instance.com"
export DOKPLOY_KEY="YOUR-GENERATED-API-KEY"

# 1. Create project
PROJECT_ID=$(dokploy --url "$DOKPLOY_URL" --key "$DOKPLOY_KEY" \
  project create --id my-project --name "My Project")

echo "Project ID: $PROJECT_ID"

# 2. Create environment in that project
ENV_ID=$(dokploy --url "$DOKPLOY_URL" --key "$DOKPLOY_KEY" \
  environment create --projectId "$PROJECT_ID" --name "staging")

echo "Environment ID: $ENV_ID"

# 3. Create compose app for that environment
COMPOSE_ID=$(dokploy --url "$DOKPLOY_URL" --key "$DOKPLOY_KEY" \
  compose create \
  --name "my-compose-app" \
  --environmentId "$ENV_ID" \
  --compose-file ./docker-compose.yml \
  --env-vars APP_ENV=staging)

echo "Compose ID: $COMPOSE_ID"

# 4. Create domain pointing to the compose app
DOMAIN_ID=$(dokploy --url "$DOKPLOY_URL" --key "$DOKPLOY_KEY" \
  domain create \
  --host example.com \
  --path / \
  --port 80 \
  --serviceName web \
  --composeId "$COMPOSE_ID" \
  --certificateType letsencrypt \
  --https)

echo "Domain ID: $DOMAIN_ID"

# 5. Deploy the compose app

dokploy --url "$DOKPLOY_URL" --key "$DOKPLOY_KEY" \
  compose deploy --id "$COMPOSE_ID"
```

This flow gives you all the IDs on stdout so you can store them, reuse them later, or plug them into CI/CD pipelines.
