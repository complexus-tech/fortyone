# Docker Self-Host

This folder contains the Docker Compose assets used by the FortyOne installer.

## Contents

- `compose.yml`: Production-ready stack for projects, API, and worker.
- `fortyone.env`: Environment template used by the installer.
- `Dockerfile.migrations`: Builds a migrations image bundled with SQL files.

## Installer

Use the hosted installer (recommended):

```bash
curl -fsSL https://fortyone.app/install.sh | sh -
```

If you already cloned the repo, you can run:

```bash
./install.sh
```

For a local build using source code:

```bash
./install.sh
# choose "Dev Build (local source)"
```
