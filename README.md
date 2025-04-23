# kubescaler
# 🧭 Kubescale Operator – Annotation Guide

The **Kubescale Operator** enables automatic scaling of Deployments, StatefulSets, and CronJobs based on custom annotations.

---

## 🔧 Supported Annotations

### 🕒 `kubescale/uptime`
Define when the resource **should be running**.

```yaml
kubescale/uptime: "Mon-Fri 08:00-20:00 Europe/Paris"
```
Format suppported:
| Annotation                  | Interpreted As                                |
|-----------------------------|-----------------------------------------------|
| `08:00-18:00`              | Every day, 08:00–18:00 UTC                    |
| `Mon-Fri 09:00-17:00`      | Mon–Fri, 09:00–17:00 UTC                      |
| `08:00-20:00 Europe/Berlin`| Every day, 08:00–20:00 in Europe/Berlin       |
| `Sat-Sun 10:00-22:00 Asia/Tokyo` | Sat–Sun, 10:00–22:00 in Asia/Tokyo       |


> Timezone must be an IANA TZ (e.g. UTC, Europe/Berlin)

🌙 kubescale/downtime
Define when the resource should be OFF.

```yaml
kubescale/downtime: "Sat-Sun 00:00-23:59 UTC"
```

Takes priority over uptime if both are present.

🚀 kubescale/up
Keep the resource running for a fixed duration
(transformed internally into uptime).

```yaml
kubescale/up: "5h"
```
Valid formats: 4m, 5h, 6d
(minutes, hours, days)

On first reconcile, it's converted into a matching uptime window.

🧠 kubescale/previous-replicas
Used internally to restore original replica count after scale-down.

```yaml
kubescale/previous-replicas: "3"
```

⚠️ Automatically managed by the operator – do not set manually.

🚫 kubescale/exclude

Skip this resource from auto-scaling logic.

```yaml
kubescale/exclude: "true"
```

⏳ kubescale/exclude-until
Skip this resource until a specific timestamp (RFC3339 format).

```yaml
kubescale/exclude-until: "2025-04-23T08:00:00Z"
```

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/kubescaler:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/kubescaler:tag
```


**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kubescaler:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/kubescaler/<tag or branch>/dist/install.yaml
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License
```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

