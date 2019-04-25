+++
type="docs"
title="Shares"
browser_title="Sail - Docs - Shares"
+++

Projects and hats can specify shares using command labels of form

`share.<share_name>="host_path:guest_path"`.

For example,

```Dockerfile
LABEL share.go_mod="~/go/pkg/mod:/root/go/pkg/mod"
```

Shares the host's Go mod cache with the guest.

---

Shares are recommended for

- Filesystem-level caches
    - Go mod cache
    - Yarn cache
- User-specific configuration
    - VS Code configuration (auto)
    - SSH keys
    - gitconfig
- Working data
    - Project files (auto)
    - Data analysis results

It's important to keep in mind that shares can easily undermine the
reproducibility and consistency of your environments. Be careful with blanket shares
such as `~:~` which introduce variance.