function startReloadUI() {
    const div = document.createElement("div")
    div.className = "msgbox-overlay"
    div.style.opacity = 1
    div.style.textAlign = "center"
    div.innerHTML = `<div class="msgbox">
<div class="msg">Reloading container</div>
</div>`
    document.querySelector(".monaco-workbench").appendChild(div)
}

function removeElementsByClass(className) {
    let elements = document.getElementsByClassName(className);
    for (let e of elements) {
        e.parentNode.removeChild(e)
    }
}

function stopReloadUI() {
    removeElementsByClass("msgbox-overlay")
}

let tty
window.addEventListener("ide-ready", () => {
    window.ide.workbench.onFileSaved((ev) => {
        if (!ev.endsWith(".sail/Dockerfile")) {
            return
        }

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
        // Show the panel and focus it to prevent the user from editing the Dockerfile.
        srv.showPanel(true)

        startReloadUI()

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
            stopReloadUI()
        }
    })
})
