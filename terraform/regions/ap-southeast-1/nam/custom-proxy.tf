# data "http" "my_ip" {
#   url = "https://checkip.amazonaws.com/"
# }

# locals {
#   my_ip                 = chomp(data.http.my_ip.response_body)
#   account_id            = "832557411742"
#   region                = "ap-southeast-1"
#   ecr_registry          = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com"
#   forward_proxy_image   = "go-somewhere/forward-proxy"
#   tag                   = "1.0.0"
#   image_arn             = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com/${local.forward_proxy_image}:${local.tag}"
# }

# resource "aws_vpc" "main" {
#   cidr_block           = "172.29.0.0/16"
#   enable_dns_support   = true
#   enable_dns_hostnames = true

#   tags = {
#     Name = "nam_vpc"
#   }
# }

# resource "aws_subnet" "public" {
#   vpc_id                  = aws_vpc.main.id
#   cidr_block              = "172.29.1.0/24"
#   availability_zone       = "ap-southeast-1a"
#   map_public_ip_on_launch = true

#   tags = {
#     Name = "nam_vpc_public_subnet"
#   }
# }

# resource "aws_internet_gateway" "igw" {
#   vpc_id = aws_vpc.main.id

#   tags = {
#     Name = "nam_vpc_igw"
#   }
# }

# resource "aws_route_table" "public" {
#   vpc_id = aws_vpc.main.id

#   route {
#     cidr_block = "0.0.0.0/0"
#     gateway_id = aws_internet_gateway.igw.id
#   }
# }

# resource "aws_route_table_association" "public_assoc" {
#   subnet_id      = aws_subnet.public.id
#   route_table_id = aws_route_table.public.id
# }

# resource "aws_security_group" "public_sg" {
#   name        = "nam_vpc_public_sg"
#   description = "Allow from anywhere"
#   vpc_id      = aws_vpc.main.id

#   ingress {
#     description = "SSH"
#     from_port   = 22
#     to_port     = 22
#     protocol    = "tcp"
#     cidr_blocks = ["${local.my_ip}/32"] # ⚠️ Use your IP for better security
#   }

#   ingress {
#     description = "Squid"
#     from_port   = 8888
#     to_port     = 8888
#     protocol    = "tcp"
#     cidr_blocks = ["${local.my_ip}/32"] # ⚠️ Use your IP for better security
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   tags = {
#     Name = "nam_vpc_public_sg"
#   }
# }

# resource "aws_key_pair" "mykey" {
#   key_name   = "mykey"
#   public_key = file("~/.ssh/id_rsa.pub") # path to your SSH public key
# }

# resource "aws_iam_role" "ec2_ecr_role" {
#   name = "ec2-ecr-access-role"

#   assume_role_policy = jsonencode({
#     Version = "2012-10-17",
#     Statement = [
#       {
#         Action = "sts:AssumeRole",
#         Principal = {
#           Service = "ec2.amazonaws.com"
#         },
#         Effect = "Allow",
#         Sid = ""
#       }
#     ]
#   })
# }

# resource "aws_iam_role_policy_attachment" "ecr_readonly_attach" {
#   role       = aws_iam_role.ec2_ecr_role.name
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
# }

# resource "aws_iam_instance_profile" "ec2_instance_profile" {
#   name = "ec2-ecr-instance-profile"
#   role = aws_iam_role.ec2_ecr_role.name
# }

# resource "aws_instance" "public_instance" {
#   ami                           = "ami-088d74defe9802f14" # Amazon Linux 2023 kernel-6.1 AMI 64-bit (x86)
#   instance_type                 = "t3.micro"
#   subnet_id                     = aws_subnet.public.id
#   vpc_security_group_ids        = [aws_security_group.public_sg.id]
#   associate_public_ip_address   = true
#   key_name                      = aws_key_pair.mykey.key_name
#   iam_instance_profile          = aws_iam_instance_profile.ec2_instance_profile.name

#   user_data = <<-EOF
# #!/bin/bash
# set -xe

# # Install dependencies
# sudo yum update -y
# sudo yum install -y docker aws-cli
# sudo systemctl enable docker
# sudo systemctl start docker
# sudo usermod -a -G docker ec2-user

# # Wait for Docker
# sleep 5

# # Login to ECR
# aws ecr get-login-password --region ${local.region} | docker login --username AWS --password-stdin ${local.ecr_registry}

# # Create systemd service
# cat <<'EOT' > /etc/systemd/system/myproxy.service
# [Unit]
# Description=My Proxy Service
# After=docker.service
# Requires=docker.service

# [Service]
# Type=oneshot
# RemainAfterExit=true
# Restart=no
# ExecStartPre=/bin/sleep 5
# ExecStartPre=/bin/bash -c 'aws ecr get-login-password --region ${local.region} | docker login --username AWS --password-stdin ${local.ecr_registry}'
# ExecStartPre=-/usr/bin/docker stop myproxy
# ExecStartPre=-/usr/bin/docker rm myproxy
# ExecStartPre=/usr/bin/docker pull ${local.image_arn}
# ExecStart=/usr/bin/docker run -d --name myproxy -p 8888:8888 ${local.image_arn}
# ExecStop=/usr/bin/docker stop myproxy

# [Install]
# WantedBy=multi-user.target
# EOT

# # Enable and start service
# sudo systemctl daemon-reload
# sudo systemctl enable myproxy.service
# sudo systemctl start myproxy.service
# EOF

#   tags = {
#     Name = "nam_public_ec2"
#   }
# }

# output "public_ip" {
#   value = aws_instance.public_instance.public_ip
# }