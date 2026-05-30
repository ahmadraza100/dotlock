# dotlock

```
 ____  ____ _____ _     ____  ____ _  __
 /  _ \/  _ Y__ __Y \   /  _ \/   _Y |/ /
 | | \|| / \| / \ | |   | / \||  / |   /
 | |_/|| \_/| | | | |_/\| \_/||  \_|   \
 \____/\____/ \_/ \____/\____/\____|_|\_\
```

Encrypted `.env` vault manager with a beautiful terminal UI.
 
[![CI](https://github.com/ahmadraza100/dotlock/actions/workflows/ci.yml/badge.svg)](https://github.com/ahmadraza100/dotlock/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ahmadraza100/dotlock)](https://goreportcard.com/report/github.com/ahmadraza100/dotlock)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/ahmadraza100/dotlock.svg)](https://pkg.go.dev/github.com/ahmadraza100/dotlock)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/ahmadraza100/dotlock)](https://github.com/ahmadraza100/dotlock/releases/latest)
 
dotlock replaces your scattered plaintext `.env` files with a single encrypted vault per project — managed through a fast, keyboard-driven terminal interface. No cloud. No plaintext on disk. No dependencies.
 
---
 
## Preview
 
![dotlock TUI](https://github.com/user-attachments/assets/f642862a-5757-45d8-b3e2-207fbc96ea93)
 
---
 
## Features
 
- age encryption — every vault uses [filippo.io/age](https://age-encryption.org) (X25519 + ChaCha20-Poly1305)
- Double encryption — each secret encrypted individually, then the entire vault encrypted again
- Multiple profiles — manage `dev`, `staging`, `prod` environments side by side
- Diff view — see exactly which keys are missing or changed between two profiles before deploying
- Import existing `.env` files instantly into any profile
- Export to shell, Docker, or GitHub Actions format in one command
- Secret masking — values hidden by default, press `v` to reveal for 30 seconds then auto-clears
- Atomic writes — vault never corrupts on crash
- Memory hygiene — decrypted bytes zeroed after use
- Zero network calls — 100% offline, no telemetry ever
- Single binary — no runtime, no dependencies, works on Mac, Linux, and Windows
---
 
## Installation
 
**Mac/Linux — quickest**
 
```bash
curl -sSL https://raw.githubusercontent.com/ahmadraza100/dotlock/main/install.sh | sh
```
 
**Go install**
 
```bash
go install github.com/ahmadraza100/dotlock@latest
```
 
**Manual** — download your platform binary from the [releases page](https://github.com/ahmadraza100/dotlock/releases/latest) and move to `/usr/local/bin`.
 
---
 
## Quick start
 
```bash
cd my-project
dotlock
```
 
---
 
## The TUI
 
dotlock opens a two-panel terminal interface focused on your current project.
 
```
╭────────────────────────────╮╭────────────────────────────────────────╮
│   profiles                 ││   secrets                              │
│                            ││                                        │
│ │ dev                      ││ │ DATABASE_URL          encrypted      │
│ │ staging                  ││ │ REDIS_URL             encrypted      │
│ │ prod                     ││ │ STRIPE_KEY            encrypted      │
│                            ││ │ JWT_SECRET            encrypted      │
╰────────────────────────────╯╰────────────────────────────────────────╯
╭────────────────────────────────────────────────────────────────────────╮
│ [n] new · [e] edit · [v] reveal · [d] delete · [i] import · [x] export│
│ [f] diff · [tab] switch · [/] search · [?] help · [q] quit            │
╰────────────────────────────────────────────────────────────────────────╯
```
 
**Profiles panel** — lists all environments in your vault (`dev`, `staging`, `prod`). Switch between them with arrow keys and Enter. Add a new profile with `[a]`, delete with `[D]`.
 
**Secrets panel** — lists all keys in the selected profile. Values are always masked. Press `[v]` to reveal a value for 30 seconds — it hides itself automatically after the timer expires. Press `[n]` to add a new secret, `[e]` to edit, `[d]` to delete.
 
**Import** — press `[i]` to import an existing `.env` file directly into the active profile. All values are encrypted on import.
 
**Export** — press `[x]` to export the active profile.


```
  DATABASE_URL     identical in both
~ PORT             exists in both — values differ
- NEW_FLAG         dev only — missing in staging
+ SENTRY_DSN      staging only — missing in dev
```
 
---
 
## Security model
 
Every secret is encrypted twice — once per entry, and again at the vault level — using `filippo.io/age` (X25519 + ChaCha20-Poly1305). The `.dotlock` file on disk is fully encrypted and safe to commit to git.
 
Private keys live at `~/.config/dotlock/keys/default.agekey` with `0600` permissions. All vault writes are atomic. Decrypted bytes are zeroed from memory immediately after use. dotlock makes zero network calls.
 
**Known limitations**
 
- Key loss means data loss — no recovery mechanism
- Single-user only — no team sharing in v0.1
- No key rotation support
- No audit log
---
 
## File locations
 
| Path | Contents |
|------|----------|
| `.dotlock` | Encrypted vault — safe to commit |
| `~/.config/dotlock/config.json` | User config — no secrets |
| `~/.config/dotlock/keys/default.agekey` | Private key — never commit this |
 
---
 
## Built with
 
- [filippo.io/age](https://age-encryption.org) — encryption
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [charmbracelet/huh](https://github.com/charmbracelet/huh) — interactive prompts
- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
---
 
## Contributing
 
```bash
git clone https://github.com/ahmadraza100/dotlock
cd dotlock
make install
make test
make lint
```
 
See [CONTRIBUTING.md](CONTRIBUTING.md) for details.
 
---
 
MIT License — see [LICENSE](LICENSE)
 


