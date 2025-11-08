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

### Key bindings (implemented partially)

| Key combo         | Alternative combo | Action                                                                                                         |
| <Esc>             |                   | Close xopen.                                                                                                   |
| <Enter>           | Ctl+M             | Run selected command and exits xopen.                                                                          |
| Alt+<Enter>       |                   | Run selected command; don't exit xopen, just clear current filter string.                                      |
| Shift+<Enter>     |                   | Run selected command in a terminal and exits xopen.                                                            |
| Alt+Shift+<Enter> |                   | Run selected command in a terminal; don't exit xopen, just clear current filter string.                        |
| Ctl+"             | Ctl+'             | NAMES mode. Filter by application titles and file names. This mode enabled by default.                         |
| /                 | Ctl+/             | PATHS mode. Filter by paths only.                                                                              |
| Ctl+@             | Ctl+2             | Switches to CATEGORIES mode.                                                                                   |
| Ctl+#             | Ctl+3             | Switches to TAGS mode.                                                                                         |
| Ctl-<Bsp>         |                   | Clears all the filter string for the current filter.                                                           |
| Alt-<Bsp>         |                   | Removes current filter and backs to the previous.                                                              |
| Ctl-Alt-<Bsp>     |                   | Resets all the filters and clears current filter string. Keeps the latest filter mode.                         |
| \|\|              | SPC-OR-SPC        | Starts a new filter set with OR logical op. The filter mode the same as in previous filter. Keep old filters.  |
| &&                | SPC-AND-SPC       | Starts a new filter set with AND logical op. The filter mode the same as in previous filter. Keep old filters. |
| &&                |                   |                                                                                                                |
| Ctl+A             |                   | Turns on ARGS mode. Just clear filter string. New typed filter string used as argument for selected command.   |
| \\                |                   | Prevents following character to be treated as a control (for example don't treat && as a switch filter).       |
| Ctl+S             |                   | Edit settings for the curently selected command. Settings saved for further xopen runs.                        |
|                   |                   |                                                                                                                |

### Filtering (implemented partially)

Filtering of executables allowed by:

- names or titles (for desktop files)
- full paths
- categories (for desktop files)
- tags (custom exe-ctld settings)

#### Rules for switching filtering modes

0. Все наборы фильтров объединяются по AND, eсли не было явно указано OR для
   каких-то фильтров. AND имеет больший приоритет, чем OR, как пример: `(filter1
   AND filter2) OR (filter3 AND filter4)`.
1. Typing / or ./ or .. turn on current filter as PATHS filter. Filter mode stays
   even the filter string erased. User could switch to another mode by hot keys.
2. При переходе из PATHS в другой режим остается только текст справа от
   последнего /, всё перед ним, включая сам слеш, стирается.
3. && или " AND " (заглавными между пробелами) добавляет строку нового фильтра в
   режиме AND.
4. || или " OR " (заглавными между пробелами) добавляет новый фильтр в режиме
   OR.
5. Настройки выбранной команды по Ctl+S включают: префикс (str), use terminal
   (bool), custom color (#rgb), log output (bool).

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
