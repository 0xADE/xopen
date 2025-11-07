# xopen

A GUI launcher for [ade-exe-ctld](https://github.com/0xADE/ade-ctld), built with [Go](https://go.dev) and the [Gio](https://gioui.org) UI toolkit.

## Features

- Fast search (by filter) and navigation of indexed applications
- Launch applications with Enter key or mouse click
- Simple, clean interface with filter field and application list

## Usage

1. Make sure `ade-exe-ctld` is running
2. Run `xopen`:
   ```bash
   make run
   # or
   go run ./cmd/xopen
   ```
3. Type to filter applications
4. Use arrow keys to navigate
5. Press Enter or click to launch

## Building

```bash
make build
```

The binary will be in `build/xopen`.

## Installation

```bash
make install
```

This installs `xopen` to `/usr/local/bin`.
