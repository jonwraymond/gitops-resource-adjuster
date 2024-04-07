# Resource Adjuster Operator

## Overview

The Resource Adjuster Operator is designed to automatically adjust Kubernetes resource requests and limits based on Vertical Pod Autoscaler (VPA) recommendations. It operates by monitoring `ResourceAdjustment` custom resources (CRs) that define mappings between VPA recommendations and specific fields in YAML or JSON files within a GitOps repository.

## Getting Started

### Prerequisites

- A Kubernetes cluster
- `kubectl` configured to interact with your cluster
- Docker for building and pushing the operator image

### Deployment

1. **Build and Push the Docker Image**

   Build the Docker image for the operator and push it to a container registry such as DockerHub.

