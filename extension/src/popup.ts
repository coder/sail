import { sailAvailable } from "./common";

const status = document.getElementById("sail-status");
const error = document.getElementById("sail-error");

sailAvailable().then(() => {
	status.innerText = "Sail is setup and working properly!";
}).catch((ex) => {
	const has = (str: string) => ex.toString().indexOf(str) !== -1;

	status.innerText = "Failed to check if Sail is available.";
	if (has("not found") || has("forbidden")) {
		status.innerText += " After installing Sail, run `sail install-for-chrome-ext`.";
	}

	error.innerText = ex.toString();
	error.style.display = "block";
});
