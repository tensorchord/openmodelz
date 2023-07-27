<div align="center">

# OpenModelZ

Simplify machine learning deployment for any environment.

</div>

<p align=center>
<a href="https://discord.gg/KqswhpVgdU"><img alt="discord invitation link" src="https://dcbadge.vercel.app/api/server/KqswhpVgdU?style=flat"></a>
<a href="https://twitter.com/TensorChord"><img src="https://img.shields.io/twitter/follow/tensorchord?style=social" alt="trackgit-views" /></a>
<a href="https://github.com/tensorchord/openmodelz#contributors-"><img alt="all-contributors" src="https://img.shields.io/github/all-contributors/tensorchord/openmodelz/main"></a>

</p>

OpenModelZ (MDZ) provides a simple CLI to deploy and manage your machine learning workloads on any cloud or home lab.

## Why use OpenModelZ

OpenModelZ is the ideal solution for practitioners who want to quickly deploy their machine learning models to a (public or private) endpoint without the hassle of spending excessive time, money, and effort to figure out the entire end-to-end process.

We created OpenModelZ in response to the difficulties of finding a simple, cost-effective way to get models into production fast. Traditional deployment methods can be complex and time-consuming, requiring significant effort and resources to get models up and running.

- Kubernetes: Setting up and maintaining Kubernetes and Kubeflow can be challenging due to their technical complexity. Data scientists spend significant time configuring and debugging infrastructure instead of focusing on model development.
- Managed services: Alternatively, using a managed service like AWS SageMaker can be expensive and inflexible, limiting the ability to customize deployment options.
- Virtual machines: As an alternative, setting up a cloud VM-based solution requires learning complex infrastructure concepts like load balancers, ingress controllers, and other components. This takes a lot of specialized knowledge and resources.

With OpenModelZ, we take care of the underlying technical details for you, and provide a simple and easy-to-use CLI to deploy your models to **any cloud (GCP, AWS, or others), your home lab, or even a single machine**.

You could **start from a single machine and scale it up to a cluster of machines** without any hassle. OpenModelZ lies at the heart of our [ModelZ](https://modelz.ai), which is a serverless inference platform. It's used in production to deploy models for our customers.

## Quick Start 🚀

Once you've installed the `mdz` you can start deploying models and experimenting with them.

There are only two concepts in `mdz`:

- **Deployment**: A deployment is a running inference service. You could configure the number of replicas, the port, and the image, and some other parameters.
- **Server**: A server is a machine that could run the deployments. It could be a cloud VM, a PC, or even a Raspberry Pi. You could start from a single server and scale it up to a cluster of machines without any hassle.

### Bootstrap `mdz`

It's super easy to bootstrap the `mdz` server. You just need to find a server (could be a cloud VM, a home lab, or even a single machine) and run the `mdz server start` command. The `mdz` server will be bootstrapped on the server and you could start deploying your models.

```
$ mdz server start
🚧 Initializing the server...
🚧 Waiting for the server to be ready...
🐋 Checking if the server is running...
Agent:
 Version:       v0.0.5
 Build Date:    2023-07-19T09:12:55Z
 Git Commit:    84d0171640453e9272f78a63e621392e93ef6bbb
 Git State:     clean
 Go Version:    go1.19.10
 Compiler:      gc
 Platform:      linux/amd64
🐳 The server is running at http://192.168.71.93.nip.io
🎉 You could set the environment variable to get started!

export MDZ_AGENT=http://192.168.71.93.nip.io
```

The internal IP address will be used as the default endpoint of your deployments. You could provide the public IP address of your server to the `mdz server start` command to make it accessible from the outside world.

```bash
# Provide the public IP as an argument
$ mdz server start 1.2.3.4
```

### Create your first deployment

Once you've bootstrapped the `mdz` server, you can start deploying your first applications.

```
$ mdz deploy --image aikain/simplehttpserver:0.1 --name simple-server --port 80
Inference simple-server is created
$ mdz list
 NAME          ENDPOINT                                                    STATUS  REPLICAS 
 simple-server http://simple-server-crxm19s3602d1zlg.192.168.71.93.nip.io   Ready   1/1      
               http://192.168.71.93.nip.io/inference/simple-server.default                               
```

You could access the deployment by visiting the endpoint URL. It will be `http://simple-server-crxm19s3602d1zlg.192.168.71.93.nip.io` in this case. The endpoint could be accessed from the outside world as well if you've provided the public IP address of your server to the `mdz server start` command.

### Scale your deployment

You could scale your deployment by using the `mdz scale` command.

```bash
$ mdz scale simple-server --replicas 3
```

The requests will be load balanced between the replicas of your deployment.

### Debug your deployment

Sometimes you may want to debug your deployment. You could use the `mdz logs` command to get the logs of your deployment.

```bash
$ mdz logs simple-server
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:16] "GET / HTTP/1.1" 200 -
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:16] "GET / HTTP/1.1" 200 -
simple-server-6756dd67ff-4bf4g: 10.42.0.1 - - [27/Jul/2023 02:32:17] "GET / HTTP/1.1" 200 -
```

You could also use the `mdz exec` command to execute a command in the container of your deployment. You do not need to ssh into the server to do that.

```
$ mdz exec simple-server ps
PID   USER     TIME   COMMAND
    1 root       0:00 /usr/bin/dumb-init /bin/sh -c python3 -m http.server 80
    7 root       0:00 /bin/sh -c python3 -m http.server 80
    8 root       0:00 python3 -m http.server 80
    9 root       0:00 ps
$ mdz exec simple-server -ti bash
bash-4.4# uname -r
5.19.0-46-generic
bash-4.4# 
```

Or you could port-forward the deployment to your local machine and debug it locally.

```
$ mdz port-forward simple-server 7860
Forwarding inference simple-server to local port 7860
```

### Add more servers

You could add more servers to your cluster by using the `mdz server join` command. The `mdz` server will be bootstrapped on the server and join the cluster automatically.

```
$ mdz server list
 NAME   PHASE  ALLOCATABLE      CAPACITY        
 node1  Ready  cpu: 16          cpu: 16         
               mem: 32784748Ki  mem: 32784748Ki 
 node2  Ready  cpu: 16          cpu: 16         
               mem: 32784748Ki  mem: 32784748Ki 
```

### Label your servers

You could label your servers to deploy your models to specific servers. For example, you could label your servers with `gpu=true` and deploy your models to servers with GPUs.

```
$ mdz server label node3 gpu=true type=nvidia-a100
$ mdz deploy --image aikain/simplehttpserver:0.1 --name simple-server --port 80 --node-labels gpu=true,type=nvidia-a100
```

## More on documentation 📝

See [OpenModelZ documentation](https://docs.open.modelz.ai/).

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
- [nip.io](https://nip.io/) for the wildcard DNS service. It makes it possible to access the server from the outside world without any setup.
