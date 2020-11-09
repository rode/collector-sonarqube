#!/bin/bash

set -e

cd $GITHUB_WORKSPACE

skaffold build -t check
docker run rode-collector-sonarqube:check curl -s https://codecov.io/bash | bash
