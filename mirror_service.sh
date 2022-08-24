DATE=`date '+%Y-%m-%d %H:%M:%S'`
echo "Mirror service started at ${DATE}" | systemd-cat -p info

mirror_script

while :
do
echo "Looping...";
sleep 30;
done