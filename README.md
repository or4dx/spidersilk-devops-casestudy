
# CSV Processor Service

## Overview

The CSV Processor is a small Golang application that reads the content of a CSV file uploaded from disk, processes the data and returns a cleanly formatted response that is displayed in the user interface. The applications is designed to validate and show case my thought process in implementing the requirements of the case study, and is not intended to be a production ready application.


## Installation

### Prerequisites

- Go 1.21+, 
- Docker, 
- kubectl >1.29+, 
- KOPS >1.29+, 
- Helm 3, 
- Ansible 2.14+, 
- Terraform >1.6+, 
- AWS CLI 2

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port for the App |
| `S3_BUCKET` | `""` | Bucket name, an empty value disables S3 |
| `AWS_REGION` | `us-east-1` | S3 client region |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
  
### Local Development

```bash
cd app && go run .                           # no S3
S3_BUCKET=csv-processor-uploads-dev go run . # with S3
go test ./...
```

### Docker Image

```bash
docker build -t or4dx/csv-processor:latest ./app
docker push or4dx/csv-processor:latest
```

### Terraform (S3 Infrastructure)

```bash
cd terraform && terraform init
terraform apply -var="env=dev"
# Creates: csv-processor-uploads-dev (versioning, AES-256 SSE, lifecycle tiering)
```

### KOPS Cluster
NOTE: the cluster config is untested and may require adjustments. Also the cluster setup is designed for AWS using KOPS, and assumes you have the necessary permissions and configurations to create resources in your AWS account.


```bash
aws s3 mb s3://csv-processor-kops-state
export KOPS_STATE_STORE=s3://csv-processor-kops-state

kops create -f k8s-cluster/cluster.yaml
kops create -f k8s-cluster/ig-masters.yaml
kops create -f k8s-cluster/ig-nodes-spot.yaml
kops create -f k8s-cluster/ig-nodes-ondemand.yaml
kops update cluster csv-processor.k8s.local --yes
kops validate cluster --wait 10m
```

### Deploy (Helm direct)

```bash
helm upgrade --install csv-processor helm/csv-processor \
  --namespace csv-processor \
  --create-namespace \
  --wait \
  --timeout 5m
```

To override values (e.g. S3 bucket):

```bash
helm upgrade --install csv-processor helm/csv-processor \
  --namespace csv-processor \
  --create-namespace \
  --set config.s3Bucket=csv-processor-uploads-dev \
  --wait \
  --timeout 5m
```

### Deploy (Ansible → Helm)

For the purposes of this case study I am checking the ansible vault file to verify the process of securely managing sensitive configuration values locally, but in a production ready scenario I would not use this but likely use a more robust secrets management solution such as HashiCorp Vault or AWS Secrets Manager to manage sensitive configuration values and avoid the need for manual password entry during deployment.

```bash
cd ansible
ansible-playbook playbooks/deploy.yaml \
  --inventory inventory/local.ini \
  --ask-vault-pass \
  -e "aws_account_id=123456789012"
```



