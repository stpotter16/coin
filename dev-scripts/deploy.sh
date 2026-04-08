#!/usr/bin/env bash

# Fail on first error
set -e

# Fail on unset variable
set -u

# Configuration
TARGET_HOST="${TARGET_HOST:-}"
SERVICE_LINK="${SERVICE_LINK:-/opt/coin/server}"
TARGET_SYSTEM="${TARGET_SYSTEM:-x86_64-linux}"  # Target architecture

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

if [ -z "$TARGET_HOST" ]; then
    echo -e "${YELLOW}ERROR: TARGET_HOST environment variable must be set${NC}"
    echo "Usage: TARGET_HOST=user@server make server/deploy"
    exit 1
fi

echo -e "${BLUE}==> Building coin package for $TARGET_SYSTEM...${NC}"
nix build --system $TARGET_SYSTEM

# Get the store path of the built package
STORE_PATH=$(readlink -f ./result)
echo -e "${BLUE}Built: $STORE_PATH${NC}"

echo -e "${BLUE}==> Copying binary and dependencies to $TARGET_HOST...${NC}"
nix copy --to ssh://$TARGET_HOST ./result

echo -e "${BLUE}==> Updating symlink on server...${NC}"
ssh -t $TARGET_HOST "sudo mkdir -p $(dirname $SERVICE_LINK) && sudo ln -sf $STORE_PATH/bin/server $SERVICE_LINK"

echo -e "${BLUE}==> Restarting coin service...${NC}"
ssh -t $TARGET_HOST 'sudo systemctl restart coin'

echo -e "${GREEN}==> Deployment complete!${NC}"
echo -e "${GREEN}Server is running: $STORE_PATH/bin/server${NC}"

