+++
type="docs"
title="Environment Editing"
browser_title="Sail - Docs - Environment Editing"
section_order=4
+++

Sail comes out the box with VS Code integrated environment editing.

1. Open the `.sail/Dockerfile` within your editor
1. Make changes
1. Press the `rebuild` button in the workbench.
![rebuild button](/rebuild-button.png)
The UI will freeze as the container is rebuilding.

## Workflow Tips
-  Ctrl+Shift+r also triggers an environment rebuild.
-  Move the active part of the Dockerfile to the bottom. Then, stable parts of your Dockerfile will stay
cached.

## Demo
_Modifying dev environment in real-time_

<video width="900px" controls src="/environment-editing.mp4"></video>
