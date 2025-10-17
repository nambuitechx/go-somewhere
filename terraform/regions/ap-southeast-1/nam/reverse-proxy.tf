###########################################
####### ---------- Setup ---------- #######
###########################################
data "http" "my_ip" {
  url = "https://checkip.amazonaws.com"
}

locals {
  my_ip                   = chomp(data.http.my_ip.response_body)
  prefix                  = "nam-vpc"
  account_id              = "832557411742"
  region                  = "ap-southeast-1"
  ecr_registry            = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com"
  reverse_proxy_image     = "go-somewhere/reverse-proxy"
  backend_image           = "go-somewhere/backend"
  tag                     = "1.0.0"
  reverse_proxy_image_arn = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com/${local.reverse_proxy_image}:${local.tag}"
  backend_image_arn       = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com/${local.backend_image}:${local.tag}"
  ec2_ami                 = "ami-088d74defe9802f14"
}

##############################################
####### ---------- Networks ---------- #######
##############################################
resource "aws_vpc" "main" {
  cidr_block              = "172.29.0.0/16"
  enable_dns_support      = true
  enable_dns_hostnames    = true

  tags = {
    Name = "${local.prefix}"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "172.29.1.0/24"
  availability_zone       = "ap-southeast-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "${local.prefix}-public-subnet"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${local.prefix}-main-igw"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = "${local.prefix}-public-rt"
  }
}

resource "aws_route_table_association" "public_assoc" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "172.29.2.0/24"
  availability_zone = "ap-southeast-1a"

  tags = {
    Name = "${local.prefix}-private-subnet"
  }
}

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name = "${local.prefix}-nat-eip"
  }
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = "${local.prefix}-main-nat"
  }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat.id
  }

  tags = {
    Name = "${local.prefix}-private-rt"
  }
}

resource "aws_route_table_association" "private_assoc" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}

#######################################################
####### ---------- EC2 Configuration ---------- #######
#######################################################
resource "aws_iam_role" "ec2_role" {
  name = "ec2-ecr-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecr_access" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_instance_profile" "ec2_profile" {
  name = "ec2-instance-profile"
  role = aws_iam_role.ec2_role.name
}

resource "aws_key_pair" "mykey" {
  key_name   = "mykey"
  public_key = file("~/.ssh/id_rsa.pub") # path to your SSH public key
}

#################################################
####### ---------- Public EC2 ---------- #######
#################################################
resource "aws_security_group" "ec2_public_sg" {
  name        = "public-ec2-sg"
  description = "Security group for public EC2 instance"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "Allow SSH from IP"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${local.my_ip}/32"] # ⚠️ Use your IP for better security
  }

  ingress {
    description = "Allow TCP from everywhere"
    from_port   = 8888
    to_port     = 8888
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # ⚠️ Use your IP for better security
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${local.prefix}-public-ec2-sg"
  }
}

resource "aws_instance" "reverse_proxy" {
  ami                         = "${local.ec2_ami}"
  instance_type               = "t3.micro"
  subnet_id                   = aws_subnet.public.id
  vpc_security_group_ids      = [aws_security_group.ec2_public_sg.id]
  iam_instance_profile        = aws_iam_instance_profile.ec2_profile.name
  associate_public_ip_address = true
  key_name                    = aws_key_pair.mykey.key_name

  user_data = <<-EOF
    #!/bin/bash
    set -xe

    # Install dependencies
    sudo yum update -y
    sudo yum install -y docker aws-cli
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker ec2-user
 
    # Wait for Docker
    sleep 5
  EOF

  tags = {
    Name = "${local.prefix}-public-ec2"
  }
}

#################################################
####### ---------- Private EC2 ---------- #######
#################################################
# resource "aws_security_group" "ec2_private_sg" {
#   name        = "private-ec2-sg"
#   description = "Security group for private EC2 instance"
#   vpc_id      = aws_vpc.main.id

#   ingress {
#     description = "Allow inbound from ALB or Bastion"
#     from_port   = 8000
#     to_port     = 8000
#     protocol    = "tcp"
#     security_groups = [aws_security_group.ec2_public_sg.id]
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   tags = {
#     Name = "${local.prefix}-private-ec2-sg"
#   }
# }

# resource "aws_instance" "app_server" {
#   ami                         = "${local.ec2_ami}"
#   instance_type               = "t3.micro"
#   subnet_id                   = aws_subnet.private.id
#   vpc_security_group_ids      = [aws_security_group.ec2_private_sg.id]
#   iam_instance_profile        = aws_iam_instance_profile.ec2_profile.name
#   associate_public_ip_address = false
#   key_name                    = aws_key_pair.mykey.key_name

#   user_data = <<-EOF
#     #!/bin/bash
#     set -xe

#     # Install Docker
#     sudo yum update -y
#     sudo yum install docker -y aws-cli
#     sudo systemctl enable docker
#     sudo systemctl start docker

#     # Wait for Docker
#     sleep 5

#     # Login to ECR
#     aws ecr get-login-password --region ${local.region} | docker login --username AWS --password-stdin ${local.ecr_registry}

#     # Pull image
#     IMAGE_URI="${local.backend_image_arn}"
#     docker pull $IMAGE_URI

#     # Create systemd service for the container
#     cat <<SERVICE > /etc/systemd/system/myapp.service
#     [Unit]
#     Description=MyApp Container
#     After=docker.service
#     Requires=docker.service

#     [Service]
#     Restart=always
#     ExecStart=/usr/bin/docker run --rm --name myapp -p 8000:8000 $IMAGE_URI
#     ExecStop=/usr/bin/docker stop myapp

#     [Install]
#     WantedBy=multi-user.target
#     SERVICE

#     # Enable and start the service
#     systemctl daemon-reload
#     systemctl enable myapp
#     systemctl start myapp
#   EOF

#   tags = {
#     Name = "${local.prefix}-private-ec2"
#   }
# }
