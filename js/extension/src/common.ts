type CommonMessageType = "list" | "error";
type ClientMessageType = CommonMessageType | "run";
type ServerMessageType = CommonMessageType | "success" | "active";

export interface ClientMessage {
	readonly type: ClientMessageType;
	readonly run_event?: {
		readonly repo: string;
	};
}

export interface ServerMessage {
	readonly type: ServerMessageType;
	readonly error_event?: {
		readonly error: string;
	};
	readonly list_event?: {
		readonly projects: ReadonlyArray<{
			readonly name: string;
			readonly url: string;
		}>;
	};
}

export interface ExtensionMessage {
	readonly type: "sail";
	readonly error?: string;
	readonly clientMessage?: ClientMessage;
	readonly serverMessage?: ServerMessage;
}

export const requestSail = (clientMessage: ClientMessage): Promise<ServerMessage> => {
	return new Promise<ServerMessage>((resolve, reject) => {
		chrome.runtime.sendMessage({
			type: "sail",
			clientMessage,
		} as ExtensionMessage, (data: ExtensionMessage) => {
			if (data.type === "sail") {
				if (data.error) {
					return reject(data.error);
				}

				if (!data.serverMessage) {
					return reject("No server message found in response")
				}

				if (data.serverMessage.type === "error") {
					return reject(data.serverMessage.error_event!.error);
				}

				resolve(data.serverMessage);
			} else {
				reject("Invalid data type: " + data.type);
			}
		});
	});
};
