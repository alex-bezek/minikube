---
title: "ngrok Ingress Controller"
linkTitle: "ngrok Ingress Controller"
weight: 1
date: 2024-03-01
---

## Overview

The ngrok Ingress Controller for Minikube allows you to easily expose your local Kubernetes services to the internet using ngrok. ngrok is a popular tool for creating secure tunnels to localhost, making it an ideal solution for testing and sharing your Minikube services without deploying them to a public cloud.

## Enable

To enable the ngrok Ingress Controller addon in Minikube, use the following command:

```bash
minikube addons enable ngrok-ingress-controller
```

During the enablement process, the addon will check for the presence of the ngrok credentials. If the credentials are not found, you can add them using the `minikube addons configure ngrok-ingress-controller` command.

## Configure

The `configure` command allows you to set up your ngrok credentials and optionally create ingress for existing services in your Minikube cluster. Here's an example of the configuration process:

```bash
minikube addons configure ngrok-ingress-controller
```

You will be prompted to:

- Replace ngrok credentials (if needed).
- Configure ingress for existing services.
- Choose between using a single domain for all services or a unique domain for each service.
- Create a policy module for securing ingress with authentication (optional).
- Create ingress objects for selected services.

Once configured, you can enable the addon, and it will create ingress on the specified domains.

## Test Installation

To verify that the ngrok Ingress Controller is installed and running correctly, you can check the pods in the `ngrok-ingress-controller` namespace:

```bash
kubectl get pods -n ngrok-ingress-controller
```

Ensure that the pods are in the `Ready` state.

## Example/Tutorial: Configure Read-Only Dashboard

To expose the Kubernetes Dashboard through ngrok, follow these steps:

1. Enable the dashboard addon:

   ```bash
   minikube addons enable dashboard
   ```

2. Configure the ngrok Ingress Controller:

   ```bash
   minikube addons configure ngrok-ingress-controller
   ```

3. Enable the ngrok Ingress Controller:

   ```bash
   minikube addons enable ngrok-ingress-controller
   ```

4. Access the provided ngrok URL to view the read-only Kubernetes Dashboard.

## Additional Usage

The ngrok Ingress Controller addon installs an ingress controller in your Minikube cluster. You can create and manage custom ingress and ngrokModuleSets for more advanced use cases.

## Disable

To disable the ngrok Ingress Controller addon, use the following command:

```bash
minikube addons disable ngrok-ingress-controller
```

Before disabling, ensure that you delete any existing ingresses with the `ngrok` ingress class to avoid issues with resource finalizers.
```

Feel free to adjust the content as needed to fit your specific requirements and use cases.