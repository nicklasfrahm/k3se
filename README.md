# k3se 🧀

[![Go Report Card](https://goreportcard.com/badge/github.com/nicklasfrahm/k3se?style=flat-square)](https://goreportcard.com/report/github.com/nicklasfrahm/k3se)
[![Release](https://img.shields.io/github/release/nicklasfrahm/k3se.svg?style=flat-square)](https://github.com/nicklasfrahm/k3se/releases/latest)
[![Go Reference](https://img.shields.io/badge/Go-reference-informational.svg?style=flat-square)](https://pkg.go.dev/github.com/nicklasfrahm/k3se)

A lightweight Kubernetes engine that deploys `k3s` clusters declaratively based on a cluster configuration file. The name is an abbreviation for _k3s engine_ and a hommage to the German word for cheese, _Käse [ˈkɛːzə]_.

## Quickstart 💡

If you want to test `k3se` you can use [Vagrant][website-vagrant]. All examples in the `examples/` folder can be used with the provided `Vagrantfile` that provisions 3 Ubuntu VMs. To bring up the VMs you can run the following command:

```bash
$ make vagrant-up
```

Once you are done testing, you can destroy the VMs with the following command:

```bash
$ make vagrant-down
```

## Prerequisites 📝

The nodes have to be accessible via SSH, either directly or via a bastion host. Further, the user on the remote nodes needs to have passwordless `sudo` set up. If this is not yet the case, you may manually do so via the following command:

```bash
$ echo "$(whoami) ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/$(whoami)
```

## Limitations 🚨

The following features are currently not supported, but are planned for future releases:

- **High availability**  
  Currently the installation will fail if more than a single node with the role `server` is specified.

- **Downsizing**  
  If nodes are removed from the cluster configuration, they are not decommissioned. We plan to enable this feature in the future using the `git` history of the cluster configuration.

## License 📄

This project is and will always be licensed under the terms of the [MIT license][file-license].

[file-license]: https://www.apache.org/licenses/LICENSE-2.0
[website-vagrant]: https://vagrantup.com
