systemctl stop mirror
systemctl disable mirror
rm /usr/bin/mirror_*  
rm /lib/systemd/system/mirror.service