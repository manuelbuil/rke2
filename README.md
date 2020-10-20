# RKE2
![RKE2](docs/assets/logo-horizontal-rke.svg)

RKE2, also known as RKE Government, is Rancher's next-generation Kubernetes distribution.

It is a fully [conformant Kubernetes distribution](https://landscape.cncf.io/selected=rke-government) that focuses on security and compliance within the U.S. Federal Government sector.

To meet these goals, RKE2 does the following:

- Provides [defaults and configuration options](security/hardening_guide.md) that allow clusters to pass the [CIS Kubernetes Benchmark](security/cis_self_assessment.md) with minimal operator intervention
- Enables [FIPS 140-2 compliance](security/fips_support.md)
- Supports SELinux policy and [Multi-Category Security (MCS)](https://selinuxproject.org/page/NB_MLS) label enforcement
- Regularly scans components for CVEs using [trivy](https://github.com/aquasecurity/trivy) in our build pipeline

For more information and detailed installation and operation instructions, [please visit our docs](https://docs.rke2.io/).

## Quick Start
Here's the ***extremely*** quick start:
```sh
curl -sfL https://get.rke2.io | sh -
systemctl enable rke2-server.service
systemctl start rke2-server.service
# Wait a bit
export KUBECONFIG=/etc/rancher/rke2/rke2.yaml PATH=$PATH:/var/lib/rancher/rke2/bin
kubectl get nodes
```
For a bit more, [check out our full quick start guide](https://docs.rke2.io/install/quickstart/).

## Installation

A full breakdown of installation methods and information can be found [here](docs/install/methods.md).

## Configuration File

The primary way to configure RKE2 is through its [config file](https://docs.rke2.io/install/install_options/install_options/#configuration-file). Command line arguments and environment variables are also available, but RKE2 is installed as a systemd service and thus these are not as easy to leverage.

By default, RKE2 will launch with the values present in the YAML file located at `/etc/rancher/rke2/config.yaml`.

An example of a basic `server` config file is below:

```yaml
# /etc/rancher/rke2/config.yaml
write-kubeconfig-mode: "0644"
tls-san:
  - "foo.local"
node-label:
  - "foo=bar"
  - "something=amazing"
```

In general, cli arguments map to their respective yaml key, with repeatable cli args being represented as yaml lists. So, an identical configuration using solely cli arguments is shown below to demonstrate this:

```bash
rke2 server \
  --write-kubeconfig-mode "0644"    \
  --tls-san "foo.local"             \
  --node-label "foo=bar"            \
  --node-label "something=amazing"
```

It is also possible to use both a configuration file and cli arguments.  In these situations, values will be loaded from both sources, but cli arguments will take precedence.  For repeatable arguments such as `--node-label`, the cli arguments will overwrite all values in the list.

Finally, the location of the config file can be changed either through the cli argument `--config FILE, -c FILE`, or the environment variable `$RKE2_CONFIG_FILE`.

## Cilium

To build the runtime image using cilium instead of canal, first of all you require to build the HelmChart:

`CHART_FILE=rke2-cilium.yaml CHART_TMP=$HOME/rke2/cilium-1.8.4.tgz charts/build-chart.sh`

That will build rke2-cilium.yaml. Note that cilium-1.8.4.tgz is part of this repo and was created by rke2-charts scripts.

Once you have rke2-cilium.yaml in the $HOME/rke2 directory, run:

`make build-image-runtime-cilium`

That will end up generating a Docker image and saving it as a tar file in build/images/rke2-runtime.tar. Remember the tag of the image, in my case it is `v1.18.9-dev-3ba3b9d5-dirty`. I'll refer to it as $TAG

Copy that tarball to a directory $YOUR_DIRECTORY/agent/images/

Then, change the systemctl file:

```bash
sudo vi /usr/local/lib/systemd/system/rke2-server.service
```

Add:
 Environment="RKE2_RUNTIME_IMAGE=rancher/rke2-runtime-cilium:$TAG"
 ExecStart=/usr/local/bin/rke2 server --data-dir=$YOUR_DIRECTORY

Reload the systemd unit and start it

```bash
sudo systemctl daemon-reload
sudo systemctl start rke2-server.service
```


## FAQ

- [How is the different from RKE1 or K3s?](https://docs.rke2.io/#how-is-this-different-from-rke-or-k3s)
- [Why two names?](https://docs.rke2.io/#why-two-names)
