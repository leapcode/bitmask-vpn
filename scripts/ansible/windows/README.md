## Ansible playbook to setup Dev env for Bitmask VPN development

### Prerequisites

On the target windows host, we need SSH access to be enabled, default shell for SSH should be PowerShell and the user account used for Ansible should be an administrator user.

- To enable OpenSSH on windows follow the [Install OpenSSH for Windows](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse?tabs=powershell#install-openssh-for-windows)guide.

- To set `PowerShell` as the default shell for OpenSSH, follow the [OpenSSH Server Configuration](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_server_configuration) guide.

- Then to enable key based access follow the [OpenSSH Key Management Guide](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_keymanagement)

The playbook has been tested to work with Windows 10.

### Playbook organisation

```
├── inventory.yaml       # Example inventory file
├── requirements.yaml    # Collection and roles the playbook uses
├── site.yaml            # Playbook
```

It currently installs Chocolatey, Qt and Qt Installer FW on the target host apart from the various development tools installed from Chocolatey

For installing Chocolatey and packages from it we make use of the [`chocolatey.chocolatey.win_chocolatey`](https://docs.ansible.com/ansible/latest/collections/chocolatey/chocolatey/win_chocolatey_module.html) module.
The included `inventory` file is just an example file for easy testing during development

### How to run the playbook

Install the required collections and modules from ansible-galaxy:

```
$ ansible-galaxy collection install -r requirements.yaml
```

Make sure you have a valid inventory file, update the provided `inventory.yaml` file with your VM or remote host's IP address and run:

```
$ ansible-playbook -i inventory.yaml site.yaml --private-key=<path_to_ssh_key>
```
