<div align="center">

# OpenModelZ

One-click machine learning deployment at scale on any cluster (GCP, AWS, Lambda labs, your home lab, or even a single machine)
</div>

<p align=center>
<a href="https://discord.gg/KqswhpVgdU"><img alt="discord invitation link" src="https://dcbadge.vercel.app/api/server/KqswhpVgdU?style=flat"></a>
<a href="https://twitter.com/TensorChord"><img src="https://img.shields.io/twitter/follow/tensorchord?style=social" alt="trackgit-views" /></a>
<a href="https://docs.open.modelz.ai"><img src="https://img.shields.io/badge/docs.open.modelz.ai-455946.svg?style=socail&logo=googlechrome&logoColor=white" alt="docs" /></a>
<a href="https://github.com/tensorchord/openmodelz#contributors-"><img alt="all-contributors" src="https://img.shields.io/github/all-contributors/tensorchord/openmodelz/main"></a>
</p>

## Why use OpenModelZ

OpenModelZ is the ideal solution for practitioners who want to quickly deploy their machine learning models to a (public or private) endpoint without the hassle of spending excessive time, money, and effort to figure out the entire end-to-end process.

We created OpenModelZ in response to the difficulties of finding a simple, cost-effective way to get models into production fast. Traditional deployment methods can be complex and time-consuming, requiring significant effort and resources to get models up and running.

- Kubernetes: Setting up and maintaining Kubernetes and Kubeflow can be challenging due to their technical complexity. Data scientists spend significant time configuring and debugging infrastructure instead of focusing on model development.
- Managed services: Alternatively, using a managed service like AWS SageMaker can be expensive and inflexible, limiting the ability to customize deployment options.
- Virtual machines: As an alternative, setting up a cloud VM-based solution requires learning complex infrastructure concepts like load balancers, ingress controllers, and other components. This takes a lot of specialized knowledge and resources.

With OpenModelZ, we take care of the underlying technical details for you, and provide a simple and easy-to-use CLI to deploy your models to **any cloud (GCP, AWS, or others), your home lab, or even a single machine**.

You could **start from a single machine and scale it up to a cluster of machines** without any hassle. Besides this, We **provision a separate subdomain for each deployment** without any extra cost and effort, making each deployment easily accessible from the outside.

