terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "draios-dev-developer"

  default_tags {
    tags = {
      Environment = "test"
      Project     = "DeepDrift-Test"
      ManagedBy   = "Terraform"
      Purpose     = "DriftDetectionTest"
    }
  }
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.100.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "deepdrift-test-vpc"
  }
}

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.100.1.0/24"
  availability_zone       = "us-east-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "deepdrift-test-public-subnet"
    Tier = "public"
  }
}

# Private Subnet
resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.100.2.0/24"
  availability_zone = "us-east-1a"

  tags = {
    Name = "deepdrift-test-private-subnet"
    Tier = "private"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "deepdrift-test-igw"
  }
}

# Route Table for Public Subnet
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "deepdrift-test-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# Security Group for Web Server
resource "aws_security_group" "web" {
  name        = "deepdrift-test-web-sg"
  description = "Security group for web servers"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "HTTP from anywhere"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS from anywhere"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "deepdrift-test-web-sg"
  }
}

# Security Group for Database
resource "aws_security_group" "database" {
  name        = "deepdrift-test-db-sg"
  description = "Security group for database"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "PostgreSQL from web servers"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.web.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "deepdrift-test-db-sg"
  }
}

# EC2 Web Server
resource "aws_instance" "web" {
  ami                    = "ami-0c3e8df62015275ea" # Amazon Linux 2023
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.web.id]

  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y httpd
              systemctl start httpd
              systemctl enable httpd
              echo "<h1>DeepDrift Test Server</h1>" > /var/www/html/index.html
              EOF

  tags = {
    Name        = "deepdrift-test-web-server"
    Application = "WebServer"
  }
}

# S3 Bucket for Application Data
resource "aws_s3_bucket" "app_data" {
  bucket = "deepdrift-test-app-data-${random_id.bucket_suffix.hex}"

  tags = {
    Name        = "deepdrift-test-app-data"
    Purpose     = "ApplicationData"
    ContentType = "mixed"
  }
}

resource "aws_s3_bucket_versioning" "app_data" {
  bucket = aws_s3_bucket.app_data.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "app_data" {
  bucket = aws_s3_bucket.app_data.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 Bucket for Logs
resource "aws_s3_bucket" "logs" {
  bucket = "deepdrift-test-logs-${random_id.bucket_suffix.hex}"

  tags = {
    Name    = "deepdrift-test-logs"
    Purpose = "Logging"
  }
}

# Random ID for bucket names
resource "random_id" "bucket_suffix" {
  byte_length = 8
}

# RDS Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "deepdrift-test-db-subnet-group"
  subnet_ids = [aws_subnet.private.id, aws_subnet.private_2.id]

  tags = {
    Name = "deepdrift-test-db-subnet-group"
  }
}

# RDS Database
resource "aws_db_instance" "main" {
  identifier           = "deepdrift-test-db"
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  storage_type         = "gp3"
  db_name              = "deepdriftdb"
  username             = "dbadmin"
  password             = "ChangeMe123!" # TODO: Use AWS Secrets Manager in production
  skip_final_snapshot  = true
  publicly_accessible  = false
  db_subnet_group_name = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.database.id]

  tags = {
    Name = "deepdrift-test-database"
  }
}

# IAM Role for Lambda
resource "aws_iam_role" "lambda_exec" {
  name = "deepdrift-test-lambda-exec-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })

  tags = {
    Name = "deepdrift-test-lambda-exec-role"
  }
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_exec.name
}

# Lambda Function
resource "aws_lambda_function" "api" {
  filename         = "lambda_function.zip"
  function_name    = "deepdrift-test-api"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "index.handler"
  source_code_hash = filebase64sha256("lambda_function.zip")
  runtime          = "python3.11"
  timeout          = 30
  memory_size      = 256

  environment {
    variables = {
      ENVIRONMENT = "test"
      DB_HOST     = aws_db_instance.main.address
    }
  }

  tags = {
    Name = "deepdrift-test-api-lambda"
  }
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "deepdrift-test-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.web.id]
  subnets            = [aws_subnet.public.id, aws_subnet.public_2.id]

  tags = {
    Name = "deepdrift-test-alb"
  }
}

resource "aws_lb_target_group" "web" {
  name     = "deepdrift-test-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = {
    Name = "deepdrift-test-target-group"
  }
}

resource "aws_lb_target_group_attachment" "web" {
  target_group_arn = aws_lb_target_group.web.arn
  target_id        = aws_instance.web.id
  port             = 80
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.web.arn
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "app" {
  name              = "/aws/deepdrift-test/app"
  retention_in_days = 7

  tags = {
    Name = "deepdrift-test-app-logs"
  }
}

# CloudWatch Alarm
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "deepdrift-test-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
  alarm_actions       = []

  dimensions = {
    InstanceId = aws_instance.web.id
  }

  tags = {
    Name = "deepdrift-test-high-cpu-alarm"
  }
}
