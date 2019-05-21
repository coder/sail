(function() {
    let oldonkeydown
    function startReloadUI() {
        const div = document.createElement("div")
        div.className = "msgbox-overlay"
        div.style.opacity = 1
        div.style.textAlign = "center"
        div.innerHTML = `<div class="msgbox">
    <div class="msg">Rebuilding container</div>
    </div>`
        // Prevent keypresses.
        oldonkeydown = document.body.onkeydown
        document.body.onkeydown = ev => {
            ev.stopPropagation()
        }
        document.querySelector(".monaco-workbench").appendChild(div)
    }

    function removeElementsByClass(className) {
        let elements = document.getElementsByClassName(className);
        for (let e of elements) {
            e.parentNode.removeChild(e)
        }
    }

    function stopReloadUI() {
        document.body.onkeydown = oldonkeydown
        removeElementsByClass("msgbox-overlay")
    }

    let tty
    let rebuilding
    function rebuild() {
        if (rebuilding) {
            return
        }
        rebuilding = true

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
            rebuilding = false
        }
    }

    window.addEventListener("ide-ready", () => {
        class rebuildAction extends window.ide.workbench.action {
            run() {
                rebuild()
            }
        }

        window.ide.workbench.actionsRegistry.registerWorkbenchAction(new window.ide.workbench.syncActionDescriptor(rebuildAction, "sail.rebuild", "Rebuild container", {
            primary: ((1 << 11) >>> 0) | 48 // That's cmd + R. See vscode source for the magic numbers.
        }), "sail: Rebuild container", "sail");

        const statusBarService = window.ide.workbench.statusbarService
        statusBarService.addEntry({
            text: "rebuild",
            tooltip: "Rebuild sail container",
            command: "sail.rebuild"
        }, 0)
    })
}())
