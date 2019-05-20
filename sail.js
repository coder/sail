(function() {
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
    function rebuild() {
        const tsrv = window.ide.workbench.terminalService

        if (tty == null) {
            tty = tsrv.createTerminal({
                name: "sail",
                isRendererOnly: true,
            }, false)
        } else {
            tty.clear()
        }
        let oldTTY = tsrv.getActiveInstance()
        tsrv.setActiveInstance(tty)
        // Show the panel and focus it to prevent the user from editing the Dockerfile.
        tsrv.showPanel(true)

        startReloadUI()

        const ws = new WebSocket("ws://" + location.host + "/sail/api/v1/reload")
        ws.onmessage = (ev) => {
            const msg = JSON.parse(ev.data)
            const out = atob(msg.v).replace(/\n/g, "\n\r")
            tty.write(out)
        }
        ws.onclose = (ev) => {
            if (ev.code === 1000) {
                tsrv.setActiveInstance(oldTTY)
            } else {
                alert("reload failed; please see logs in sail terminal")
            }
            stopReloadUI()
        }
    }

    window.addEventListener("ide-ready", () => {
        const statusBarService = window.ide.workbench.statusbarService
        statusBarService.addEntry({
            text: "rebuild",
            tooltip: "press super+alt+r to rebuild",
            command: "rebuild-sail-container"
        }, 0)

        const commandRegistry = window.ide.workbench.commandRegistry
        commandRegistry.registerCommand({
            id: "rebuild-sail-container",
            handler: (accessor, args) => {
                rebuild()
            },
            description: {
                description: "Rebuild sail container",
                args: []
            }
        })
    })
}())