#!/bin/sh

[ -z "$1" ] && echo "$0 <version_number>" && exit 1

tar zcvf HANA-Firewall-$1.tar.gz hana-firewall hana-firewall.8 hana-firewall.d hana-firewall.service sysconfig.hana-firewall
