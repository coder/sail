import { ExtensionMessage, WebSocketMessage } from "./common";

export class SailConnector {
	private port: chrome.runtime.Port;
	private connectPromise: Promise<string>;

	public connect(): Promise<string> {
		if (this.connectPromise) {
			return this.connectPromise;
		}

		this.connectPromise = new Promise<string>((resolve, reject) => {
			this.port = chrome.runtime.connectNative("com.coder.sail");
			this.port.onMessage.addListener((message) => {
				if (!message.url) {
					return reject("Invalid handshake message");
				}

				resolve(message.url);
			});
			this.port.onDisconnect.addListener(() => {
				if (chrome.runtime.lastError) {
					this.connectPromise = undefined;

					return reject(chrome.runtime.lastError.message);
				}
				this.port = undefined;
			});
		});

		return this.connectPromise;
	}

	public dispose(): void {
		this.port.disconnect();
		this.connectPromise = undefined;
	}
}

// Get the sail URL.
const connector = new SailConnector();
let connectError: string | undefined = "Not connected yet";
connector.connect().then(() => connectError = undefined).catch((ex) => {
	connectError = `Failed to connect: ${ex.toString()}`;
});

// doConnection attempts to connect to Sail over WebSocket.
const doConnection = (socketUrl: string, projectUrl: string, onMessage: (data: WebSocketMessage) => void): Promise<WebSocket> => {
	return new Promise<WebSocket>((resolve, reject) => {
		const socket = new WebSocket(socketUrl);
		socket.addEventListener("open", () => {
			socket.send(JSON.stringify({
				project: projectUrl,
			}));

			resolve(socket);
		});
		socket.addEventListener("close", (event) => {
			const v = `sail socket was closed: ${event.code}`;
			onMessage({ type: "error", v });
			reject(v);
		});

		socket.addEventListener("message", (event) => {
			const data = JSON.parse(event.data);
			if (!data) {
				return;
			}
			const type = data.type;
			const content = type === "data" ? atob(data.v) : data.v;

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

chrome.runtime.onMessage.addListener((data: ExtensionMessage, sender, sendResponse: (msg: ExtensionMessage) => void) => {
	if (data.type === "sail") {
		if (data.projectUrl) {
			// Launch a sail connection.
			if (!sender.tab) {
				// Only allow from content scripts.
				return;
			}

			// onMessage forwards WebSocketMessages to the tab that
			// launched Sail.
			const onMessage = (message: WebSocketMessage) => {
				chrome.tabs.sendMessage(sender.tab.id, message);
			};
			connector.connect().then((sailUrl) => {
				const socketUrl = sailUrl.replace("http:", "ws:") + "/api/v1/run";
				return doConnection(socketUrl, data.projectUrl, onMessage).then((conn) => {
					sendResponse({
						type: "sail",
					});
				});
			}).catch((ex) => {
				sendResponse({
					type: "sail",
					error: ex.toString(),
				});
			})
		} else {
			// Check if we can get a sail URL.
			connector.connect().then(() => {
				sendResponse({
					type: "sail",
				})
			}).catch((ex) => {
				sendResponse({
					type: "sail",
					error: ex.toString(),
				});
			});
		}

		return true;
	}
});
