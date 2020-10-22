#!/bin/bash

# Setup swapfile
sudo fallocate -l 1G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile swap swap defaults 0 0' | sudo tee -a /etc/fstab

# Setup hitpoints service
sudo mv /home/ubuntu/hitpoints.service /etc/systemd/system/hitpoints.service
sudo chmod 644 /etc/systemd/system/hitpoints.service
sudo mkdir -p /var/www
sudo chmod +x /home/ubuntu/hitpoints
sudo mv /home/ubuntu/hitpoints /var/www/hitpoints
sudo systemctl start hitpoints
sudo systemctl enable hitpoints
