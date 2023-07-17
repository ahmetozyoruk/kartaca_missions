# First Section

## Table of Contents

 * [Creating single node with kubespray](#creating-node)
 * [Creating policy with kyverno](#creating-policy)

## Creating single node with kubespray

**First, create a vm**

We need a vm to create a cluster in kubespray. First of all, we have to provide this vm in our own local or from any vm provider. Then we need to install a distribution that kubespray supports in vm.

Kubespray supports distributions:
 * Flatcar Container Linux by Kinvolk
 * Debian Bullseye, Buster, Jessie, Stretch
 * Ubuntu 16.04, 18.04, 20.04, 22.04
 * CentOS/RHEL 7, 8, 9
 * Fedora 35, 36
 * Fedora CoreOS
 * openSUSE Leap 15.x/Tumbleweed
 * Oracle Linux 7, 8, 9
 * Getting Linux 8, 9
 * RockyLinux 8, 9
 * Kylin Linux Advanced Server V10
 * Amazon Linux 2

**Requirements**

 * Minimum required version of Kubernetes is v1.25
 * Ansible v2.11+, Jinja 2.11+ and python-netaddr is installed on the machine that will run Ansible commands
 * The target servers must have access to the Internet in order to pull docker images. Otherwise, additional configuration is required (See Offline Environment)
 * The target servers are configured to allow IPv4 forwarding.
 * If using IPv6 for pods and services, the target servers are configured to allow IPv6 forwarding.
 * The firewalls are not managed, you'll need to implement your own rules the way you used to. in order to avoid any issue during deployment you should disable your firewall.
 * If kubespray is run from non-root user account, correct privilege escalation method should be configured in the target servers. Then the ansible_become flag or command parameters --become or -b should be specified.

**Building your own inventory**

Ansible inventory can be stored in 3 formats: YAML, JSON, or INI-like. There is
an example inventory located
[here](https://github.com/kubernetes-sigs/kubespray/blob/master/inventory/sample/inventory.ini).

You can use an
[inventory generator](https://github.com/kubernetes-sigs/kubespray/blob/master/contrib/inventory_builder/inventory.py)
to create or modify an Ansible inventory. Currently, it is limited in
functionality and is only used for configuring a basic Kubespray cluster inventory, but it does
support creating inventory file for large clusters as well. It now supports
separated ETCD and Kubernetes control plane roles from node role if the size exceeds a
certain threshold. Run `python3 contrib/inventory_builder/inventory.py help` for more information.

Example inventory generator usage:

```ShellSession
cp -r inventory/sample inventory/mycluster
declare -a IPS=(192.168.122.170)
CONFIG_FILE=inventory/mycluster/hosts.yml python3 contrib/inventory_builder/inventory.py ${IPS[@]}
```

Then use `inventory/mycluster/hosts.yml` as inventory file.

**Starting deployment**

```ShellSession
# Clean up old Kubernete cluster with Ansible Playbook - run the playbook as root
# The option `--become` is required, as for example cleaning up SSL keys in /etc/,
# uninstalling old packages and interacting with various systemd daemons.
# Without --become the playbook will fail to run!
# And be mind it will remove the current kubernetes cluster (if it's running)!
ansible-playbook -i inventory/mycluster/hosts.yaml  --become --become-user=root reset.yml

# Deploy Kubespray with Ansible Playbook - run the playbook as root
# The option `--become` is required, as for example writing SSL keys in /etc/,
# installing packages and interacting with various systemd daemons.
# Without --become the playbook will fail to run!
ansible-playbook -i inventory/mycluster/hosts.yaml  --become --become-user=root cluster.yml 
```
**NOTE:** After the new system is installed, the rejected connection denial appears when you try to log on to the Debian Linux server as a root user. Here is the example information:

```ShellSession
$ ssh root@192.168.122.170
ssh: connect to host 192.168.122.170 port 22: Connection refused
```

To enable SSH root login, perform the following steps:

1. Run the following command to configure the SSH server:
```ShellSession
echo 'PermitRootLogin=yes'  | sudo tee -a /etc/ssh/sshd_config
```
3. Restart the SSH server:
```ShellSession
sudo systemctl restart sshd
```
**Result:**
You will be able to use SSH login using the root account. The following output indicates the login is successful:

**Example Output:**

```ShellSession
[ahmet@ahmethost ~]$ ssh root@192.168.122.170
root@192.168.122.170's password: 
Welcome to Ubuntu 22.04.2 LTS (GNU/Linux 5.19.0-46-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

Expanded Security Maintenance for Applications is not enabled.

14 updates can be applied immediately.
To see these additional updates run: apt list --upgradable

1 additional security update can be applied with ESM Apps.
Learn more about enabling ESM Apps service at https://ubuntu.com/esm


The programs included with the Ubuntu system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Ubuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by
applicable law.

root@kubernetmachine-Standard-PC-Q35-ICH9-2009:~#
```

Then try this to run the playbook as root

```ShellSession
ansible-playbook -i inventory/mycluster/hosts.yaml  --become --become-user=root cluster.yml --user=root --ask-pass
```

## Creating policy with kyverno

