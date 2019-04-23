import { ClientMessage, ServerMessage, ExtensionMessage } from "./common";

export class SailConnector {
	private port: chrome.runtime.Port;
	private _connected: boolean = false;
	private resolveQueue: Array<(msg: ServerMessage) => void> = [];
	private connectPromise: Promise<void>;

	public connect(): Promise<void> {
		if (this.connectPromise) {
			return this.connectPromise;
		}

		return this.connectPromise = new Promise<void>((resolve, reject) => {
			this.port = chrome.runtime.connectNative("com.coder.sail");
			this.port.onMessage.addListener((message) => {
				if (this.resolveQueue.length > 0) {
					const func = this.resolveQueue.shift();
					func(message);
				} else {
					// Initial msg upon connect
					if (message.type === "active") {
						resolve();
					}
				}
			});
			this.port.onDisconnect.addListener(() => {
				if (chrome.runtime.lastError) {
					return reject(chrome.runtime.lastError.message);
				}
				this._connected = false;
				this.port = undefined;
			});
			this._connected = true;
		});
	}

	public sendMessage(clientMessage: ClientMessage): Promise<ServerMessage> {
		return new Promise<ServerMessage>((resolve, reject) => {
			if (!this._connected) {
				return reject("not connected");
			}

			this.resolveQueue.push(resolve);
			this.port.postMessage(clientMessage);
		});
	}

	public dispose(): void {
		this.port.disconnect();
		this.connectPromise = undefined;
		this._connected = false;
	}
}

const connector = new SailConnector();
let connectError: string | undefined = "Not connected yet";
connector.connect().then(() => connectError = undefined).catch((ex) => {
	connectError = `Failed to connect: ${ex.toString()}`;
});

chrome.runtime.onMessage.addListener((data: ExtensionMessage, sender, sendResponse: (msg: ExtensionMessage) => void) => {
	if (data.type === "sail") {
		if (!data.clientMessage) {
			return sendResponse({
				type: "sail",
				error: "No client message specified",
			});
		}

		if (connectError) {
			return sendResponse({
				type: "sail",
				error: connectError,
			});
		}

		connector.sendMessage(data.clientMessage).then((serverMessage) => {
			sendResponse({
				type: "sail",
				serverMessage,
			});
		});

		return true;
	}

});
