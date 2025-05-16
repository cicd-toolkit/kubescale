# kubescale

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

Scale up and down Kubernetes resources after work hours

**Homepage:** <https://github.com/cicd-toolkit/kubescale>

## How to install this chart

To install the chart with the release name `my-release`:

```console
helm install my-release --repo https://cicd-toolkit.github.io/kubescale kubescale
```

To install with custom values file:

```console
helm repo add cicd-toolkit https://cicd-toolkit.github.io/kubescale
helm install my-release cicd-toolkit/kubescale -f values.yaml
```

## Source Code

* <https://github.com/cicd-toolkit/helm-charts>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| containerSecurityContext.readOnlyRootFilesystem | bool | `true` |  |
| containerSecurityContext.runAsGroup | int | `1001` |  |
| containerSecurityContext.runAsNonRoot | bool | `true` |  |
| containerSecurityContext.runAsUser | int | `1001` |  |
| containerSecurityContext.seLinuxOptions | object | `{}` |  |
| containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| extraLabels | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.args | list | `[]` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/cicd-toolkit/cicd-toolkit/kubescale"` |  |
| image.tag | string | `"latest"` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podSecurityContext.fsGroupChangePolicy | string | `"Always"` |  |
| podSecurityContext.supplementalGroups | list | `[]` |  |
| podSecurityContext.sysctls | list | `[]` |  |
| priorityClassName | string | `""` |  |
| rbac.create | bool | `true` |  |
| rbac.extraRules | list | `[]` |  |
| rbac.serviceAccountName | string | `"kubescale"` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | int | `1` |  |
| resources.limits.memory | string | `"1Gi"` |  |
| resources.requests.cpu | string | `"150m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| giuliocalzolari |  | <https://github.com/giuliocalzolari> |
