# colima-cli

Go-based Colima and Docker VM diagnostics and management. Version **0.0.1**.

## Installation

```bash
git clone https://github.com/r14r/colima-cli.git
cd colima-cli
just install
```

Installs into `/Users/Shared/CLOUD/DeveloperTools/bin` by default; override using `INSTALL_DIR`.

## Examples

```bash
colima-cli status
colima-cli status --quick
colima-cli status --json
colima-cli status --resources
colima-cli doctor
colima-cli test
colima-cli list
colima-cli start --cpu 4 --memory 8
colima-cli stop
colima-cli restart
colima-cli exec -- uname -a
colima-cli docker ps -a
colima-cli docker images
colima-cli docker df
colima-cli context current
colima-cli context use --yes
colima-cli docker prune --yes
colima-cli delete --yes
colima-cli wait 60
```

Status checks do not change Docker context. Docker commands explicitly use the Colima context. Destructive actions require `--yes`.

## Development

```bash
just precommit
just build
just release
```

Requires Go 1.21+ to build. No external Go dependencies. Colima, Docker and Lima required for runtime commands.

## v0.1.0: Containers, Compose, dashboard

```bash
colima-cli containers
colima-cli containers --running
colima-cli containers --stats
colima-cli compose ls
colima-cli compose ps -a
colima-cli compose logs -f
colima-cli compose up -d --build
colima-cli compose down
colima-cli compose config
colima-cli compose stats
colima-cli dashboard --interval 5
colima-cli dashboard --once
```

All Docker operations target the Colima context. The dashboard is read-only. `compose stats` looks up project container IDs and shows a one-shot stats snapshot.
