import { ExtensionMessage } from "./common";

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
					return reject("Invalid handshaking message");
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

const connector = new SailConnector();
let connectError: string | undefined = "Not connected yet";
connector.connect().then(() => connectError = undefined).catch((ex) => {
	connectError = `Failed to connect: ${ex.toString()}`;
});

chrome.runtime.onMessage.addListener((data: ExtensionMessage, sender, sendResponse: (msg: ExtensionMessage) => void) => {
	if (data.type === "sail") {
		connector.connect().then((url) => {
			sendResponse({
				type: "sail",
				url,
			})
		}).catch((ex) => {
			sendResponse({
				type: "sail",
				error: ex.toString(),
			});
		});

		return true;
	}
});
