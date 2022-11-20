## Ansible playbook to setup Dev env for Bitmask VPN development

### Prerequisites on the target macOS host

Although the playbook should work for any macOS version >= 10.15 (Catalina), it has been tested on macOS versions 10.15 and 12.6.

- SSH is enabled, go to **System Settings > Sharing** and tick **Remote Login**
- Ansible user for the host should be a sudoer
- Public key authentication for SSH is setup

### Playbook organisation

```
├── inventory.yaml       # Example inventory file
├── requirements.yaml    # Collection and roles the playbook uses
├── site.yaml            # Playbook
```

It currently installs Homebrew and Qt Installer FW on the target host apart from the various development tools installed from Homebrew.
All the dependencies from homebrew are defined using the `homebrew_installed_packages` variable

For installing Homebrew and packages from it we make use of the `geerlingguy.mac.homebrew` role from the [mac ansible collection.](https://galaxy.ansible.com/geerlingguy/mac)
The included `inventory` file is just an example file for easy testing during development

> **NOTE:** The playbook doesn't add Qt, QtIFW and Golang `bin` directories to `PATH` on some macOS versions. 
This needs to be set by the user before running the `make` targets. 
To get the needed filepath for Qt and Golang `bin` directories, use `brew info <go@1.17 | qt5>`. 
QtIFW gets installed to a directory named _Qt_ in the user's home folder, filepath to add to `PATH` is `~/Qt/QtIFW-4.4.2/bin`.

### How to run the playbook

Install the required collections and modules from ansible-galaxy:

```
$ ansible-galaxy collection install -r requirements.yaml
$ ansible-galaxy role install -r requirements.yaml
```

Make sure you have a valid inventory file, update the provided `inventory.yaml` file with your VM or remote host's IP address and run:

```
$ ansible-playbook -i inventory.yaml site.yaml --private-key=<path_to_ssh_key> --ask-become-pass
```
Or to target the localhost run:

```
$ ansible-playbook --connection=local --ask-become-pass site.yaml
```
