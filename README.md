linode-inventory
================

Ansible Inventory plugin for use with Linode

The Ansible repository contains an [Linode inventory plugin](https://github.com/ansible/ansible/tree/devel/plugins/inventory).

This plugin allows filtering based on the Display Group and has no external dependencies. 
Also, it includes all information in the `--list` form (with only 2 api requests to Linode). 
The `--host` form returns an empty hash.

See [Developing Dynamic Inventory Sources](http://docs.ansible.com/developing_inventory.html) for more information.

## Usage

Download the appropriate package from releases.

Create a `linode-inventory.ini` file with your Linode API key. See the example ini file.

**Command Line Usage**

Print out your available linodes

    ./linode-inventory <displaygroup>

**Ansible Usage**

Ansible will execute with the `--list` or `--host` flag. The `--host` flag will always return an empty hash.

    ./linode-inventory --list