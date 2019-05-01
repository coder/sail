let reloadInProgress = false
let tty
window.addEventListener("ide-ready", (ev) => {
    window.ide.workbench.onFileSaved((ev) => {
        if (!ev.endsWith(".sail/Dockerfile")) {
            return
        }

        if (reloadInProgress) {
            alert("reload is still in progress")
            return
        }
        reloadInProgress = true

        const srv = window.ide.workbench.terminalService

        if (tty == null) {
            tty = srv.createTerminal({
                name: "sail",
                isRendererOnly: true,
            }, false)
        } else {
            tty.clear()
        }
        let oldTTY = srv.getActiveInstance()
        srv.setActiveInstance(tty)

        const ws = new WebSocket("ws://" + location.host + "/sail/api/v1/reload")
        ws.onmessage = (ev) => {
            const msg = JSON.parse(ev.data)
            const out = atob(msg.v).replace(/\n/g, "\n\r")
            tty.write(out)
        }
        ws.onclose = (ev) => {
            if (ev.code === 1000) {
                srv.setActiveInstance(oldTTY)
            } else {
                alert("reload failed; please see logs in sail terminal")
            }
            reloadInProgress = false
        }
    })
})
