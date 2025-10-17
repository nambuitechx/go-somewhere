# data "http" "my_ip" {
#   url = "https://checkip.amazonaws.com/"
# }

# locals {
#   my_ip = chomp(data.http.my_ip.response_body)
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
#     description = "Postgresql"
#     from_port   = 5432
#     to_port     = 5432
#     protocol    = "tcp"
#     cidr_blocks = ["${local.my_ip}/32"] # ⚠️ Use your IP for better security
#   }

#   ingress {
#     description = "Squid"
#     from_port   = 3128
#     to_port     = 3128
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

# # resource "aws_iam_role" "ec2_ecr_role" {
# #   name = "ec2-ecr-access-role"

# #   assume_role_policy = jsonencode({
# #     Version = "2012-10-17",
# #     Statement = [
# #       {
# #         Action = "sts:AssumeRole",
# #         Principal = {
# #           Service = "ec2.amazonaws.com"
# #         },
# #         Effect = "Allow",
# #         Sid = ""
# #       }
# #     ]
# #   })
# # }

# # resource "aws_iam_role_policy_attachment" "ecr_readonly_attach" {
# #   role       = aws_iam_role.ec2_ecr_role.name
# #   policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
# # }

# # resource "aws_iam_instance_profile" "ec2_instance_profile" {
# #   name = "ec2-ecr-instance-profile"
# #   role = aws_iam_role.ec2_ecr_role.name
# # }

# resource "aws_instance" "public_instance" {
#   ami                           = "ami-088d74defe9802f14" # Amazon Linux 2023 kernel-6.1 AMI 64-bit (x86)
#   instance_type                 = "t3.micro"
#   subnet_id                     = aws_subnet.public.id
#   vpc_security_group_ids        = [aws_security_group.public_sg.id]
#   associate_public_ip_address   = true
#   key_name                      = aws_key_pair.mykey.key_name
# #   iam_instance_profile          = aws_iam_instance_profile.ec2_instance_profile.name

#   user_data = <<-EOF
#     #!/bin/bash
#     sudo yum update -y

#     sudo yum install -y docker
#     sudo systemctl enable docker
#     sudo systemctl restart docker
#     sudo usermod -a -G docker ec2-user

#     sudo yum install -y squid
#     cat <<EOCONF > /etc/squid/squid.conf
#     http_port 3128
#     acl allowed src ${local.my_ip}/32
#     http_access allow allowed
#     http_access deny all
#     cache deny all
#     access_log /var/log/squid/access.log
#     cache_log /var/log/squid/cache.log
#     EOCONF

#     sudo systemctl enable squid
#     sudo systemctl restart squid
#   EOF

#   tags = {
#     Name = "nam_public_ec2"
#   }
# }

# output "public_ip" {
#   value = aws_instance.public_instance.public_ip
# }
