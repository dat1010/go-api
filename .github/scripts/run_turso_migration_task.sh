#!/bin/bash
set -euo pipefail

if [[ -z "${CLUSTER:-}" || -z "${SERVICE:-}" || -z "${TASK_DEFINITION:-}" ]]; then
  echo "CLUSTER, SERVICE, and TASK_DEFINITION must be set"
  exit 1
fi

if [[ -z "${TURSO_DATABASE_URL:-}" || -z "${TURSO_AUTH_TOKEN:-}" ]]; then
  echo "TURSO_DATABASE_URL and TURSO_AUTH_TOKEN must be set"
  exit 1
fi

if [[ -z "${DB_SECRET_ARN:-}" && -z "${PG_DSN:-}" ]]; then
  echo "DB_SECRET_ARN or PG_DSN must be set"
  exit 1
fi

TASK_DEFINITION_JSON=$(aws ecs describe-task-definition \
  --task-definition "$TASK_DEFINITION" \
  --query 'taskDefinition' \
  --output json)

CONTAINER_NAME=$(echo "$TASK_DEFINITION_JSON" | jq -r '.containerDefinitions[0].name')

NETWORK_CONFIG=$(aws ecs describe-services \
  --cluster "$CLUSTER" \
  --services "$SERVICE" \
  --query 'services[0].networkConfiguration.awsvpcConfiguration' \
  --output json)

OVERRIDES=$(jq -n \
  --arg name "$CONTAINER_NAME" \
  --arg tursoURL "$TURSO_DATABASE_URL" \
  --arg tursoToken "$TURSO_AUTH_TOKEN" \
  --arg dbSecretArn "${DB_SECRET_ARN:-}" \
  --arg pgDsn "${PG_DSN:-}" \
  '{
    containerOverrides: [
      {
        name: $name,
        command: ["./turso_migrate"],
        environment: [
          {name: "TURSO_DATABASE_URL", value: $tursoURL},
          {name: "TURSO_AUTH_TOKEN", value: $tursoToken}
        ]
      }
    ]
  }
  | if $dbSecretArn != "" then
      .containerOverrides[0].environment += [{name: "DB_SECRET_ARN", value: $dbSecretArn}]
    else
      .
    end
  | if $pgDsn != "" then
      .containerOverrides[0].environment += [{name: "PG_DSN", value: $pgDsn}]
    else
      .
    end')

echo "Starting one-off Turso migration task..."

TASK_ARN=$(aws ecs run-task \
  --cluster "$CLUSTER" \
  --task-definition "$TASK_DEFINITION" \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration=$(echo "$NETWORK_CONFIG" | jq -c '.')" \
  --overrides "$OVERRIDES" \
  --query 'tasks[0].taskArn' \
  --output text)

if [[ "$TASK_ARN" == "None" || -z "$TASK_ARN" ]]; then
  echo "Failed to start migration task."
  exit 1
fi

echo "Waiting for task to complete: $TASK_ARN"
aws ecs wait tasks-stopped --cluster "$CLUSTER" --tasks "$TASK_ARN"

EXIT_CODE=$(aws ecs describe-tasks \
  --cluster "$CLUSTER" \
  --tasks "$TASK_ARN" \
  --query "tasks[0].containers[?name=='$CONTAINER_NAME'].exitCode | [0]" \
  --output text)

if [[ "$EXIT_CODE" != "0" ]]; then
  echo "Migration task failed with exit code $EXIT_CODE."
  exit 1
fi

echo "Migration task completed successfully."
