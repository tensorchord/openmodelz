<div align="center">

# OpenModelZ

Turn Any Cloud (Or HomeLab) Into Your Personal AI Lab

</div>

<p align=center>
<a href="https://discord.gg/KqswhpVgdU"><img alt="discord invitation link" src="https://dcbadge.vercel.app/api/server/KqswhpVgdU?style=flat"></a>
<a href="https://twitter.com/TensorChord"><img src="https://img.shields.io/twitter/follow/tensorchord?style=social" alt="trackgit-views" /></a>
</p>

OpenModelZ (MDZ) provides a simple CLI to deploy and manage your machine learning workloads on any cloud or home lab.

## Why use MDZ?

OpenModelZ is the ideal solution for practitioners who want to quickly deploy their machine learning models to an endpoint without the hassle of spending excessive time, money, and effort to figure out the entire end-to-end process.

We created OpenModelZ in response to the difficulties of finding a simple, cost-effective way to get models into production fast. Traditional deployment methods can be complex and time-consuming, requiring significant effort and resources to get models up and running.

- Kubernetes: Setting up and maintaining Kubernetes and Kubeflow can be challenging due to their technical complexity. Data scientists spend significant time configuring and debugging infrastructure instead of focusing on model development.
- Managed services: Alternatively, using a managed service like AWS SageMaker can be expensive and inflexible, limiting the ability to customize deployment options.
- Virtual machines: As an alternative, setting up a cloud VM-based solution requires learning complex infrastructure concepts like load balancers, ingress controllers, and other components. This takes a lot of specialized knowledge and resources.

With OpenModelZ, we take care of the underlying technical details for you, and provide a simple and easy-to-use CLI to deploy your models to any cloud (GCP, AWS, or others), your home lab, or even a single machine.

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
