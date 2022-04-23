# k3se 🧀

A lightweight Kubernetes engine that deploys `k3s` clusters declaratively based on a cluster configuration file. The name is a hommage to the German word for cheese, _Käse [ˈkɛːzə]_.

## Requirements

The nodes have to be accessible via SSH, either directly or via a bastion host. Further, the user on the remote nodes needs to have passwordless `sudo` set up. If this is not yet the case, you may manually do so via the following command:

```bash
$ echo "$(whoami) ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/$(whoami)
```

## License 📄

This project is and will always be licensed under the terms of the [MIT license][file-license].

[file-license]: https://www.apache.org/licenses/LICENSE-2.0
