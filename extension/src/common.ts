export interface ExtensionMessage {
	readonly type: "sail";
	readonly error?: string;
	readonly projectUrl?: string;
}

export interface WebSocketMessage {
	readonly type: string;
	readonly v: any;
}

export const launchSail = (projectUrl: string, onMessage: (WebSocketMessage) => void): Promise<void> => {
	const listener = (message: any) => {
		if (message.type && message.v) {
			onMessage(message);
		}
	};
	chrome.runtime.onMessage.addListener(listener);

	return new Promise<void>((resolve, reject) => {
		chrome.runtime.sendMessage({
			type: "sail",
			projectUrl: projectUrl,
		}, (response: ExtensionMessage) => {
			if (response.type === "sail") {
				if (response.error) {
					chrome.runtime.onMessage.removeListener(listener);
					return reject(response.error);
				}

				resolve();
			}
		});
	});
};

export const sailAvailable = (): Promise<void> => {
	return new Promise<void>((resolve, reject) => {
		chrome.runtime.sendMessage({
			type: "sail",
		}, (response: ExtensionMessage) => {
			if (response.type === "sail") {
				if (response.error) {
					return reject(response.error);
				}

				resolve();
			}
		});
	});
};
