# pomo

A CLI pomodoro timer built for personal use.
Supports progress bars and long break support.

## Installation

Requires [Go 1.24+](https://go.dev/dl/).

```bash
go install github.com/steenfuentes/pomo@latest
```

Ensure `$HOME/go/bin` is in your `PATH`:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

## Usage

```bash
pomo start                    # 50min work, 10min short break, infinite cycles
pomo start -p 25 -s 5         # 25min work, 5min short break
pomo start -e 4 -l 15         # 15min long break every 4 cycles
pomo start -c 4               # Run exactly 4 work cycles then exit
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--pomodoro` | `-p` | 50 | Work duration (minutes) |
| `--short` | `-s` | 10 | Short break duration (minutes) |
| `--long` | `-l` | 15 | Long break duration (minutes) |
| `--long-every` | `-e` | 0 | Long break frequency (0 = disabled) |
| `--cycles` | `-c` | 0 | Total work cycles (0 = infinite) |

## License

MIT
