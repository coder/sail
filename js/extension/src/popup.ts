import { requestSail } from "./common";

const root = document.getElementById("root") as HTMLElement;
// const projects = document.getElementById("projects") as HTMLUListElement;

console.log("We request sail");
requestSail({
	type: "list",
}).then((msg) => {
	console.log("WE GOT MSG", msg);
	msg.list_event!.projects.forEach((project) => {
		const li = document.createElement("li");
		li.innerHTML = project.name;
		root.appendChild(li);
	});
}).catch((ex) => {
	document.body.innerText = ex.toString();
});
