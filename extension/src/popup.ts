import { sailAvailable } from "./common";

const root = document.getElementById("root") as HTMLElement;
document.body.style.width = "250px";

sailAvailable().then(() => {
	document.body.innerText = "Sail is setup and working properly!";
}).catch((ex) => {
	const has = (str: string) => ex.toString().indexOf(str) !== -1;

	if (has("not found") || has("forbidden")) {
		document.body.innerText = "After installing Sail, run `sail install-for-chrome-ext`.\n\n" + ex.toString();
	} else {
		document.body.innerText = ex.toString();
	}
});
