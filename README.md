# Simple Microservice

```shell
# Build the Docker image
docker build -t my-go-microservice .

# Run the container
docker run -p 8999:8080 my-go-microservice
```

Then visit `localhost:8999` in your browser.

```shell
$ curl localhost:8999
> Order received: 166412168
```
