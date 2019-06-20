export interface ExtensionMessage {
	readonly type: "sail";
	readonly error?: string;
	readonly url?: string;
}

export interface SocketMessage {
	readonly type: "init" | "open" | "message";
	readonly data?: string;
}

export const requestSail = (): Promise<string> => {
	return new Promise<string>((resolve, reject) => {
		chrome.runtime.sendMessage({
			type: "sail",
		}, (response) => {
			if (response.type === "sail") {
				if (response.error) {
					return reject(response.error);
				}
				
				resolve(response.url);
			}
		});
	});
};