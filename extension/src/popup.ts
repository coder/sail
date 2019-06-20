import { requestSail } from "./common";

const root = document.getElementById("root") as HTMLElement;
// const projects = document.getElementById("projects") as HTMLUListElement;
document.body.style.width = "150px";

requestSail().then((url) => {
	document.body.innerText = "Sail is setup and working properly!";
}).catch((ex) => {
	const has = (str: string) => ex.toString().indexOf(str) !== -1;

	if (has("not found")) {
		document.body.innerText = "After installing sail, run `sail setup-extension`.";
	} else {
		document.body.innerText = ex.toString();
	}
});
