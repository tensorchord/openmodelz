# Bloomz 560M OpenAI Compatible API

This is a simple API that allows you to use the Bloomz 560M as a OpenAI Gym environment.

## Deploy

```bash
$ mdz deploy --image modelzai/llm-bloomz-560m:23.06.13 --name llm
```

### Get the deployment

```bash
$ mdz list
 NAME  ENDPOINT                                      STATUS  REPLICAS 
 llm   http://localhost:31112/inference/llm.default  Ready   1/1      
```

### Test the deployment

```python
import openai
openai.api_base="http://localhost:31112/inference/llm.default"
openai.api_key="any"
openai.debug = True

# create a chat completion
chat_completion = openai.ChatCompletion.create(model="", messages=[
    {"role": "user", "content": "Who are you?"},
    {"role": "assistant", "content": "I am a student"},
    {"role": "user", "content": "What do you learn?"},
    {"role": "assistant", "content": "I learn math"},
    {"role": "user", "content": "Do you like english?"}
], max_tokens=100)
```
