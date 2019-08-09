import {
	sailAvailable,
	getApprovedHosts,
	setApprovedHosts,
	addApprovedHost
} from "./common";
import "./config.scss";

const sailStatus = document.getElementById("sail-status");
const sailAvailableStatus = document.getElementById("sail-available-status");
const approvedHostsEntries = document.getElementById("approved-hosts-entries");
const approvedHostsRemoveError = document.getElementById("approved-hosts-remove-error");
const approvedHostsAdd = document.getElementById("approved-hosts-add");
const approvedHostsAddInput = document.getElementById("approved-hosts-add-input") as HTMLInputElement;
const approvedHostsBadInput = document.getElementById("approved-hosts-bad-input");
const approvedHostsError = document.getElementById("approved-hosts-error");

// Check if the native manifest is installed.
sailAvailable().then(() => {
	sailAvailableStatus.innerText = "Sail is setup and working properly!";
}).catch((ex) => {
	const has = (str: string) => ex.toString().indexOf(str) !== -1;

	sailStatus.classList.add("error");
	let message = "Failed to connect to Sail.";
	if (has("not found") || has("forbidden")) {
		message = "After installing Sail, run <code>sail install-for-chrome-ext</code>.";
	}
	sailAvailableStatus.innerHTML = message;

	const pre = document.createElement("pre");
	pre.innerText = ex.toString();
	sailStatus.appendChild(pre);
});

// Create event listeners to add approved hosts.
approvedHostsAdd.addEventListener("click", (e: Event) => {
	e.preventDefault();
	submitApprovedHost();
});
approvedHostsAddInput.addEventListener("keyup", (e: KeyboardEvent) => {
	if (e.keyCode === 13) {
		e.preventDefault();
		submitApprovedHost();
	}
});
let invalidInputTimeout: number = null;
let errorTimeout: number = null;
const submitApprovedHost = (): Promise<void> => {
	let host = approvedHostsAddInput.value.toLowerCase();
	if (!host) {
		return;
	}

	// Validation logic. Users can put in a full URL or a valid host and it
	// should be parsed successfully.
	const match = host.match(/^\s*(https?:\/\/)?((\.?[a-z\d_-]+)+)(\/.*)?\s*$/);
	if (!match) {
		approvedHostsBadInput.style.display = "block";
		clearTimeout(invalidInputTimeout);
		invalidInputTimeout = setTimeout(() => {
			approvedHostsBadInput.style.display = "none";
		}, 5000);
		return;
	}
	host = match[2];

	return addApprovedHost(host)
		.then(() => {
			approvedHostsAddInput.value = "";
		})
		.catch((ex) => {
			console.error("Failed to add host to approved hosts list.", ex);
			approvedHostsRemoveError.style.display = "block";
			clearTimeout(errorTimeout);
			errorTimeout = setTimeout(() => {
				approvedHostsError.style.display = "none";
			}, 5000);
		})
		.finally(() => {
			reloadApprovedHostsTable()
				.then((hosts) => console.log("Reloaded approved hosts.", hosts))
				.catch((ex) => {
					alert("Failed to reload approved hosts from extension storage.\n\n" + ex.toString());
				});
		});
};

// Handles click events for remove buttons in the approved hosts table.
let removeErrorTimeout: number = null;
const removeBtnHandler = function (e: Event) {
	e.preventDefault();
	const host = this.dataset.host;
	if (!host) {
		return;
	}

	getApprovedHosts()
		.then((hosts) => {
			const index = hosts.indexOf(host);
			if (index > -1) {
				hosts.splice(index, 1);
			}

			return setApprovedHosts(hosts);
		})
		.catch((ex) => {
			console.error("Failed to remove host from approved hosts list.", ex);
			approvedHostsRemoveError.style.display = "block";
			clearTimeout(removeErrorTimeout);
			removeErrorTimeout = setTimeout(() => {
				approvedHostsRemoveError.style.display = "none";
			}, 5000);
		})
		.finally(() => {
			reloadApprovedHostsTable()
				.then((hosts) => console.log("Reloaded approved hosts.", hosts))
				.catch((ex) => {
					alert("Failed to reload approved hosts from extension storage.\n\n" + ex.toString());
				});
		});
};

// Load approved hosts into the table.
const reloadApprovedHostsTable = (): Promise<String[]> => {
	return new Promise<String[]>((resolve, reject) => {
		getApprovedHosts().then((hosts) => {
			// Clear table.
			while (approvedHostsEntries.firstChild) {
				approvedHostsEntries.removeChild(approvedHostsEntries.firstChild);
			}

			if (hosts.length === 0) {
				// No approved hosts.
				const tr = document.createElement("tr");
				const td = document.createElement("td");
				td.colSpan = 2;
				td.innerText = "No approved host entries found.";
				tr.appendChild(td);
				approvedHostsEntries.appendChild(tr);
				return resolve([]);
			}

			for (let host of hosts) {
				host = host.toLowerCase();

				let cells = [] as (HTMLElement|Text)[];
				cells.push(document.createTextNode(host));

				// Remove button. Click event is a reusable
				// function that grabs the host name from
				// btn.dataset.host.
				const removeBtn = document.createElement("button");
				removeBtn.innerText = "Remove";
				removeBtn.classList.add("host-remove-btn");
				removeBtn.dataset.host = host;
				removeBtn.addEventListener("click", removeBtnHandler);
				cells.push(removeBtn);

				// Add the cells to a new row in the table.
				const tr = document.createElement("tr");
				for (let cell of cells) {
					const td = document.createElement("td");
					td.appendChild(cell);
					tr.appendChild(td);
				}
				approvedHostsEntries.appendChild(tr);
			}

			return resolve(hosts);
		}).catch(reject);
	});
};

reloadApprovedHostsTable()
	.then((hosts) => console.log("Loaded approved hosts.", hosts))
	.catch((ex) => {
		alert("Failed to load approved hosts from extension storage.\n\n" + ex.toString());
	});
