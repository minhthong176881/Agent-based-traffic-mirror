tc qdisc del dev ens33 ingress
tc qdisc del dev lo ingress
tc qdisc del dev ens33 root
ip link del vxlan108
