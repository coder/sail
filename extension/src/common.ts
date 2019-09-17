// approvedHostsKey is the key in extension storage used for storing the
// string[] containing hosts approved by the user. For versioning purposes, the
// number at the end of the key should be incremented if the method used to
// store approved hosts changes.
export const approvedHostsKey = "approved_hosts_0";

// defaultApprovedHosts is the default approved hosts list. This list should
// only include GitHub.com, GitLab.com, BitBucket.com, etc.
export const defaultApprovedHosts = [
	".github.com",
	".gitlab.com",
	//".bitbucket.com",
];

// ExtensionMessage is used for communication within the extension.
export interface ExtensionMessage {
	readonly type: "sail";
	readonly error?: string;
	readonly projectUrl?: string;
}

// WebSocketMessage is a message from sail itself, sent over the WebSocket
// connection.
export interface WebSocketMessage {
	readonly type: string;
	readonly v: any;
}

// launchSail starts an instance of sail and instructs it to launch the
// specified project URL. Terminal output will be sent to the onMessage handler.
export const launchSail = (projectUrl: string, onMessage: (WebSocketMessage) => void): Promise<void> => {
	return new Promise<void>((resolve, reject) => {
		const port = chrome.runtime.connect();
		port.onMessage.addListener((message: WebSocketMessage): void => {
			if (message.type && message.v) {
				onMessage(message);
			}
			if (message.type === "error") {
				port.disconnect();
			}
		});

		const responseListener = (response: ExtensionMessage): void => {
			if (response.type === "sail") {
				port.onMessage.removeListener(responseListener);
				if (response.error) {
					return reject(response.error);
				}

				resolve();
			}
		};

		port.onMessage.addListener(responseListener);
		port.postMessage({
			type: "sail",
			projectUrl: projectUrl,
		});
	});
};

// sailAvailable resolves if the native host manifest is available and allows
// the extension to connect to Sail. This does not attempt a connection to Sail.
export const sailAvailable = (): Promise<void> => {
	return new Promise<void>((resolve, reject) => {
		const port = chrome.runtime.connect();

		const responseListener = (response: ExtensionMessage): void => {
			if (response.type === "sail") {
				port.onMessage.removeListener(responseListener);
				port.disconnect();
				if (response.error) {
					return reject(response.error);
				}

				resolve();
			}
		};

		port.onMessage.addListener(responseListener);
		port.postMessage({
			type: "sail",
		});
	});
};

// getApprovedHosts gets the approved hosts list from storage.
export const getApprovedHosts = (): Promise<string[]> => {
	return new Promise((resolve, reject) => {
		chrome.storage.sync.get(approvedHostsKey, (items) => {
			if (chrome.runtime.lastError) {
				return reject(chrome.runtime.lastError.message);
			}

			if (!Array.isArray(items[approvedHostsKey])) {
				// No approved hosts.
				return resolve(defaultApprovedHosts);
			}

			resolve(items[approvedHostsKey]);
		});
	});
};

// setApprovedHosts sets the approved hosts key in storage. No validation is
// performed.
export const setApprovedHosts = (hosts: string[]): Promise<void> => {
	return new Promise((resolve, reject) => {
		chrome.storage.sync.set({ [approvedHostsKey]: hosts }, () => {
			if (chrome.runtime.lastError) {
				return reject(chrome.runtime.lastError.message);
			}

			resolve();
		});
	});
};

// addApprovedHost adds a single host to the approved hosts list. No validation
// (except duplicate entry checking) is performed. The host is lowercased
// automatically.
export const addApprovedHost = async (host: string): Promise<void> => {
	host = host.toLowerCase();

	// Check for duplicates.
	let hosts = await getApprovedHosts();
	if (hosts.includes(host)) {
		return;
	}

	// Add new host and set approved hosts.
	hosts.push(host);
	await setApprovedHosts(hosts);
};
