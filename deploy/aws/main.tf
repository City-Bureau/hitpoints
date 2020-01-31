provider "aws" {
  region = var.region
}

resource "aws_iam_role" "hitpoints_role" {
  name               = "hitpoints_iam_role"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_instance_profile" "hitpoints_profile" {
  name = "hitpoints_profile"
  role = "hitpoints_iam_role"
}

resource "aws_iam_role_policy" "hitpoints_role_policy" {
  name   = "hitpoints_role_policy"
  role   = aws_iam_role.hitpoints_role.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": ["${aws_s3_bucket.hitpoints.arn}"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:DeleteObject"
      ],
      "Resource": ["${aws_s3_bucket.hitpoints.arn}/*"]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "hitpoints" {
  bucket = "${var.prefix}-output"
  acl    = "private"
  tags   = var.tags

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_security_group" "hitpoints" {
  name   = "Hitpoints Security Group"
  vpc_id = var.vpc_id
  tags   = var.tags

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.allow_ssh ? ["0.0.0.0/0"] : []
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 943
    to_port     = 943
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 1194
    to_port     = 1194
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "hitpoints" {
  ami                         = data.aws_ami.ubuntu.id
  instance_type               = var.instance_type
  key_name                    = var.keyname
  security_groups             = [aws_security_group.hitpoints.id]
  subnet_id                   = var.subnet_id
  iam_instance_profile        = aws_iam_instance_profile.hitpoints_profile.id
  associate_public_ip_address = true
  tags                        = var.tags
  volume_tags                 = var.tags

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = self.public_ip
    private_key = file(var.ssh_priv_key_path)
  }

  provisioner "file" {
    content = templatefile("${path.module}/../files/hitpoints.service", {
      ssl          = var.ssl
      domain       = var.domain
      command      = "s3"
      command_args = "--role --bucket ${aws_s3_bucket.hitpoints.id}"
    })
    destination = "/home/ubuntu/hitpoints.service"
  }

  provisioner "file" {
    source      = "${path.module}/../../release/hitpoints-linux-amd64/hitpoints"
    destination = "/home/ubuntu/hitpoints"
  }

  provisioner "remote-exec" {
    script = "${path.module}/../files/setup-hitpoints.sh"
  }

  lifecycle {
    ignore_changes = [security_groups]
  }
}

data "aws_ami" "ubuntu" {
  most_recent = "true"

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"]
}