OpenModelZ forms the core of our [ModelZ](https://modelz.ai) platform, which is a serverless machine learning inference service. It is utilized in a production environment to provision models for our clients.

## Quick Start 🚀

### Install `mdz`

You can install OpenModelZ using the following command:

```text copy
pip install openmodelz
```

You could verify the installation by running the following command:

```text copy
mdz
```

Once you've installed the `mdz` you can start deploying models and experimenting with them.

### Bootstrap `mdz`

It's super easy to bootstrap the `mdz` server. You just need to find a server (could be a cloud VM, a home lab, or even a single machine) and run the `mdz server start` command.

> Notice: We may require the root permission to bootstrap the `mdz` server on port 80.

```
$ mdz server start
🚧 Creating the server...
🚧 Initializing the load balancer...
🚧 Initializing the GPU resource...
🚧 Initializing the server...
🚧 Waiting for the server to be ready...
🐋 Checking if the server is running...
Agent:
 Version:       v0.0.13
 Build Date:    2023-07-19T09:12:55Z
 Git Commit:    84d0171640453e9272f78a63e621392e93ef6bbb
 Git State:     clean
 Go Version:    go1.19.10
 Compiler:      gc
 Platform:      linux/amd64
🐳 The server is running at http://192.168.71.93.modelz.live
🎉 You could set the environment variable to get started!

export MDZ_URL=http://192.168.71.93.modelz.live
```

The internal IP address will be used as the default endpoint of your deployments. You could provide the public IP address of your server to the `mdz server start` command to make it accessible from the outside world.

```bash
# Provide the public IP as an argument
$ mdz server start 1.2.3.4
```

You could also specify the registry mirror to speed up the image pulling process. Here is an example:

```bash /--mirror-endpoints/
$ mdz server start --mirror-endpoints https://docker.mirrors.sjtug.sjtu.edu.cn
```

### Create your first UI-based deployment

Once you've bootstrapped the `mdz` server, you can start deploying your first applications. We will use jupyter notebook as an example in this tutorial. You could use any docker image as your deployment.

```text
$ mdz deploy --image jupyter/minimal-notebook:lab-4.0.3 --name jupyter --port 8888 --command "jupyter notebook --ip='*' --NotebookApp.token='' --NotebookApp.password=''"
Inference jupyter is created
$ mdz list
 NAME     ENDPOINT                                                   STATUS  INVOCATIONS  REPLICAS
 jupyter  http://jupyter-9pnxdkeb6jsfqkmq.192.168.71.93.modelz.live  Ready           488  1/1
          http://192.168.71.93/inference/jupyter.default                                                                         
```

You could access the deployment by visiting the endpoint URL. The endpoint will be automatically generated for each deployment with the following format: `<name>-<random-string>.<ip>.modelz.live`.

It is `http://jupyter-9pnxdkeb6jsfqkmq.192.168.71.93.modelz.live` in this case. The endpoint could be accessed from the outside world as well if you've provided the public IP address of your server to the `mdz server start` command. 

![jupyter notebook](./images/jupyter.png)

### Create your first OpenAI compatible API server

You could also create API-based deployments. We will use [OpenAI compatible API server with Bloomz 560M](https://github.com/tensorchord/modelz-llm#run-the-self-hosted-api-server) as an example in this tutorial.

```text
$ mdz deploy --image modelzai/llm-bloomz-560m:23.07.4 --name simple-server
Inference simple-server is created
$ mdz list
 NAME           ENDPOINT                                                         STATUS  INVOCATIONS  REPLICAS 
 jupyter        http://jupyter-9pnxdkeb6jsfqkmq.192.168.71.93.modelz.live        Ready           488  1/1      
                http://192.168.71.93/inference/jupyter.default                                                 
 simple-server  http://simple-server-lagn8m9m8648q6kx.192.168.71.93.modelz.live  Ready             0  1/1      
                http://192.168.71.93/inference/simple-server.default                                           
```

You could use OpenAI python package and the endpoint `http://simple-server-lagn8m9m8648q6kx.192.168.71.93.modelz.live` in this case, to interact with the deployment.

```python
import openai
openai.api_base="http://simple-server-lagn8m9m8648q6kx.192.168.71.93.modelz.live"
openai.api_key="any"

# create a chat completion
chat_completion = openai.ChatCompletion.create(model="bloomz", messages=[
    {"role": "user", "content": "Who are you?"},
    {"role": "assistant", "content": "I am a student"},
    {"role": "user", "content": "What do you learn?"},
], max_tokens=100)
```

### Scale your deployment

You could scale your deployment by using the `mdz scale` command.

```text /scale/
$ mdz scale simple-server --replicas 3
```

The requests will be load balanced between the replicas of your deployment. 

You could also tell the `mdz` to **autoscale your deployment** based on the inflight requests. Please check out the [Autoscaling](https://docs.open.modelz.ai/deployment/autoscale) documentation for more details.

### Debug your deployment

Sometimes you may want to debug your deployment. You could use the `mdz logs` command to get the logs of your deployment.

```text /logs/
$ mdz logs simple-server
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:16] "GET / HTTP/1.1" 200 -
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:16] "GET / HTTP/1.1" 200 -
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:17] "GET / HTTP/1.1" 200 -
```

You could also use the `mdz exec` command to execute a command in the container of your deployment. You do not need to ssh into the server to do that.

```text /exec/
$ mdz exec simple-server ps
PID   USER     TIME   COMMAND
    1 root       0:00 /usr/bin/dumb-init /bin/sh -c python3 -m http.server 80
    7 root       0:00 /bin/sh -c python3 -m http.server 80
    8 root       0:00 python3 -m http.server 80
    9 root       0:00 ps
```

```text /exec/
$ mdz exec simple-server -ti bash
bash-4.4# 
```

Or you could port-forward the deployment to your local machine and debug it locally.

```text /port-forward/
$ mdz port-forward simple-server 7860
Forwarding inference simple-server to local port 7860
```

### Add more servers

You could add more servers to your cluster by using the `mdz server join` command. The `mdz` server will be bootstrapped on the server and join the cluster automatically.

```text /join/
$ mdz server join <internal ip address of the previous server>
$ mdz server list
 NAME   PHASE  ALLOCATABLE      CAPACITY        
 node1  Ready  cpu: 16          cpu: 16         
               mem: 32784748Ki  mem: 32784748Ki 
               gpu: 1           gpu: 1      
 node2  Ready  cpu: 16          cpu: 16         
               mem: 32784748Ki  mem: 32784748Ki 
               gpu: 1           gpu: 1      
```

### Label your servers

You could label your servers to deploy your models to specific servers. For example, you could label your servers with `gpu=true` and deploy your models to servers with GPUs.

```text /--node-labels gpu=true,type=nvidia-a100/
$ mdz server label node3 gpu=true type=nvidia-a100
$ mdz deploy ... --node-labels gpu=true,type=nvidia-a100
```

## Roadmap 🗂️

Please checkout [ROADMAP](https://docs.open.modelz.ai/community).

## Contribute 😊

We welcome all kinds of contributions from the open-source community, individuals, and partners.

- Join our [discord community](https://discord.gg/KqswhpVgdU)!

## Contributors ✨

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/gaocegege"><img src="https://avatars.githubusercontent.com/u/5100735?v=4?s=70" width="70px;" alt="Ce Gao"/><br /><sub><b>Ce Gao</b></sub></a><br /><a href="https://github.com/tensorchord/openmodelz/commits?author=gaocegege" title="Code">💻</a> <a href="https://github.com/tensorchord/openmodelz/pulls?q=is%3Apr+reviewed-by%3Agaocegege" title="Reviewed Pull Requests">👀</a> <a href="#tutorial-gaocegege" title="Tutorials">✅</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/VoVAllen"><img src="https://avatars.githubusercontent.com/u/8686776?v=4?s=70" width="70px;" alt="Jinjing Zhou"/><br /><sub><b>Jinjing Zhou</b></sub></a><br /><a href="#question-VoVAllen" title="Answering Questions">💬</a> <a href="https://github.com/tensorchord/openmodelz/issues?q=author%3AVoVAllen" title="Bug reports">🐛</a> <a href="#ideas-VoVAllen" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://blog.mapotofu.org/"><img src="https://avatars.githubusercontent.com/u/12974685?v=4?s=70" width="70px;" alt="Keming"/><br /><sub><b>Keming</b></sub></a><br /><a href="https://github.com/tensorchord/openmodelz/commits?author=kemingy" title="Code">💻</a> <a href="#design-kemingy" title="Design">🎨</a> <a href="#infra-kemingy" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tddschn"><img src="https://avatars.githubusercontent.com/u/45612704?v=4?s=70" width="70px;" alt="Teddy Xinyuan Chen"/><br /><sub><b>Teddy Xinyuan Chen</b></sub></a><br /><a href="https://github.com/tensorchord/openmodelz/commits?author=tddschn" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://xuanwo.io/"><img src="https://avatars.githubusercontent.com/u/5351546?v=4?s=70" width="70px;" alt="Xuanwo"/><br /><sub><b>Xuanwo</b></sub></a><br /><a href="#content-Xuanwo" title="Content">🖋</a> <a href="#design-Xuanwo" title="Design">🎨</a> <a href="#ideas-Xuanwo" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/cutecutecat"><img src="https://avatars.githubusercontent.com/u/19801166?v=4?s=70" width="70px;" alt="cutecutecat"/><br /><sub><b>cutecutecat</b></sub></a><br /><a href="#ideas-cutecutecat" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://xieydd.github.io/"><img src="https://avatars.githubusercontent.com/u/20329697?v=4?s=70" width="70px;" alt="xieydd"/><br /><sub><b>xieydd</b></sub></a><br /><a href="#ideas-xieydd" title="Ideas, Planning, & Feedback">🤔</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

## Acknowledgements 🙏

- [K3s](https://github.com/k3s-io/k3s) for the single control-plane binary and process.
- [OpenFaaS](https://github.com/openfaas) for their work on serverless function services. It laid the foundation for OpenModelZ.
- [sslip.io](https://github.com/cunnie/sslip.io) for the wildcard DNS service. It makes it possible to access the server from the outside world without any setup.
