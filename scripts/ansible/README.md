## Ansible playbooks to setup build environment for bitmask-vpn on macOS and Windows

This contains two very simple ansible playbooks to create windows and macOS build hosts, We used ansible to make it scalable to more then one host at a time and for easy extensibility later.
The `windows` and `macos` subdirectories contain the playbooks and documentation around how to use them. The following sections provide instructions for installing Ansible on the control node and testing these playbooks during development.

### Installing Ansible

We need to install Ansible on the node from where we'll be running the playbooks, we'll need the cli tools `ansible`, `ansible-playbook` to use these playbooks.

> **NOTE:** Control node needs to be a linux or macOS host, we haven't tested from an Windows control node.

You can follow the official documentation for [Installing Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

Or if you already have `python` installed (which is the case for most linux distros and macOS):

```
$ python3 -m pip install --user ansible

# Additionally `ansible-lint` is another useful package to install
```

### Developing

While making changes to these playbooks its very useful to test and verify new changes locally, for this we suggest to use [`quickemu`](https://github.com/quickemu-project) to quickly spin up macOS and Windows test VMs.

#### Installing `quickemu`

At the time of writing this, `quickemu` only supports running on _Linux distros_. There's on-going work to support macOS but not available yet. Please look at their [README](https://github.com/quickemu-project/quickemu#------quickemu) for specific installation instructions for your OS.

#### Testing workflow with `quickemu`

##### Creating a Windows 10 VM

First we need to obtain the Windows 10 installation ISO, run the following commands to download the ISO:

```
$ mkdir win10 # we are creating this directory which will contain all the files needed by this VM
$ cd win10
$ quickget windows 10
```

After running the above commands, the contents of the `win10` directory should be similar to following:

```
$ ls win10
windows-10  windows-10.conf
```
More details about the `conf` file's and the directorie's contents can be found in their [README](https://github.com/quickemu-project/quickemu#------quickemu).

Then start the VM and follow the usual windows setup process:

```
# ensure your inside the win10/ dir
$ quickemu --vm windows-10.conf
```

##### Creating a macOS 10.15 VM

Similar to the Windows 10 VM above we first need to obtain the installation image, the following sequence of commands creates a macOS 10.15 VM:

```
$ mkdir macos-10.15 # dir to contain all the VM related files
$ cd macos-10.15
$ quickget macos catalina # download the catalina (10.15) macos image

# once the image is downloaded the contents of macos-10.15/ dir should look similar to following
$ ls macos-10.15
macos-catalina  macos-catalina.conf
```

To start the VM run the following command:

```
$ cd macos-10.15
$ quickemu --vm macos-catalina.conf
```
This will start the vm and launch _macOS installation wizard_, follow the instructions from [macOS Guest](https://github.com/quickemu-project/quickemu#macos-guest) section of the `quickemu` README to finish the installation.


After creating the VM with the desired OS, follow the `README` for the specific [windows](windows/) or [macOS](macos/) playbook.
