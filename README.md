# hana-firewall
This is a firewall utility that takes HANA instance numbers and HANA network service definitions as input, and generates firewalld service definitions in XML format.

The utility has been redesigned in SLES 15. In contrast to the version distributed with SLES 12, it no longer controls the system firewall and instead merely generates firewalld service definition files. You must associate the service definitions with appropriate network interfaces using firewalld itself.

# Version History
Version 1.x were originally written between 2015 and 2016 for SLES 12. The latest version 2.x are completely rewritten in order to work with `firewalld`, these versions are distributed with SLES 15.

# Author
Howard Guo <hguo@suse.com>
