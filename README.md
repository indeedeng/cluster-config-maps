ClusterConfigMaps
===

![OSS Lifecycle](https://img.shields.io/osslifecycle/indeedeng/default-template.svg)

ClusterConfigMaps is kubernetes custom resource and kubernetes csi driver plugin which
enables creating cluster-scoped configmaps.

* [Prerequisites](#prerequisites)
* [Installing](#installing)
* [Usage](#usage)
* [Limitations](#limitations)

Prerequisites
===
ClusterConfigMaps require kubernetes 1.17+ and can either be installed via helm or bare manifests.

Installing
===
See the helm chart available in the /deploy directory.

Usage
===
ClusterConfigMaps are a kubernetes custom resource similar to the native kubernetes ConfigMap resource.
A ClusterConfigMap supports an arbitrary number of key-values with string data, and can be mounted into
containers via a volume.

Example:
```yaml
kind: ClusterConfigMap
apiVersion: indeed.com/v1alpha1
metadata:
  name: example-ccm
data:
  hello_world.txt: |
    hello world!
  test-file.properties: |
    foo=bar
    foo.bar=baz
```

Mounting the above example into a pod spec:
```yaml
kind: Pod
apiVersion: v1
metadata:
  name: ccm-example
spec:
  containers:
    - name: ubuntu
      image: ubuntu:latest
      command:
        - sleep
        - infinity
      volumeMounts:
        - name: csi-example-volume
          mountPath: /example
  volumes:
    - name: csi-example-volume
      csi:
        driver: clusterconfigmaps.indeed.com
        volumeAttributes:
          name: example-ccm
          mode: "0644" # optional, defaults to 0644
```

Limitations
===
ClusterConfigMaps have a few limitations compared to the native kubernetes ConfigMap resource.

* Kubernetes does not provide a way for csi drivers to supply environment variables to pods, unlike ConfigMaps or Secrets. As such, ClusterConfigMaps can not be used as an environment variable source. 
* Additionally, ClusterConfigMaps can not support subpaths or selecting individual items.
  * This is primarily due to the lack of an ability to pass complex variables into `volumeAttributes`.
* ClusterConfigMaps do not currently support reloading the contents after modifications. This may be added in future releases.

Contributions
===
We welcome contributions! Feel free to open an issue or submit a PR!

Maintainers
===
ClusterConfigMaps is maintained by Indeed Engineering.

While we are always busy helping people get jobs, we will try to respond to GitHub issues, pull requests, and questions within a couple of business days.

## Code of Conduct
This project is governed by the [Contributer Covenant v1.4.1](CODE_OF_CONDUCT.md)

For more information please contact opensource@indeed.com.

## License
This project uses the [Apache 2.0](LICENSE) license.