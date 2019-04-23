import { ExtensionMessage, ClientMessage, ServerMessage, requestSail, sailLogoEmbed } from "./common";

const openInSailId = "open-in-sail";

const injectGithub = (): void => {
	const pageActions = document.querySelector(".pagehead-actions");

	{
		if (document.getElementById(openInSailId)) {
			return;
		}

		// Inject top open with Sail button
		const li = document.createElement("li");
		li.id = openInSailId;
		const a = document.createElement("a");
		a.setAttribute("aria-label", "Open repository in Sail");
		a.className = "open-in-sail btn btn-sm tooltipped tooltipped-s";
		a.innerText = "Open in Sail";
		li.appendChild(a);
		pageActions.insertBefore(li, pageActions.children[0]);
		a.addEventListener("click", (event) => {
			event.preventDefault();
			event.stopPropagation();


			if (a.classList.contains("disabled")) {
				return;
			}

			a.classList.add("disabled");
			requestSail({
				type: "run",
				run_event: {
					repo: window.location.href.replace("https://github.com/", ""), // "codercom/code-server",
				},
			}).then((serverMsg) => {
				a.classList.remove("disabled");
				console.log(serverMsg);
			}).catch((ex) => {
				a.classList.remove("disabled");
			});
		});
	}
};

window.addEventListener("load", injectGithub);
window.addEventListener("pjax:end", injectGithub);
