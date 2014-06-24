linode-inventory
================

**Ansible dynamic inventory plugin for use with Linode.**

The Ansible repository already contains an [Linode dynamic inventory plugin](https://github.com/ansible/ansible/tree/devel/plugins/inventory), which may be useful to you.

This plugin differs from the 'official' one in the following ways:

 * Allows filtering of hosts based on the Linode 'Display Group'. This can help ensure production and staging machines aren't mixed. Configured in the `linode-inventory.ini` with the `display-group` key.
 
 * For each host, sets the `ansible_ssh_host` variable using the public ip. This eliminates the need to reference hosts by their ip, or maintain your `/etc/hosts` file. You can then create another inventory file in the same directory, and reference the hosts by their Linode label.
 
 * Returns host variables in the `_meta` top level element, reducing the number of api calls to Linode and speeding up the provisioning process. This eliminates the need to call the executable with `--host` for each host.
 
 * Uses Linode api batch requests. Only makes 2 requests to the api when called with `--list`.
 
 * No external dependencies.
 
 * Creates less variables per host, but adding more would be trival. Open a pull-request if you need one defined. 

See [Developing Dynamic Inventory Sources](http://docs.ansible.com/developing_inventory.html) for more information.

## Download

**Grab the latest release from [Releases](https://github.com/awilliams/linode-inventory/releases)**

## Usage

 * Download the appropriate package from releases.
 
 * Place the executable inside your ansible directory, alongside other inventory files in a directory or wherever you like.

 * Create a `linode-inventory.ini` file with your Linode API key in the same directory as the executable. See the inlcluded example ini file `linode-inventory.example.ini`.

 * Test the output

 `./linode-inventory --list`

