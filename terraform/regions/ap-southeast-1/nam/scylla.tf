# locals {
#   prefix = "nam-vpc"
# }

# resource "aws_vpc" "main" {
#   cidr_block            = "172.29.0.0/16"
#   enable_dns_support    = true
#   enable_dns_hostnames  = true

#   tags = {
#     Name = "${local.prefix}"
#   }
# }

# resource "aws_subnet" "public" {
#   vpc_id                  = aws_vpc.main.id
#   cidr_block              = "172.29.1.0/24"
#   availability_zone       = "ap-southeast-1a"

#   tags = {
#     Name = "${local.prefix}-public-subnet"
#   }
# }

# resource "aws_internet_gateway" "igw" {
#   vpc_id = aws_vpc.main.id

#   tags = {
#     Name = "${local.prefix}-my-igw"
#   }
# }

# resource "aws_route_table" "public" {
#   vpc_id = aws_vpc.main.id

#   route {
#     cidr_block = "0.0.0.0/0"
#     gateway_id = aws_internet_gateway.igw.id
#   }

#   tags = {
#     Name = "${local.prefix}-public-rt"
#   }
# }

# resource "aws_route_table_association" "public_assoc" {
#   subnet_id      = aws_subnet.public.id
#   route_table_id = aws_route_table.public.id
# }

# resource "aws_security_group" "public_sg" {
#   vpc_id = aws_vpc.main.id

#   ingress {
#     description = "Allow port 22"
#     from_port   = 22
#     to_port     = 22
#     protocol    = "tcp"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   ingress {
#     description = "Allow port 7000"
#     from_port   = 7000
#     to_port     = 7000
#     protocol    = "tcp"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   ingress {
#     description = "Allow port 7001"
#     from_port   = 7001
#     to_port     = 7001
#     protocol    = "tcp"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   ingress {
#     description = "Allow port 9042"
#     from_port   = 9042
#     to_port     = 9042
#     protocol    = "tcp"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
# }

# resource "aws_key_pair" "mykey" {
#   key_name   = "mykey"
#   public_key = file("~/.ssh/id_rsa.pub")
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

# resource "aws_instance" "scylla_seed" {
#   ami                         = "ami-088d74defe9802f14" # Amazon Linux 2023 kernel-6.1 AMI 64-bit (x86)
#   instance_type               = "t3.micro"
#   subnet_id                   = aws_subnet.public.id
#   vpc_security_group_ids      = [aws_security_group.public_sg.id]
#   associate_public_ip_address = true
#   iam_instance_profile        = aws_iam_instance_profile.ec2_instance_profile.name
#   key_name                    = aws_key_pair.mykey.key_name

#   user_data = templatefile("./start-scylla.sh", {
#     SEED_IP = "172.29.1.11"
#     IP      = "172.29.1.11"
#   })

#   private_ip = "172.29.1.11"

#   tags = { Name = "${local.prefix}-scylla-seed" }
# }

# resource "aws_instance" "scylla_node2" {
#   ami                     = "ami-088d74defe9802f14" # Amazon Linux 2023 kernel-6.1 AMI 64-bit (x86)
#   instance_type           = "t3.micro"
#   subnet_id               = aws_subnet.public.id
#   vpc_security_group_ids  = [aws_security_group.public_sg.id]
#   iam_instance_profile    = aws_iam_instance_profile.ec2_instance_profile.name
#   key_name                = aws_key_pair.mykey.key_name

#   user_data = templatefile("./start-scylla.sh", {
#     SEED_IP = "172.29.1.11"
#     IP      = "172.29.1.12"
#   })

#   private_ip = "172.29.1.12"

#   tags = { Name = "${local.prefix}-scylla-node2" }
#   depends_on = [aws_instance.scylla_seed]
# }

# resource "aws_instance" "scylla_node3" {
#   ami                     = "ami-088d74defe9802f14" # Amazon Linux 2023 kernel-6.1 AMI 64-bit (x86)
#   instance_type           = "t3.micro"
#   subnet_id               = aws_subnet.public.id
#   vpc_security_group_ids  = [aws_security_group.public_sg.id]
#   iam_instance_profile    = aws_iam_instance_profile.ec2_instance_profile.name
#   key_name                = aws_key_pair.mykey.key_name

#   user_data = templatefile("./start-scylla.sh", {
#     SEED_IP = "172.29.1.11"
#     IP      = "172.29.1.13"
#   })

#   private_ip = "172.29.1.13"

#   tags = { Name = "${local.prefix}-scylla-node3" }
#   depends_on = [aws_instance.scylla_seed]
# }
