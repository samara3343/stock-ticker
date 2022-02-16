# Stock ticker

Stock ticker is a web server that serves stock data for last N days for a specified stock.

## Prerequisites

* go 1.17+
* docker

## How to build?

using docker

```shell
docker build -t your-org/stock-ticker -t your-org/stock-ticker:v1.0.0 .
```

using local machine

```shell
go build -mod=vendor -o=./bin/stock-ticker .
```

## How to publish the image?

Once the docker image is build you can push the image to a public or private registry.

* Tag your build with a correct version such as v1.0.0.
* Make sure you're authorised to push to a registry (docker login).
* Push the image.

```shell
docker build -t your-org/stock-ticker:v1.0.0 -t your-org/stock-ticker:latest .
docker push your-org/stock-ticker:v1.0.0
docker push your-org/stock-ticker:latest
```

## How to run?

### Docker

The app requires `ALPHA_VANTAGE_KEY_NAME`, `ALPHA_VANTAGE_KEY_VALUE`, `ALPHA_VANTAGE_URL`, `QUERY_STOCK_SYMBOL`,
`QUERY_NDAYS` environment variables setto run successfully.

Example:

```shell 
docker run --rm \
-e ALPHA_VANTAGE_KEY_NAME=<value> \
-e ALPHA_VANTAGE_KEY_VALUE=<value> \
-e ALPHA_VANTAGE_URL=<value> \
-e QUERY_STOCK_SYMBOL=<value> \
-e QUERY_NDAYS=<value> \
-p 8080:8080 your-org/stock-ticker:latest serve
```

Run the following script for the details

```shell
docker run --rm  your-org/stock-ticker:latest serve --help
```

### Minikube

Push the image to the Minikube cache

```shell
docker build -t your-org/stock-ticker .
minikube cache add your-org/stock-ticker:latest
```

Deploy the application

Make sure ./.kube/secrets/ALPHA_VANTAGE_KEY_VALUE file exists and contains a correct API key (ensure no new line added
at the end)

```shell
kubectl create secret generic stock-ticker-secret --from-file=./.kube/secrets/ALPHA_VANTAGE_KEY_VALUE -o yaml --dry-run | kubectl apply -f -
kubectl apply -f ./.kube
```

If running minikube the following command can be run to get the minikube IP

```shell
minikube ip
```

Test the service, use the IP from the previous command

```shell
curl http://<minikube-ip>/api/v1/stock-closing-price-info

```

## API

The application exposes a few API endpoints

* /api/health

A health endpoint, returns 200 when the server has booted.

* /api/v1/stock-closing-price-info

An endpoint that returns the information about a specific stock for the past N days. Response example:

```json
{
  "prices": [
    {
      "date": "2022-02-08",
      "price": "304.56"
    },
    {
      "date": "2022-02-07",
      "price": "300.95"
    },
    {
      "date": "2022-02-04",
      "price": "305.94"
    },
    {
      "date": "2022-02-03",
      "price": "301.25"
    },
    {
      "date": "2022-02-02",
      "price": "313.46"
    },
    {
      "date": "2022-02-01",
      "price": "308.76"
    },
    {
      "date": "2022-01-31",
      "price": "310.98"
    }
  ],
  "averagePrice": "306.5571"
}
```