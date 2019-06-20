import { requestSail } from "./common";

const doConnection = (socketUrl: string, projectUrl: string, onMessage: (data: {
	readonly type: "data" | "error";
	readonly v: string;
}) => void): Promise<WebSocket> => {
	return new Promise<WebSocket>((resolve, reject) => {
		const socket = new WebSocket(socketUrl);
		socket.addEventListener("open", () => {
			socket.send(JSON.stringify({
				project: projectUrl,
			}));

			resolve(socket);
		});
		socket.addEventListener("close", (event) => {
			reject(`socket closed: ${event.code}`);
		});

		socket.addEventListener("message", (event) => {
			const data = JSON.parse(event.data);
			if (!data) {
				return;
			}
			const type = data.type;
			const content = atob(data.v);

			switch (type) {
				case "data":
				case "error":
					onMessage({ type, v: content });
					break;
				default:
					throw new Error("unknown message type: " + type);
			}
		});
	});
};

const ensureButton = (): void | HTMLElement => {
	const buttonId = "openinsail";
	const btn = document.querySelector("#" + buttonId) as HTMLElement;
	if (btn) {
		return btn;
	}

	const githubMenu = document.querySelector(".get-repo-select-menu");
	let button: HTMLElement | void;
	if (githubMenu) {
		// GitHub
		button = createGitHubButton();

		githubMenu.parentElement.appendChild(button);

	}
	const gitlabMenu = document.querySelector(".project-repo-buttons") as HTMLElement;
	if (gitlabMenu) {
		// GitLab
		button = createGitLabButton(gitlabMenu);
	}

	if (button) {
		button.id = buttonId;
		button.innerText = "Open in Sail";
		button.title = "Open in Sail";
		button.classList.add("disabled");

		button.addEventListener("click", (event) => {
			event.preventDefault();
			event.stopPropagation();

			const btn = button as HTMLElement;
			if (btn.classList.contains("disabled")) {
				return;
			}

			const cloneUrl = getCloneUrl();
			if (!cloneUrl) {
				return;
			}

			btn.classList.add("disabled");

			const term = document.createElement("div");
			term.style.cssText = `
			border-top-left-radius: 5px;
			position: fixed;
			bottom: 0;
			right: 0;
			width: 35vw;
			height: 40vh;
			background: black;
			padding: 10px;
			padding-top: 30px;
			font-family: monospace;
			white-space: pre;
			overflow-y: auto;
			color: white;
			`;
			const text = document.createElement("div");
			term.appendChild(text);
			document.body.appendChild(term);
			const x = document.createElement("div");
			x.innerText = "X";
			x.style.cssText = `
			position: fixed;
			right: 25px;
			bottom: 37vh;
			color: white;
			font-weight: bold;
			cursor: pointer;
			`;
			x.addEventListener("click", () => {
				term.remove();
			});
			x.title = "Close";
			term.appendChild(x);

			requestSail().then((socketUrl) => {
				return doConnection(socketUrl.replace("http:", "ws:") + "/api/v1/run", cloneUrl, (data) => {
					if (data.type === "data") {
						text.innerText += data.v;
						term.scrollTop = term.scrollHeight;
					}
				});
			}).then((socket) => {
				socket.addEventListener("close", () => {
					btn.innerText = "Open in Sail";
					btn.classList.remove("disabled");
					term.remove();
				});
			}).catch((ex) => {
				btn.innerText = ex.toString();
				setTimeout(() => {
					btn.innerText = "Open in Sail";
					btn.classList.remove("disabled");
					term.remove();
				}, 5000);
			});
		});

		requestSail().then(() => (button as HTMLElement).classList.remove("disabled"))
			.catch((ex) => {
				if (ex.toString().indexOf("host not found") !== -1) {
					(button as HTMLElement).style.opacity = "0.5";
					(button as HTMLElement).title = "Setup Sail using the extension icon in the top-right!";
				}
			});
	}

	return button;
};

const getCloneUrl = (): void | string => {
	const gitlabInput = document.getElementById("http_project_clone") as HTMLInputElement;
	if (gitlabInput) {
		return gitlabInput.value;
	}
	const githubInput = document.querySelector(".https-clone-options .input-group input") as HTMLInputElement;
	if (githubInput) {
		return githubInput.value;
	}
};

const createGitHubButton = (): HTMLElement => {
	const a = document.createElement("a");
	a.className = "open-in-sail btn btn-sm";
	a.style.cssText = `
		background: linear-gradient(180deg,#5883ff 0%,#344fd4 100%);
		color: white;
		margin-left: 10px;
	`;
	return a;
};

const createGitLabButton = (parent: HTMLElement): HTMLElement => {
	const wrapper = document.createElement("div");
	wrapper.className = "project-clone-holder";
	const a = document.createElement("a");
	a.href = "#";
	a.className = "open-in-sail btn btn-primary btn-xs";
	a.style.lineHeight = "18px";
	a.style.marginLeft = "10px";
	wrapper.appendChild(a);
	parent.appendChild(wrapper);
	return a;
};

window.addEventListener("load", ensureButton);
window.addEventListener("pjax:end", ensureButton);
