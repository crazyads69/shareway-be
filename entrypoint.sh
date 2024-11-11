#!/bin/sh

set -e

# Function to wait for a service
wait_for_service() {
    host="$1"
    port="$2"
    service_name="$3"
    echo "Waiting for $service_name to be ready..."
    while ! nc -z "$host" "$port"; do
        echo "$service_name is unavailable - sleeping"
        sleep 1
    done
    echo "$service_name is up and running!"
}

# Wait for PostgreSQL
wait_for_service "postgres_db" "5432" "PostgreSQL"

# Wait for Redis
wait_for_service "redis" "6379" "Redis"

# Wait for RabbitMQ
# wait_for_service "rabbitmq" "5672" "RabbitMQ"

echo "All services are up - starting the application"
exec "$@"