#!/bin/bash
set -e

# Store the task family name separately
TASK_FAMILY=$TASK_DEFINITION

# Get the current task definition
TASK_DEFINITION_JSON=$(aws ecs describe-task-definition \
  --task-definition $TASK_FAMILY \
  --query 'taskDefinition' \
  --output json)

# Debug - print container definition
echo "Current container definitions:"
echo "$TASK_DEFINITION_JSON" | jq -r '.containerDefinitions'

# Update the task definition with the new image and secrets
NEW_TASK_DEFINITION=$(echo "$TASK_DEFINITION_JSON" | jq --arg IMAGE "$AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-api:$VERSION" \
  --arg VERSION "$VERSION" \
  --arg SECRET_ARN "$SECRET_ARN" \
  --arg DB_SECRET_ARN "$DB_SECRET_ARN" \
  '.containerDefinitions[0].image = $IMAGE
  | .containerDefinitions[0].environment += [
      { "name": "VERSION", "value": $VERSION },
      { "name": "USE_HTTPS", "value": "true" },
      { "name": "DB_SECRET_ARN", "value": $DB_SECRET_ARN }
    ]
  | .containerDefinitions[0].secrets = [
      { "name": "MY_LITTLE_SECRET",   "valueFrom": "\($SECRET_ARN):my_little_secret::" },
      { "name": "AUTH0_DOMAIN",       "valueFrom": "\($SECRET_ARN):AUTH0_DOMAIN::" },
      { "name": "AUTH0_AUDIENCE",     "valueFrom": "\($SECRET_ARN):AUTH0_AUDIENCE::" },
      { "name": "AUTH0_CLIENT_ID",    "valueFrom": "\($SECRET_ARN):AUTH0_CLIENT_ID::" },
      { "name": "AUTH0_CLIENT_SECRET","valueFrom": "\($SECRET_ARN):AUTH0_CLIENT_SECRET::" },
      { "name": "AUTH0_CALLBACK_URL", "valueFrom": "\($SECRET_ARN):AUTH0_CALLBACK_URL::" },
      { "name": "TURSO_DATABASE_URL", "valueFrom": "\($SECRET_ARN):TURSO_DATABASE_URL::" },
      { "name": "TURSO_AUTH_TOKEN",   "valueFrom": "\($SECRET_ARN):TURSO_AUTH_TOKEN::" },
      { "name": "LAMBDA_ARN",   "valueFrom": "\($SECRET_ARN):LAMBDA_ARN::" }
    ]')

# Add port 8080 separately, ensuring we don't duplicate
echo "Adding port 8080 if it doesn't exist"
NEW_TASK_DEFINITION=$(echo "$NEW_TASK_DEFINITION" | jq '
  # First check if port 8080 is missing
  if (.containerDefinitions[0].portMappings | map(select(.containerPort == 8080)) | length) == 0 then
    # Port 8080 is missing, so add it
    .containerDefinitions[0].portMappings += [{"containerPort": 8080, "hostPort": 8080, "protocol": "tcp"}]
  else
    # Port 8080 already exists
    .
  end
')

# Debug - print updated port mappings
echo "Updated port mappings:"
echo "$NEW_TASK_DEFINITION" | jq -r '.containerDefinitions[0].portMappings'

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
