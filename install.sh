cp mirror_service.sh /usr/bin/
chmod +x /usr/bin/mirror_service.sh

cp mirror_script /usr/bin/

FILE=/usr/bin/mirror_config.yaml
if [ -f "$FILE" ]; then
    echo "Configuration file existed!"
else 
    mirror_script config
    cp mirror_config.yaml /usr/bin/
fi

cp mirror.service /lib/systemd/system/

systemctl start mirror
systemctl enable mirror