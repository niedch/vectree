---
id: 2020-06-30_ae-on-microk8s
aliases: []
tags:
  - AutomationEngine
  - Kubernetes
---
# Installation Guide for AE on Microk8s

> [!info]
Tested on Ubuntu:18.04

This installation guide includes setup of:
- [Microk8s setup](#microk8s-setup)
- [Pull secret creation](#pull-secret-creation)
- [Ingress routes](#[optional]-Ingress-routes)
- [Helm install of AE](#run-helm-install)
- [Customize installation](#customize-installation)
- [TD:DR;](#tl:dr;)

## Microk8s setup
For the microk8s i use **snap**.

```bash
sudo snap install microk8s --classic
```

### Microk8s plugins
Microk8s has a plugin support, the same can be accomplished on other vendors aswell by installing different helm charts.
```bash
microk8s.enable dns ingress helm3
```
Verify the installation:
- **Ingress**
In the cluster there should be a namespace called **ingress**. Run `microk8s.kubectl get all -n ingress` to see a similar output like:

```
NAME                                          READY   STATUS    RESTARTS   AGE
pod/nginx-ingress-microk8s-controller-gl2pt   1/1     Running   0          3m42s

NAME                                               DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/nginx-ingress-microk8s-controller   1         1         1       1            1           <none>          3m42s
```

- **DNS**
In namespace **kube-system** to deploments should be running. `microk8s.kubectl get deployments -n kube-system`:

```
NAME            READY   UP-TO-DATE   AVAILABLE   AGE
coredns         1/1     1            1           6m23s
tiller-deploy   1/1     1            1           4m56s
```

- **Helm**
In helm3 does not need a tiller pod, so when you can successfully execute: `microk8s.helm3 ls`, you should be fine!

## Pull secret creation
This section will explain how to setup the pull-secret using the **gcr service account**.

### Create pull-secret
Because we are **good cloud engineers** we create a separate namespace and install the AE on that.

```bash
microk8s.kubectl create ns aetryout
microk8s.kubectl config set-context --current --namespace=aetryout
```

This secret is needed by kubernetes otherwise it cannot authenticate againsed the registry. We create the secret based on the current **local docker config**.

```bash
cat <<EOT > automic-dev-ro.json
{
  "type": "service_account",
  "project_id": "esd-automic-saas",
  "private_key_id": "fb9a89c2b7473cf8add18c23934ef075e7c7d457",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDeUchKaLWa7xVV\nr8NDhWaLv7H9GJSVpk3r3DmOvQeDKA9PtMKUC1i4htJwJBbp2CEqUbiRYkYOjnNZ\nX+r2UA7ct9eqWAdD4rF3/tugoq6dzvQdnHmjtEjIXazuUTQC+fZCpTOByW4UBD9N\n1AepHrc6Un1Rq/6JohqLiYi4n2pF4dxZQEnVJYMhxdSnU4E4mtVe21ODTPFb4Xz9\n0V/ycIDCzfzSdb97R8UICxCEJvLDRH+HnECsmX+NlUlxQ0HVUYcxJaN55M6O9JwQ\nATCPlEnG73EQW3Cw0yAeZFyeYl8XRj10zqY5HlZkP3YCVNfIzghUwZ50B7BlKmxJ\nmVcY+Es3AgMBAAECggEAVvWRTP2hD122MCKEU6xh3IbaVX/gWprGvtuQzfLFdflc\n59XyCBtaFC90L7YGGmjWLCnz8jYI5he1Kb/ZdYgCEDZ+zpwJF3Yb6a5P9Qi9GXAC\nT3TNpYlWsLznb/5mREXGm/HncDw8aOryYfxuFKo1jEQIzcHjWa2FCZB94I1Gcdd4\niUMJkCS64VcFMaM0Q17fdRfja7ul83r6GvziwrbZf1NHyg1X+3rLuEsaXbLbYu3O\nEg9JHxCEW7nv3P4iWHEc9bGbefhyDzJ29bpCUQIQwQu7I3ZM5GEqJXsXwpBOUpzV\nTOsP/FDqTe9EPaXKWKYrCUPL3I7cgnsBwf7Ux9rt+QKBgQDv9LoZ6YFQI7pVln6e\nQpUHyQTXwuBaG355cgt/B3OYcPoUfnn6muA/YLBrbnfrRwTP9xvdITN25marpDyE\nd8nAt6eMxIgFEmuiSfQkseVl4cHK5rgPLpSrsMKl0S1k9hy5PThzUKwwBj9dJvv0\nbEbvkVquX7pqqCp7iEHLkuGvHwKBgQDtLyziYM+BblMYMqMVrUtYl+L4r+QJux9B\nGnc7ajr3fMl+07CmYmqYj/ZeT+NWgEGnNQLhlMeQKbpsDwHnINj6Emz+uuA0ZBZ3\n6q3t2UZ5Kv6PiUUaujP2s4PfM1k/05CwNJExCq5Fy9ySvlDnaV2Cy8rfMtGIkyVg\nq12eezsY6QKBgCRwJVKKAvkIc+NLVy7xLXBhNjsNfMQyKKKIjvZbS1J61X8HNb64\nhhUZubCWtd8kibaK79BEmmwT0MN/zTDQf/Kj8O2Paphak72xPUHVQeCWx7boEks7\n55eq3+QOP7Z1KSd4BHp+ZadlS3n50YjsaFk42WxhXQ2VO95GcrdXNq5BAoGBAIoX\nCT7TtnxYyzAvaxvXxSJTa+X2IgI4W73/tqN0+dfVY0rf3N1CN2WTi5DlWiqmiZLc\nHk1P3dBlOxBmvGjgivMtfx/flWFrVFmE3La55XnuOj8/YGhrOI3Nfl2Y+8FZX8f3\nEwFGgqhIRKd6/od8pODd3cONRskJQp3Bp8P5YzLhAoGAcomGH7oW3oKx68ht4w4Q\nQx1P5/GiUH26IGYExYR7JlQ1M/ZS8EeLOA4M14yhPSuN1RB12yLBh+kM8wWD+hIr\nlEqNbpyc2GE+rTJMSDFeZEQZPOfCrn6wXgzawpux4gTVfflkdkkHIPfmeptp4yqd\nLZBBTLCkFL5nxCjOj/zKxXw=\n-----END PRIVATE KEY-----\n",
  "client_email": "automic-dev-ro@esd-automic-saas.iam.gserviceaccount.com",
  "client_id": "112013472759876826194",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/automic-dev-ro%40esd-automic-saas.iam.gserviceaccount.com"
}
EOT
microk8s.kubectl create secret docker-registry pullsecret \
      --docker-server=gcr.io \
      --docker-username=_json_key \
      --docker-password="$(cat ./automic-dev-ro.json)" \
      --docker-email=automic-dev-ro@esd-automic-saas.iam.gserviceaccount.com
```

### [Optional] Ingress routes
This section is how to setup the Ingress routes.

#### Create TLS certificate
For creating the Certificates i use `mkcert` 

> [!info]
> Thanks Peter for suggesting this tool :)

Installing mkcert:
```bash
sudo apt update
sudo apt install libnss3-tools
wget https://github.com/FiloSottile/mkcert/releases/download/v1.1.2/mkcert-v1.1.2-linux-amd64
mv mkcert-v1.1.2-linux-amd64 mkcert
chmod +x mkcert
cp mkcert /usr/local/bin/
```

Now you can start creating the star certificate for a xip.io address, there is only the question to which address we need to resolve, you can find that out by using this bash script.

```bash
echo "$(ip -o route get 8.8.8.8 | awk '{print $7}').xip.io"
mkcert "*.$(ip -o route get 8.8.8.8 | awk '{print $7}').xip.io"
mkcert --install
```

> [!info]
> you may need to restart Chrome for the installed CA to take effect

Now you can create the tls secret for ingress.

```bash
microk8s.kubectl create secret tls certsecret \
--key _wildcard.*.xip.io-key.pem \
--cert _wildcard.*.xip.io.pem
```

### Run Helm install
Before we need to create a values.yaml which defines a couple of things.

> [!info]
> This is subject to change because in the end for a default installation you should need to create a values.yaml, but this is a technical dept that we have.

```bash
cat <<EOT > values.yml
images:
  operator:
    location: gcr.io/esd-automic-saas/automic/aa/tp1/install-operator:2.0.0
  application:
    namespaceUrl: gcr.io/esd-automic-saas/automic/aa/tp1/
  pullSecret: pullsecret

openshift:
  enabled: false

crd:
  install: true

ingress:
  enabled: true
  applicationHostname: $(ip -o route get 8.8.8.8 | awk '{print $7}').xip.io
  secretName: certsecret

spec:
  version: "12.5.0"
  awiReplicas: 1
  cpReplicas: 1
  jcpRestReplicas: 1
  jcpWsReplicas: 1
  jwpReplicas: 1
  wpReplicas: 3
  db: temp

operator:
  serviceAccount: null
EOT
microk8s.helm3 install automic https://downloads.automic.com/jart/prj3/dlc/resources/downloads/GCP-tech-preview_automic.automation_12_5_0.tgz -f values.yml
```

### Customize Installation
Which values you can set in the **values.yaml**, are documented in the helm chart itself, inorder to get that you need to download the helm chart end extract the document.

```bash
wget https://downloads.automic.com/jart/prj3/dlc/resources/downloads/GCP-tech-preview_automic.automation_12_5_0.tgz

tar -xvf GCP-tech-preview*.tgz
cat aa-install-operator/README.md
```

In this README for example describe how to setup the ae on kubernetes with an external database.

### Verify Installation Status
You can verify the status of the installation with the following command. A completed installation will return "Provisioned", an in-progress installation "Provisioning":
```bash
microk8s.kubectl get baa automic-automation -o jsonpath="{$.status.stage}"
```

## Next Steps
### Obtain credentials
In order to login to the containerized AWI, you need the credentials of the client 0 or 100 user. These can be obtained using the following shell commands:

```bash
echo "Client 0 User: $(kubectl get secrets client0-user -o jsonpath={$.data.user} | base64 -d)"
echo "Client 0 Password: $(microk8s kubectl get secrets client0-user -o jsonpath={$.data.password} | base64 -d)"
echo "Client 100 User: $(kubectl get secrets client100-user -o jsonpath={$.data.user} | base64 -d)"
echo "Client 100 Password: $(microk8s kubectl get secrets client100-user -o jsonpath={$.data.password} | base64 -d)"
```

### Identify Nodeport (optional)
If you haven't setup an ingress, you need to find out the nodePort under which AWI is running:

```bash
microk8s.kubectl get svc awi -o jsonpath="http://localhost:{$.spec.ports[0].nodePort}/"
```

You can then access AWI using Output, e.g. http://localhost:32028/

## TL:DR;
if you are bored of reading stuff :D here is a copy-based script which sets up microk8s and installs the AE.

### Prerequisites
make sure the following tools are installed:

```bash
snap --version
mkcert --version
```

Ultimate script:

```bash
sudo snap install microk8s --classic
microk8s.enable dns ingress helm3
microk8s.kubectl create ns aetryout
microk8s.kubectl config set-context --current --namespace=aetryout

cat <<EOT > automic-dev-ro.json
{
  "type": "service_account",
  "project_id": "esd-automic-saas",
  "private_key_id": "fb9a89c2b7473cf8add18c23934ef075e7c7d457",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDeUchKaLWa7xVV\nr8NDhWaLv7H9GJSVpk3r3DmOvQeDKA9PtMKUC1i4htJwJBbp2CEqUbiRYkYOjnNZ\nX+r2UA7ct9eqWAdD4rF3/tugoq6dzvQdnHmjtEjIXazuUTQC+fZCpTOByW4UBD9N\n1AepHrc6Un1Rq/6JohqLiYi4n2pF4dxZQEnVJYMhxdSnU4E4mtVe21ODTPFb4Xz9\n0V/ycIDCzfzSdb97R8UICxCEJvLDRH+HnECsmX+NlUlxQ0HVUYcxJaN55M6O9JwQ\nATCPlEnG73EQW3Cw0yAeZFyeYl8XRj10zqY5HlZkP3YCVNfIzghUwZ50B7BlKmxJ\nmVcY+Es3AgMBAAECggEAVvWRTP2hD122MCKEU6xh3IbaVX/gWprGvtuQzfLFdflc\n59XyCBtaFC90L7YGGmjWLCnz8jYI5he1Kb/ZdYgCEDZ+zpwJF3Yb6a5P9Qi9GXAC\nT3TNpYlWsLznb/5mREXGm/HncDw8aOryYfxuFKo1jEQIzcHjWa2FCZB94I1Gcdd4\niUMJkCS64VcFMaM0Q17fdRfja7ul83r6GvziwrbZf1NHyg1X+3rLuEsaXbLbYu3O\nEg9JHxCEW7nv3P4iWHEc9bGbefhyDzJ29bpCUQIQwQu7I3ZM5GEqJXsXwpBOUpzV\nTOsP/FDqTe9EPaXKWKYrCUPL3I7cgnsBwf7Ux9rt+QKBgQDv9LoZ6YFQI7pVln6e\nQpUHyQTXwuBaG355cgt/B3OYcPoUfnn6muA/YLBrbnfrRwTP9xvdITN25marpDyE\nd8nAt6eMxIgFEmuiSfQkseVl4cHK5rgPLpSrsMKl0S1k9hy5PThzUKwwBj9dJvv0\nbEbvkVquX7pqqCp7iEHLkuGvHwKBgQDtLyziYM+BblMYMqMVrUtYl+L4r+QJux9B\nGnc7ajr3fMl+07CmYmqYj/ZeT+NWgEGnNQLhlMeQKbpsDwHnINj6Emz+uuA0ZBZ3\n6q3t2UZ5Kv6PiUUaujP2s4PfM1k/05CwNJExCq5Fy9ySvlDnaV2Cy8rfMtGIkyVg\nq12eezsY6QKBgCRwJVKKAvkIc+NLVy7xLXBhNjsNfMQyKKKIjvZbS1J61X8HNb64\nhhUZubCWtd8kibaK79BEmmwT0MN/zTDQf/Kj8O2Paphak72xPUHVQeCWx7boEks7\n55eq3+QOP7Z1KSd4BHp+ZadlS3n50YjsaFk42WxhXQ2VO95GcrdXNq5BAoGBAIoX\nCT7TtnxYyzAvaxvXxSJTa+X2IgI4W73/tqN0+dfVY0rf3N1CN2WTi5DlWiqmiZLc\nHk1P3dBlOxBmvGjgivMtfx/flWFrVFmE3La55XnuOj8/YGhrOI3Nfl2Y+8FZX8f3\nEwFGgqhIRKd6/od8pODd3cONRskJQp3Bp8P5YzLhAoGAcomGH7oW3oKx68ht4w4Q\nQx1P5/GiUH26IGYExYR7JlQ1M/ZS8EeLOA4M14yhPSuN1RB12yLBh+kM8wWD+hIr\nlEqNbpyc2GE+rTJMSDFeZEQZPOfCrn6wXgzawpux4gTVfflkdkkHIPfmeptp4yqd\nLZBBTLCkFL5nxCjOj/zKxXw=\n-----END PRIVATE KEY-----\n",
  "client_email": "automic-dev-ro@esd-automic-saas.iam.gserviceaccount.com",
  "client_id": "112013472759876826194",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/automic-dev-ro%40esd-automic-saas.iam.gserviceaccount.com"
}
EOT
microk8s.kubectl create secret docker-registry pullsecret \
      --docker-server=gcr.io \
      --docker-username=_json_key \
      --docker-password="$(cat ./automic-dev-ro.json)" \
      --docker-email=automic-dev-ro@esd-automic-saas.iam.gserviceaccount.com

mkcert "*.$(ip -o route get 8.8.8.8 | awk '{print $7}').xip.io"
mkcert --install

microk8s.kubectl create secret tls certsecret \
--key _wildcard.*.xip.io-key.pem \
--cert _wildcard.*.xip.io.pem

cat <<EOT > values.yml
images:
  operator:
    location: gcr.io/esd-automic-saas/automic/aa/tp1/install-operator:2.0.0
  application:
    namespaceUrl: gcr.io/esd-automic-saas/automic/aa/tp1/
  pullSecret: pullsecret

openshift:
  enabled: false

crd:
  install: true

ingress:
  enabled: true
  applicationHostname: $(ip -o route get 8.8.8.8 | awk '{print $7}').xip.io
  secretName: certsecret

spec:
  version: "12.5.0"
  awiReplicas: 1
  cpReplicas: 1
  jcpRestReplicas: 1
  jcpWsReplicas: 1
  jwpReplicas: 1
  wpReplicas: 3
  db: temp

operator:
  serviceAccount: null
EOT

microk8s.helm3 install automic https://downloads.automic.com/jart/prj3/dlc/resources/downloads/GCP-tech-preview_automic.automation_12_5_0.tgz -f values.yml
```

# Backlinks

[[!Kubernetes]]


