# ali-fc-webhook

update ali fc function by webhook

# Usage

Github Actions will build docker image and push to Aliyun ACR automatically, use Custom Container to deploy this service. And call the http trigger to update your functions.

1. push code by git
2. Github Actions build image and push to registry
3. registry will call this service by webhook trigger.
4. this service update specific fc function with new image.

ps: this function can even update itself.
