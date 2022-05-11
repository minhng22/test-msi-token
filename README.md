# Authenticate and Authorize member cluster with hub cluster

- Install infrascture

```sh
  $ az aks create --resource-group demo --name hubCluster --node-count 1  --generate-ssh-keys --enable-aad --enable-azure-rbac

  $ az aks create --resource-group demo --name memberCluster --node-count 1  --generate-ssh-keys --enable-managed-identity
```

- Create a namespace for test purpose in hub cluster

```sh
 $ kubectl apply -f namespace.yaml
```

- Create Role and RoleBinding

```sh
 $ kubectl apply -f hub_roles.yaml
```

- Get the member cluster vmss managed identity (*preq:* Install `jq`. An easy option is to install with `homebrew`)
  
  ```sh
  $ az aks show --resource-group demo --name memberCluster | jq .identityProfile.kubeletidentity.clientId
  ```
  Update var `MemberClusterClientId` in `main.go` with this value.

- Get `hubCluster` server api address (an easy option is using Azure portal) and update var `HubServerApiAddress` in `main.go` with this value.
- Create new managed identity Credential using the member cluster clientId
- Use AKS scope `6dae42f8-4368-4678-94ff-3960e28e3630` to create TokenRequestOptions.
- Get the token from the created managed identity.
- Create rest config using the following:
  - BearerToken --> token
  - Host -->   by running `kubectl config view -o jsonpath='{.clusters[0].cluster.server}'`
  - set TLS as insecure = `true`
- Create clientSet using this config
- Do any operation to confirm.

- Build your image

```sh
  docker build  . -t <docker_repo>/member-agent:v1
  docker push <docker_repo>/member-agent:v1
```

- Deploy the agent/app to member cluster

```sh
 $ kubectl apply -f member-agent-deployment.yaml
```

- **Expectations:** 
  - The log should show that access to pods is forbidden for namespace `default'.
  - The log should show that there is 0 pod in `member-a` namespace and no error.
  - Try deploy `pod.yaml` in hub cluster. After successful deployment. The log should show that pod `demo-msi` is found in namespace `member-a`