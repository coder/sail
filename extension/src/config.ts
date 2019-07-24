import {
	sailAvailable,
	getApprovedHosts,
	setApprovedHosts,
	addApprovedHost
} from "./common";

const sailAvailableStatus = document.getElementById("sail-available-status");
const approvedHostsEntries = document.getElementById("approved-hosts-entries");
const approvedHostsAdd = document.getElementById("approved-hosts-add");
const approvedHostsAddInput = document.getElementById("approved-hosts-add-input") as HTMLInputElement;

// Check if the native manifest is installed.
sailAvailable().then(() => {
	sailAvailableStatus.innerText = "Sail is setup and working properly!";
}).catch((ex) => {
	const has = (str: string) => ex.toString().indexOf(str) !== -1;

	if (has("not found") || has("forbidden")) {
		sailAvailableStatus.innerText = "After installing Sail, run `sail install-for-chrome-ext`.\n\n" + ex.toString();
	} else {
		sailAvailableStatus.innerText = ex.toString();
	}
});

// Create event listener to add approved hosts.
approvedHostsAdd.addEventListener("click", (e: Event) => {
	e.preventDefault();
	// TODO: safe to lowercase?
	const host = approvedHostsAddInput.value.toLowerCase();
	// TODO: validate here
	if (!host) {
		return;
	}
	console.log(host);

	addApprovedHost(host)
		.then(() => {
			approvedHostsAddInput.value = "";
		})
		.catch((ex) => {
			alert("Failed to add host to approved hosts list.\n\n" + ex.toString());
		})
		.finally(() => {
			reloadApprovedHostsTable();
		});
});

// Handles click events for remove buttons in the approved hosts table.
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
			alert("Failed to remove host from approved hosts list.\n\n" + ex.toString());
		})
		.finally(() => {
			reloadApprovedHostsTable();
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
	// TODO: context
	.catch((ex) => console.error(ex));
