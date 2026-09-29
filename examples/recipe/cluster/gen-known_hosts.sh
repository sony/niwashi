#!/bin/bash

. ./.env

# Retrieve SSH host keys for both containers using ssh-keyscan
echo "Retrieving SSH host keys..."
{
  ssh-keyscan -p ${CP_SSH_PORT} 127.0.0.1 2>/dev/null
  ssh-keyscan -p ${WORKER_SSH_PORT} 127.0.0.1 2>/dev/null
} > known_hosts
echo "Done! known_hosts file created."
