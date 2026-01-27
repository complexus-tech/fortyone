# Fortyone AWS Deployment Plan (ECS Fargate + RDS + Docker Hub)

This document captures the finalized, beginner-friendly plan for deploying the Fortyone backend on AWS using ECS Fargate, RDS Postgres, Docker Hub images, and SSM tunneling for private access.

## Overview

- Compute: ECS Fargate (private subnets, no public IPs)
- Database: RDS Postgres (private subnets)
- Registry: Docker Hub
- Access: SSM port forwarding for Asynq UI, Jaeger, and DB
- Optional public API: ALB DNS (no custom domain required)
- Cost optimizations: NAT + S3 Gateway VPC Endpoint

## Naming Prefix

Use the prefix `fortyone` for all resources.

## Networking (2 AZs)

### VPC

- `fortyone-vpc` — `10.0.0.0/16`

### Subnets

Public:

- `fortyone-public-a` — `10.0.1.0/24`
- `fortyone-public-b` — `10.0.2.0/24`

Private app:

- `fortyone-private-app-a` — `10.0.11.0/24`
- `fortyone-private-app-b` — `10.0.12.0/24`

Private db:

- `fortyone-private-db-a` — `10.0.21.0/24`
- `fortyone-private-db-b` — `10.0.22.0/24`

### Internet + NAT

- Internet Gateway: `fortyone-igw`
- NAT Gateway: `fortyone-nat-a` (in `fortyone-public-a`)
- Elastic IP: `fortyone-nat-eip`

### Route Tables

- `fortyone-rt-public` → `0.0.0.0/0` to IGW
- `fortyone-rt-private` → `0.0.0.0/0` to NAT

## VPC Endpoints (Cost Optimization)

- S3 Gateway Endpoint attached to `fortyone-rt-private`

Optional later if NAT costs grow:

- Interface endpoints: `logs`, `ssm`, `secretsmanager`, `kms`, `ecr.api`, `ecr.dkr`

## Security Groups

- `fortyone-sg-ecs`
  - Inbound: none
  - Outbound: allow all
- `fortyone-sg-rds`
  - Inbound: 5432 from `fortyone-sg-ecs`
- Optional ALB SG: `fortyone-sg-alb`
  - Inbound: 80/443 from internet (or your IPs)
- Optional SSM tunnel EC2 SG: `fortyone-sg-ssm`
  - Inbound: none
  - Outbound: allow all

## RDS (Private)

- Instance: `fortyone-db`
- Subnet group: `fortyone-db-subnet-group`
- Engine: Postgres
- Class: `db.t4g.micro`
- Public access: disabled
- Security group: `fortyone-sg-rds`
- Store credentials in Secrets Manager

## ECS Cluster + IAM

- Cluster: `fortyone-ecs-cluster`
- Task execution role: `fortyone-ecs-task-exec-role` (AmazonECSTaskExecutionRolePolicy)
- Task role: `fortyone-ecs-task-role` (Secrets Manager/SSM if used)

### Service Discovery

- Cloud Map namespace: `fortyone.local`
- Services:
  - `api.fortyone.local`
  - `worker.fortyone.local`
  - `jaeger.fortyone.local`
  - `asynq-ui.fortyone.local` (if separate)

## Task Definitions

- `fortyone-api-task` (Docker Hub image)
- `fortyone-worker-task` (Docker Hub image)
- `fortyone-jaeger-task` (official Jaeger image)
- `fortyone-asynq-ui-task` (if separate)

### Logs (CloudWatch)

- `/ecs/fortyone-api`
- `/ecs/fortyone-worker`
- `/ecs/fortyone-jaeger`
- `/ecs/fortyone-asynq-ui`

## ECS Services (Private Subnets)

All services:

- Subnets: `fortyone-private-app-*`
- Assign public IP: disabled
- Security group: `fortyone-sg-ecs`

Services:

- `fortyone-api-svc` (desired count 1)
- `fortyone-worker-svc` (desired count 1)
- `fortyone-jaeger-svc` (desired count 1)
- `fortyone-asynq-ui-svc` (desired count 1, if separate)

## Optional Public API (AWS-generated URL)

If you want a public endpoint without a custom domain:

- ALB: `fortyone-alb` (public subnets)
- Target group: `fortyone-api-tg`
- Use ALB DNS: `fortyone-alb-xxxx.<region>.elb.amazonaws.com`

Note: HTTPS on ALB requires your own domain + ACM cert. If you avoid a custom domain, use HTTP on the ALB DNS for now.

## SSM Tunneling (Private Access)

Create a tiny EC2 instance for port forwarding:

- EC2 instance: `fortyone-ssm-tunnel`
- Subnet: `fortyone-public-a`
- SG: `fortyone-sg-ssm`
- IAM role: `AmazonSSMManagedInstanceCore`

Tunnel targets:

- Asynq UI: `asynq-ui.fortyone.local:8080` → `localhost:8080`
- Jaeger UI: `jaeger.fortyone.local:16686` → `localhost:16686`
- RDS: `fortyone-db.<region>.rds.amazonaws.com:5432` → `localhost:5432`

## AWS Console Locations (Beginner Friendly)

- VPC: `VPC` → VPCs, Subnets, Route Tables, Internet Gateways, NAT Gateways, Endpoints
- ECS: `ECS` → Clusters, Task Definitions, Services
- RDS: `RDS` → Databases, Subnet Groups
- IAM: `IAM` → Roles
- CloudWatch Logs: `CloudWatch` → Logs
- EC2: `EC2` → Instances (for SSM tunnel)
- SSM: `Systems Manager` → Session Manager
- ALB: `EC2` → Load Balancers
- Secrets Manager: `Secrets Manager`

## Reference Docs

- VPC: https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html
- ECS Fargate: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/getting-started-fargate.html
- RDS Postgres: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html
- SSM Session Manager: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html
- VPC endpoints (S3): https://docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-s3.html
- ECS service discovery: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-discovery.html
