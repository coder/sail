+++
type="docs"
title="Fuzzy Run"
browser_title="Sail - Docs - Fuzzy Run"
section_order=0
+++

If you frequently use `sail` from the command line,
install [fzf](https://github.com/junegunn/fzf) and create some shell
aliases to save you some keystrokes.

## Open project

This commands plops you into fzf to quickly open project.

```
sail run $(sail ls | cut -f1 -d" " | tail -n +2 | fzf --height 5)
```