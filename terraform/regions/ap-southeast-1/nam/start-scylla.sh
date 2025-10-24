#!/bin/bash
set -e

# Variables injected by Terraform templatefile()
IP="${IP}"
SEED_IP="${SEED_IP}"

# Install Docker
sudo yum update -y
sudo yum install -y docker

sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ec2-user

sleep 5

echo "Starting setup for Scylla node"
echo "Local IP: ${IP}"
echo "Seed IP: ${SEED_IP}"

# Create systemd service for the container
cat <<'SERVICE' > /etc/systemd/system/scylla.service
[Unit]
Description=Scylla Container
After=docker.service
Requires=docker.service

[Service]
Restart=always
ExecStart=/usr/bin/docker run --rm --name scylla --network host scylladb/scylla:latest --listen-address=${IP} --broadcast-address=${IP} --rpc-address=${IP} --broadcast-rpc-address=${IP} --seeds=${SEED_IP} --smp=1 --memory=500M --overprovisioned 1
ExecStop=/usr/bin/docker stop scylla

[Install]
WantedBy=multi-user.target
SERVICE

# Replace placeholders with actual values (since heredoc with quotes disables variable expansion)
sed -i "s|\${IP}|${IP}|g" /etc/systemd/system/scylla.service
sed -i "s|\${SEED_IP}|${SEED_IP}|g" /etc/systemd/system/scylla.service

# Enable and start the service
systemctl daemon-reload
systemctl enable scylla
systemctl start scylla

echo "Scylla service started successfully on ${IP}"
