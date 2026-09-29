#!/bin/bash

. ./.env

ssh-keygen -t ed25519 -f ${PRIVATE_KEY_PATH} -N ""
