This directory `hana-firewall.d` contains HANA network service definition files.

Be aware that, the files here define how a HANA network service will use instance number as part of their TCP and/or
UDP port numbers. Given these definitions, `hana-firewall` program will generate XML service definition files for
`firewalld`.

See the manpage of `hana-firewall` for usage instructions.