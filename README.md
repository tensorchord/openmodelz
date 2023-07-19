<div align="center">

# OpenModelZ

Turn Any Cloud (Or HomeLab) Into Your Personal AI Lab

</div>

<p align=center>
<a href="https://discord.gg/KqswhpVgdU"><img alt="discord invitation link" src="https://dcbadge.vercel.app/api/server/KqswhpVgdU?style=flat"></a>
<a href="https://twitter.com/TensorChord"><img src="https://img.shields.io/twitter/follow/tensorchord?style=social" alt="trackgit-views" /></a>
</p>

OpenModelZ provides a simple CLI to deploy and manage your machine learning workloads on any cloud or home lab.

# Where to use OpenModelZ?

You could use OpenModelZ to:

- Quickly prototype new machine learning models. OpenModelZ allows you to deploy a Gradio or Streamlit application. This lets you focus on experimenting with and improving your models, without getting bogged down in infrastructure.
- Serve and test your models in a production environment. OpenModelZ provides a simple interface for deploying your models in a production environment. It also allows you to easily scale your models up or down based on demand.
- Share your models with teammates or collaborators easily. OpenModelZ abstracts away the complexity of Kubernetes, giving your collaborators a simple way to access and provide feedback on your models.
- Gain insights into your models' performance and reliability. OpenModelZ exposes Prometheus metrics and health checks for your deployed models, providing insight into latency, throughput, errors and other key indicators.

## Quick Start

Once you've installed the `mdz` you can start deploying models and experimenting with them.

### Bootstrap `mdz`

Before you actually start using `mdz`, you need to bootstrap it first.

```bash
$ mdz server start
🚧 Initializing the server...
🚧 Waiting for the server to be ready...
🐋 Checking if the server is running...
Agent:
 Name: 		agent
 Orchestration: kubernetes
 Version: 	v0.0.5
 Build Date: 	2023-07-19T09:12:55Z
 Git Commit: 	84d0171640453e9272f78a63e621392e93ef6bbb
 Git State: 	clean
 Go Version: 	go1.19.10
 Compiler: 	gc
 Platform: 	linux/amd64
🐳 The server is running at http://0.0.0.0:31112
🎉 You could set the environment variable to get started!

export MDZ_AGENT=http://0.0.0.0:31112
```

### Deploy your first applications

Once you've bootstrapped the `mdz` server, you can start deploying your first applications.

```bash
$ mdz deploy --image modelzai/llm-bloomz-560m:23.06.13 --name llm
```

This will deploy the model `modelzai/llm-bloomz-560m:23.06.13` as a serverless function. You can access it at `http://

# Acknowledgements

- [Kubeflow](https://github.com/kubeflow) gives us a lot of insights on simplifying the machine learning deployments.
- [K3s](https://github.com/k3s-io/k3s) for the single control-plane binary and process.
- [OpenFaaS](https://github.com/openfaas) for their work on serverless function services. It laid the foundation for OpenModelZ.
