#!/bin/bash
set -e

# Store the task family name separately
TASK_FAMILY=$TASK_DEFINITION

# Get the current task definition
TASK_DEFINITION_JSON=$(aws ecs describe-task-definition \
  --task-definition $TASK_FAMILY \
  --query 'taskDefinition' \
  --output json)

# Debug - see the current task definition structure
echo "Current Task Definition:"
echo "$TASK_DEFINITION_JSON" | jq '.containerDefinitions[].name'

# Find the index of the "web" container
WEB_CONTAINER_INDEX=$(echo "$TASK_DEFINITION_JSON" | jq 'map(.containerDefinitions[].name == "web") | index(true)')
if [ "$WEB_CONTAINER_INDEX" == "null" ]; then
  echo "Cannot find 'web' container in task definition. Container names are:"
  echo "$TASK_DEFINITION_JSON" | jq '.containerDefinitions[].name'
  exit 1
fi

# Update the task definition with the new image and secrets
NEW_TASK_DEFINITION=$(echo "$TASK_DEFINITION_JSON" | jq --arg IMAGE "$AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-api:$VERSION" \
  --arg VERSION "$VERSION" \
  --arg SECRET_ARN "$SECRET_ARN" \
  '.containerDefinitions[0].image = $IMAGE
  | .containerDefinitions[0].environment += [
      { "name": "VERSION", "value": $VERSION },
      { "name": "USE_HTTPS", "value": "true" }
    ]
  | .containerDefinitions[0].secrets = [
      { "name": "MY_LITTLE_SECRET",   "valueFrom": "\($SECRET_ARN):my_little_secret::" },
      { "name": "AUTH0_DOMAIN",       "valueFrom": "\($SECRET_ARN):AUTH0_DOMAIN::" },
      { "name": "AUTH0_CLIENT_ID",    "valueFrom": "\($SECRET_ARN):AUTH0_CLIENT_ID::" },
      { "name": "AUTH0_CLIENT_SECRET","valueFrom": "\($SECRET_ARN):AUTH0_CLIENT_SECRET::" },
      { "name": "AUTH0_CALLBACK_URL", "valueFrom": "\($SECRET_ARN):AUTH0_CALLBACK_URL::" },
      { "name": "SSL_CERT_PATH",      "valueFrom": "\($SECRET_ARN):SSL_CERT_PATH::" },
      { "name": "SSL_KEY_PATH",       "valueFrom": "\($SECRET_ARN):SSL_KEY_PATH::" }
    ]')

# Ensure port 8080 exists, then add 80 and 443 if they don't exist
NEW_TASK_DEFINITION=$(echo "$NEW_TASK_DEFINITION" | jq '
  # Ensure container has portMappings
  if .containerDefinitions[0].portMappings == null then
    .containerDefinitions[0].portMappings = []
  else . end |
  
  # Ensure port 8080 exists
  if ([.containerDefinitions[0].portMappings[] | select(.containerPort == 8080)] | length) == 0 then
    .containerDefinitions[0].portMappings += [{"containerPort": 8080, "hostPort": 8080, "protocol": "tcp"}]
  else . end |
  
  # Add port 443 if missing
  if ([.containerDefinitions[0].portMappings[] | select(.containerPort == 443)] | length) == 0 then
    .containerDefinitions[0].portMappings += [{"containerPort": 443, "hostPort": 443, "protocol": "tcp"}]
  else . end |
  
  # Add port 80 if missing
  if ([.containerDefinitions[0].portMappings[] | select(.containerPort == 80)] | length) == 0 then
    .containerDefinitions[0].portMappings += [{"containerPort": 80, "hostPort": 80, "protocol": "tcp"}]
  else . end
')

# Debug - see the updated task definition
echo "Updated port mappings:"
echo "$NEW_TASK_DEFINITION" | jq '.containerDefinitions[0].portMappings'

# Extract CPU and memory values if they exist
CPU_VALUE=$(echo "$TASK_DEFINITION_JSON" | jq -r '.cpu')
MEMORY_VALUE=$(echo "$TASK_DEFINITION_JSON" | jq -r '.memory')

CPU_PARAM=""
if [ "$CPU_VALUE" != "null" ]; then
  CPU_PARAM="--cpu $CPU_VALUE"
fi

MEMORY_PARAM=""
if [ "$MEMORY_VALUE" != "null" ]; then
  MEMORY_PARAM="--memory $MEMORY_VALUE"
fi

# Register the new task definition and capture its ARN
NEW_TASK_DEF_ARN=$(aws ecs register-task-definition \
  --family $TASK_FAMILY \
  --container-definitions "$(echo "$NEW_TASK_DEFINITION" | jq '.containerDefinitions')" \
  $CPU_PARAM $MEMORY_PARAM \
  --task-role-arn "$(echo "$TASK_DEFINITION_JSON" | jq -r '.taskRoleArn')" \
  --execution-role-arn "$(echo "$TASK_DEFINITION_JSON" | jq -r '.executionRoleArn')" \
  --network-mode "$(echo "$TASK_DEFINITION_JSON" | jq -r '.networkMode')" \
  --query 'taskDefinition.taskDefinitionArn' --output text)

# Update the ECS service with the new task definition
aws ecs update-service \
  --cluster $CLUSTER \
  --service $SERVICE \
  --task-definition $NEW_TASK_DEF_ARN \
  --force-new-deployment
