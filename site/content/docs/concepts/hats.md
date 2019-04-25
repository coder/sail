+++
type="docs"
title="Hats"
browser_title="Sail - Docs - Hats"
+++

A _hat_ is a build directory with a Dockerfile which has it's `FROM` clause
replaced with the repository-provided image. Essentially, hats let you
personalize your development environment.

For example:

```Dockerfile
FROM ubuntu-dev
RUN sudo apt install fish
RUN chsh user -s $(which fish)
```

is a hat that would install fish, and configure it as the default
shell regardless of which shell the repository-provided image uses.

The `FROM ubuntu-dev` will be replaced with `FROM <repo_image>` when sail
assembles your dev container.

---

`sail` promotes the use of Ubuntu/apt-based dev containers so that hats are
reliable.

You can only wear a single hat at a time.

### GitHub

To enable expirementation, hats can be used from github like so:

`-hat github:ammario/dotfiles`

---

Hats enable personalization, so **GitHub hats should just be used for expirementation.**