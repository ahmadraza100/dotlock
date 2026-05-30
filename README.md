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

![dotlock TUI]

![dotlock Diff](screenshot-diff.png)

---

## Features

- Three-panel TUI — profiles and secrets in one view, navigate with arrow keys
- age encryption — every vault uses [filippo.io/age](https://age-encryption.org) (X25519 + ChaCha20-Poly1305)
- Double encryption — each secret is encrypted individually, then the entire vault is encrypted again
- Multiple profiles — manage `dev`, `staging`, `prod` environments side by side
- Diff view — see exactly which keys are missing or changed between two profiles
- Secret masking — values hidden by default, press `v` to reveal for 30 seconds then auto-clears
- Import existing `.env` files instantly
- Atomic writes — vault never corrupts on crash
- Memory hygiene — decrypted bytes zeroed after use
- Zero network calls — 100% offline, no telemetry ever
- Single binary — no runtime, no dependencies, works on Mac, Linux, and Windows

---

## Installation

**Mac/Linux**

```bash
curl -sSL https://raw.githubusercontent.com/ahmadraza100/dotlock/main/install.sh | sh
```

**Homebrew**

```bash
brew tap ahmadraza100/tap
brew install dotlock
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
dotlock init
dotlock ui
```

---

## CLI reference

```bash
dotlock init                         # create vault in current directory
dotlock set KEY                      # add or update a secret
dotlock set KEY --profile staging    # target a specific profile
dotlock get KEY                      # decrypt and print a value
dotlock list                         # list all keys
dotlock delete KEY                   # remove a secret
dotlock diff dev staging             # compare two profiles
dotlock export --format shell        # export KEY="value"
dotlock export --format docker       # export KEY=value
dotlock export --format github       # export via gh secret set
dotlock profile create staging       # create a new profile
dotlock profile list                 # list all profiles
dotlock profile use staging          # switch active profile
dotlock profile delete staging       # delete a profile
dotlock ui                           # open the TUI
dotlock version                      # print version
```

---

## TUI controls

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate |
| `Tab` | Switch panels |
| `n` | New secret |
| `e` | Edit secret |
| `v` | Reveal value for 30 seconds |
| `d` | Delete secret |
| `D` | Delete profile |
| `a` | Add profile |
| `i` | Import `.env` file |
| `x` | Export secrets |
| `f` | Diff two profiles |
| `/` | Search |
| `?` | Help |
| `q` | Quit |

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