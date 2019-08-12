# Igniter

**Automaticlly add DNS records to Cloudflare**  

Igniter is meant to be deployed inside a Kubernetes Cluster. From there, it watches the `/ingress` endpoint of the Kubernetes API. Upon changes of that ingress it will modify DNS records on Cloudflare accordingly. This way, one can use Cloudflare's Infrastructure without wildcard DNS records.

## How to deploy

In `examples/` one can find a Kubernetes Deployment as reference. One should hold `CLOUDFLARE_API_TOKEN` as a secret in Kubernetes.
