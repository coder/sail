+++
type="docs"
title="Hats"
browser_title="Sail - Docs - Hats"
section_order=3
+++

A _hat_ is a build directory with a Dockerfile that allows you to extend
every project environment with your own personalization. Hats allow you to
bring your own tooling, configuration, and workflow into every Sail project
you work on.

In order for Sail to extend the project's environment, the hat Dockerfile's
`FROM` clause is replaced with the repository-provided image.

For example:

```Dockerfile
FROM ubuntu

RUN sudo apt install fish
RUN chsh user -s $(which fish)
```

is a hat that would install fish, and configure it as the default
shell regardless of which shell the repository-provided image uses.

The `FROM ubuntu` will be replaced with `FROM <repo_image>` when sail
assembles your dev container in order to extend the project's environment.

---

`sail` promotes the use of Ubuntu/apt-based dev containers so that hats are
reliable.

You can only wear a single hat at a time.

### GitHub

To enable expirementation, hats can be used from github like so:

`-hat github:ammario/dotfiles`

---

Hats enable personalization, so **GitHub hats should just be used for experimentation.**
