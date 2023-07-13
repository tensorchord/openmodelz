<div align="center">

# OpenModelZ

Serverless Inference for Machine Learning Models

</div>

<p align=center>
<a href="https://discord.gg/KqswhpVgdU"><img alt="discord invitation link" src="https://dcbadge.vercel.app/api/server/KqswhpVgdU?style=flat"></a>
<a href="https://twitter.com/TensorChord"><img src="https://img.shields.io/twitter/follow/tensorchord?style=social" alt="trackgit-views" /></a>
</p>

OpenModelZ simplifies deploying machine learning models in a serverless manner. All it requires is a Docker image containing your model inference server. OpenModelZ then automatically scales your endpoints based on traffic.

# Where to use OpenModelZ?

You could use OpenModelZ to:

- Quickly prototype new machine learning models. OpenModelZ allows you to deploy a Gradio or Streamlit application. This lets you focus on experimenting with and improving your models, without getting bogged down in infrastructure.
- Serve and test your models in a production environment. OpenModelZ provides a simple interface for deploying your models in a production environment. It also allows you to easily scale your models up or down based on demand.
- Share your models with teammates or collaborators easily. OpenModelZ abstracts away the complexity of Kubernetes, giving your collaborators a simple way to access and provide feedback on your models.
- Gain insights into your models' performance and reliability. OpenModelZ exposes Prometheus metrics and health checks for your deployed models, providing insight into latency, throughput, errors and other key indicators.

# Why OpenModelZ?

OpenModelZ offers the following benefits:

**Framework Agnostic**: You can deploy and serve your models using any machine learning framework. OpenModelZ supports frameworks like [text-generation-inference](https://github.com/huggingface/text-generation-inference), [Gradio](https://www.gradio.app/), [FastChat](https://github.com/lm-sys/FastChat), [vllm](https://github.com/vllm-project/vllm), and [Mosec](https://github.com/mosecorg/mosec) without requiring any modifications.

**Model Agnostic**: You could deploy any machine learning models, including but not limited to Stable Diffusion, MPT, FastChat-T5 and many more.

**Request-based autoscaling**: OpenModelZ uses request-based autoscaling to dynamically adjust the amount of GPU resources allocated to your workload based on request volume.

**Simple yet powerful**: OpenModelZ achieves its goals in a simple way without relying on complex systems like Istio and Knative. It provides an easy-to-use interface for deploying and managing your machine learning workloads while still offering powerful features.

## Quick Start

### Installation

### Deploy the inference

Once you've installed the `omz` you can start deploying inferences via the `omz deploy` command:

```bash
# Deploy bloomz with OpenAI compatible API
omz deploy --image modelzai/llm-bloomz-560m:23.06.13
```

The inference will be deployed, and you could use OpenAI python package to interact with it:

```python
import openai
```

# Acknowledgements

- [OpenFaaS](https://github.com/openfaas) for their work on serverless function services. It laid the foundation for OpenModelZ.
- [Kubeflow](https://github.com/kubeflow) gives us a lot of insights on simplifying the machine learning deployments.
