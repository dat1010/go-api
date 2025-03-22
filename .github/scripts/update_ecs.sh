#!/bin/bash
set -e

# Get the current task definition
TASK_DEFINITION=$(aws ecs describe-task-definition \
  --task-definition $TASK_DEFINITION \
  --query 'taskDefinition' \
  --output json)

# Update the task definition with the new image and secrets
NEW_TASK_DEFINITION=$(echo "$TASK_DEFINITION" | jq --arg IMAGE "$AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-api:$VERSION" \
  --arg VERSION "$VERSION" \
  '.containerDefinitions[0].image = $IMAGE 
  | .containerDefinitions[0].environment += [
      {
        "name": "VERSION",
        "value": $VERSION
      }
    ]
  | .containerDefinitions[0].secrets = ([
      {
        "name": "MY_LITTLE_SECRET",
        "valueFrom": "'"$SECRET_ARN"':my_little_secret::"
      }
    ])')

# Extract CPU and memory values if they exist
CPU_VALUE=$(echo "$TASK_DEFINITION" | jq -r '.cpu')
MEMORY_VALUE=$(echo "$TASK_DEFINITION" | jq -r '.memory')

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
  --family $TASK_DEFINITION \
  --container-definitions "$(echo "$NEW_TASK_DEFINITION" | jq '.containerDefinitions')" \
  $CPU_PARAM $MEMORY_PARAM \
  --task-role-arn "$(echo "$TASK_DEFINITION" | jq -r '.taskRoleArn')" \
  --execution-role-arn "$(echo "$TASK_DEFINITION" | jq -r '.executionRoleArn')" \
  --network-mode "$(echo "$TASK_DEFINITION" | jq -r '.networkMode')" \
  --query 'taskDefinition.taskDefinitionArn' --output text)

# Update the ECS service with the new task definition
aws ecs update-service \
  --cluster $CLUSTER \
  --service $SERVICE \
  --task-definition $NEW_TASK_DEF_ARN \
  --force-new-deployment