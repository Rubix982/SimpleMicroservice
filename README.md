# Simple Microservice

## Docker

```shell
# Build the Docker image
docker build -t my-go-microservice .

# Run the container
docker run -p 8999:8080 my-go-microservice
```

Then visit `localhost:8999` in your browser.

```shell
$ curl localhost:8999/order
> Order received: 166412168
```

## Kubernetes

```shell
cd k8s/ && kubectl apply -f deployment.yaml
```

Then visit `localhost:8999` in your browser.

```shell
$ curl http://localhost:30001/order
> Order received: 274822750
```
