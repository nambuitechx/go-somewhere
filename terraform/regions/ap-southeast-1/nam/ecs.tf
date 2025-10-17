# #
# # VPC and Subnet
# #
# resource "aws_vpc" "main" {
#   cidr_block           = "172.29.0.0/16"
#   enable_dns_support   = true
#   enable_dns_hostnames = true

#   tags = {
#     Name = "nam-vpc"
#   }
# }

# resource "aws_subnet" "public_a" {
#   vpc_id                  = aws_vpc.main.id
#   cidr_block              = "172.29.1.0/24"
#   availability_zone       = "ap-southeast-1a"
#   map_public_ip_on_launch = true

#   tags = {
#     Name = "nam-vpc-public-subnet-a"
#   }
# }

# resource "aws_subnet" "public_b" {
#   vpc_id                  = aws_vpc.main.id
#   cidr_block              = "172.29.2.0/24"
#   availability_zone       = "ap-southeast-1b"
#   map_public_ip_on_launch = true

#   tags = {
#     Name = "nam-vpc-public-subnet-a"
#   }
# }

# resource "aws_internet_gateway" "igw" {
#   vpc_id = aws_vpc.main.id

#   tags = {
#     Name = "nam-vpc-igw"
#   }
# }

# resource "aws_route_table" "public_a" {
#   vpc_id = aws_vpc.main.id

#   route {
#     cidr_block = "0.0.0.0/0"
#     gateway_id = aws_internet_gateway.igw.id
#   }

#   tags = {
#     Name = "nam-vpc-route-table-public-a"
#   }
# }

# resource "aws_route_table" "public_b" {
#   vpc_id = aws_vpc.main.id

#   route {
#     cidr_block = "0.0.0.0/0"
#     gateway_id = aws_internet_gateway.igw.id
#   }

#   tags = {
#     Name = "nam-vpc-route-table-public-b"
#   }
# }

# resource "aws_route_table_association" "public_assoc_a" {
#   subnet_id      = aws_subnet.public_a.id
#   route_table_id = aws_route_table.public_a.id
# }

# resource "aws_route_table_association" "public_assoc_b" {
#   subnet_id      = aws_subnet.public_b.id
#   route_table_id = aws_route_table.public_b.id
# }

# #
# # Load Balancer
# #
# resource "aws_security_group" "ecs_alb_sg" {
#   vpc_id = aws_vpc.main.id
#   name   = "nam-vpc-ecs-alb-sg"

#   ingress {
#     from_port   = 80
#     to_port     = 80
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

# resource "aws_lb" "ecs_alb" {
#   name               = "nam-vpc-ecs-public-alb"
#   internal           = false
#   load_balancer_type = "application"
#   subnets            = [aws_subnet.public_a.id, aws_subnet.public_b.id]
#   security_groups    = [aws_security_group.ecs_alb_sg.id]
# }

# resource "aws_lb_target_group" "ecs_tg" {
#   name        = "nam-vpc-ecs-tg"
#   vpc_id      = aws_vpc.main.id
#   target_type = "ip"
#   port        = 8000
#   protocol    = "HTTP"

#   health_check {
#     path                = "/health"
#     interval            = 30
#     timeout             = 5
#     healthy_threshold   = 3
#     unhealthy_threshold = 3
#   }
# }

# resource "aws_lb_listener" "ecs_listener" {
#   load_balancer_arn = aws_lb.ecs_alb.arn
#   port              = 80
#   protocol          = "HTTP"

#   default_action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.ecs_tg.arn
#   }
# }

# #
# # ECS
# #
# resource "aws_ecs_cluster" "main" {
#   name = "go-somewhere-cluster"
# }

# data "aws_iam_policy_document" "ecs_task_assume" {
#   statement {
#     actions = ["sts:AssumeRole"]
#     principals {
#       type        = "Service"
#       identifiers = ["ecs-tasks.amazonaws.com"]
#     }
#   }
# }

# resource "aws_iam_role" "ecs_task_execution_role" {
#   name               = "nam-vpc-ecs-task-role"
#   assume_role_policy = data.aws_iam_policy_document.ecs_task_assume.json
# }

# resource "aws_iam_role_policy_attachment" "ecs_logs_policy" {
#   role       = aws_iam_role.ecs_task_execution_role.name
#   policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
# }

# resource "aws_cloudwatch_log_group" "ecs_app_logs" {
#   name              = "/ecs/go-somewhere/backend"
# }

# resource "aws_ecs_task_definition" "backend" {
#   family                    = "web-task"
#   network_mode              = "awsvpc"
#   requires_compatibilities  = ["FARGATE"]
#   cpu                       = "256"
#   memory                    = "512"
#   execution_role_arn        = aws_iam_role.ecs_task_execution_role.arn
#   container_definitions     = jsonencode([
#     {
#       name      = "backend"
#       image     = "832557411742.dkr.ecr.ap-southeast-1.amazonaws.com/go-somewhere/backend:1.0.0"
#       essential = true
#       portMappings = [
#         {
#           containerPort = 8000
#           hostPort      = 8000
#         }
#       ]
#       logConfiguration = {
#         logDriver = "awslogs"
#         options = {
#           awslogs-group         = aws_cloudwatch_log_group.ecs_app_logs.name
#           awslogs-region        = "ap-southeast-1"
#           awslogs-stream-prefix = "ecs"
#         }
#       }
#       # healthCheck = {
#       #   command = ["CMD-SHELL", "curl -f http://localhost:8000/health || exit 1"],
#       #   interval = 30,
#       #   timeout = 5,
#       #   retries = 3,
#       #   startPeriod = 120
#       # }
#     }
#   ])
# }

# resource "aws_security_group" "ecs_sg" {
#   vpc_id = aws_vpc.main.id
#   name   = "nam-vpc-ecs-sg"

#   ingress {
#     from_port       = 8000
#     to_port         = 8000
#     protocol        = "tcp"
#     security_groups = [aws_security_group.ecs_alb_sg.id]
#   }

#   egress {
#     from_port   = 0
#     to_port     = 0
#     protocol    = "-1"
#     cidr_blocks = ["0.0.0.0/0"]
#   }
# }

# resource "aws_ecs_service" "backend" {
#   name            = "backend"
#   cluster         = aws_ecs_cluster.main.id
#   task_definition = aws_ecs_task_definition.backend.arn
#   desired_count   = 1
#   launch_type     = "FARGATE"

#   network_configuration {
#     subnets          = [aws_subnet.public_a.id, aws_subnet.public_b.id]
#     security_groups  = [aws_security_group.ecs_sg.id]
#     assign_public_ip = true
#   }

#   load_balancer {
#     target_group_arn = aws_lb_target_group.ecs_tg.arn
#     container_name   = "backend"
#     container_port   = 8000
#   }

#   depends_on = [aws_lb_listener.ecs_listener]
# }
