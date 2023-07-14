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

Once you've installed the `omz` you can start deploying models and experimenting with them.

### Deploy OpenAI API compatible inferences

You could deploy OpenAI API compatible inferences with `omz deploy openai-chat` command. A OpenAI API compatible server will be deployed with the model serverlessly.

```bash
# Deploy bloomz with OpenAI compatible API
omz deploy openai-chat --model bloomz-560m
```

After that, you could use `omz list` to check the status of your deployment. And you could use `omz infer openai-chat bloomz` to experiment with it.

```
$ omz list
$ omz infer openai-chat bloomz --interactive
> user: Hello, who are you?
> bloomz: I am an AI. How can I help you today?
...
```

Besides, you could use OpenAI python package to interact with the deployed model.

```python
import openai
openai.api_base="<your agent url>/inference/bloomz.default"
openai.api_key="any"
openai.debug = True

# create a chat completion
chat_completion = openai.ChatCompletion.create(model="", messages=[
    {"role": "user", "content": "Who are you?"},
    {"role": "assistant", "content": "I am a student"},
    {"role": "user", "content": "What do you learn?"},
], max_tokens=100)
```

### Deploy Civitai models

You could deploy Civitai models with `omz deploy civitai` command. A stable diffusion web ui will be deployed with the model serverlessly.

```bash
# Deploy stable diffusion web ui with base models on civitai
omz deploy civitai https://civitai.com/models/25694 --name epicrealism
```

After that, you could use `omz infer civitai epicrealism` to experiment with it.

```bash
omz infer civitai epicrealism --prompt "A photo of a cat"
```

### Deploy Huggingface spaces

You could deploy Huggingface spaces with `omz deploy huggingface` command. A Huggingface spaces will be deployed with the model serverlessly.

```bash
# Deploy Huggingface space application.
omz deploy huggingface Manjushri/Music-Genie-GPU --name music-genie
```

### Share any deployed model with your teammates

You could share your deployed models with your teammates with `omz share` command. A shareable link will be generated for your teammates to access your deployed models.

```bash
# Share your deployed models with your teammates
omz share bloomz
https://3860-101-87-90-254.ngrok.io -> bloomz
```

### Local experiment

OpenModelZ runs your models in your cluster by default. But you could also run your models locally with docker.

```bash
omz local-run openai-chat bloomz
```

### Observe your models

You could use `omz logs` to get the logs.

```bash
omz logs bloomz
```

# Acknowledgements

- [OpenFaaS](https://github.com/openfaas) for their work on serverless function services. It laid the foundation for OpenModelZ.
- [Kubeflow](https://github.com/kubeflow) gives us a lot of insights on simplifying the machine learning deployments.
- [LocalAI](https://github.com/go-skynet/LocalAI) for their work on OpenAI API compatible inferences.
