# Contributing to Druid Operator

First off, thanks for taking the time to contribute to the Druid Operator! 🎉

The following is a set of guidelines for contributing to Druid Operator and its packages. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

## Table of Contents

- [Prerequisites](#prerequisites)
  - [Windows](#windows)
  - [MacOS](#macos)
  - [Linux](#linux)
- [Development Setup](#development-setup)
  - [1. Fork and Clone](#1-fork-and-clone)
  - [2. Create a Local Cluster](#2-create-a-local-cluster)
  - [3. Install Dependencies](#3-install-dependencies)
  - [4. Run the Operator](#4-run-the-operator)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [End-to-End Tests](#end-to-end-tests)
- [Project Structure](#project-structure)

## Prerequisites

You will need the following tools installed on your development machine:

*   **Go** (v1.20+)
*   **Docker** (v20.10+)
*   **Kind** (v0.20+)
*   **Kubectl** (latest)
*   **Helm** (v3+)
*   **Make**

### Windows

**Recommended**: Use Docker Desktop with WSL 2 backend.

run the following commands in PowerShell (Admin):

```powershell
# Install core tools
winget install -e --id GoLang.Go
winget install -e --id Docker.DockerDesktop
winget install -e --id Kubernetes.kind
winget install -e --id Kubernetes.kubectl
winget install -e --id Helm.Helm
winget install -e --id GnuWin32.Make
```

### MacOS

Using [Homebrew](https://brew.sh/):

```bash
brew install go
brew install --cask docker
brew install kind
brew install kubectl
brew install helm
brew install make
```

### Linux

Using `apt` (Ubuntu/Debian) or `brew` (Linuxbrew):

```bash
# Using Linuxbrew (Recommended for unified versioning)
brew install go kind kubectl helm make

# OR using apt (Ubuntu)
sudo apt update
sudo apt install -y golang-go make
# For Docker, Kind, Kubectl, and Helm, please refer to their official installation guides 
# as apt repositories might lag behind.
```

## Development Setup

### 1. Fork and Clone

1.  Fork the [druid-operator repository](https://github.com/datainfrahq/druid-operator) on GitHub.
2.  Clone your fork locally:

```bash
git clone https://github.com/<your-username>/druid-operator.git
cd druid-operator
```

### 2. Create a Local Cluster

We use **Kind** (Kubernetes in Docker) for local development.

```bash
kind create cluster --name druid
```

### 3. Install Dependencies

Deploy the Druid Operator using Helm to set up CRDs and basic resources.

```bash
# Add the DataInfra Helm repo
helm repo add datainfra https://charts.datainfra.io
helm repo update

# Install the operator (this installs CRDs and the controller)
helm -n druid-operator-system upgrade -i --create-namespace cluster-druid-operator datainfra/druid-operator
```

### 4. Run the Operator

You can run the operator source code locally against your Kind cluster. This is useful for rapid development without building Docker images for every change.

```bash
# Verify you are pointing to the correct context
kubectl config use-context kind-druid

# Run the controller locally
make run
```

The operator logs will appear in your terminal.

## Testing

### Unit Tests

Run the unit tests to verify your changes.

```bash
make test
```

### End-to-End Tests

To run the full end-to-End suite (this spins up a Kind cluster and runs validation):

```bash
make e2e
```

## Project Structure

*   `apis/`: Kubernetes API definitions (CRDs).
*   `controllers/`: Core controller logic using Kubebuilder.
*   `chart/`: Helm chart for the operator.
*   `e2e/`: End-to-End test scripts and configurations.
*   `docs/`: Documentation files.
*   `Makefile`: Build and test automation commands.
