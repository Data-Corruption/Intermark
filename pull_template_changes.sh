#!/bin/bash

set -e  # Exit on any error
set -o pipefail  # Catch errors in piped commands

log_info() { echo -e "$1"; }
log_success() { echo -e "🟢 $1"; }
log_warning() { echo -e "🟡 $1"; }
log_error() { echo -e "🔴 $1"; exit 1; }

# Check required commands
for cmd in curl npm; do
    command -v "$cmd" >/dev/null 2>&1 || log_error "Missing required command: $cmd"
done

# Install NPM dependencies if package.json exists
if [ -f "package.json" ]; then
    log_info "Installing NPM dependencies..."
    if [ -f "package-lock.json" ]; then
        npm ci && log_success "NPM dependencies installed using 'npm ci'!"
    else
        npm install && log_success "NPM dependencies installed using 'npm install'!"
    fi
else
    log_error "Missing package.json found. Skipping NPM install."
fi

UPSTREAM_REPO="https://github.com/Data-Corruption/Intermark.git"
UPSTREAM_EXISTS=$(git remote | grep upstream)

# Add the upstream remote if not already added
if [ -z "$UPSTREAM_EXISTS" ]; then
    log_info "Adding upstream repository: $UPSTREAM_REPO"
    git remote add upstream $UPSTREAM_REPO
fi

log_info "Fetching updates from upstream..."
git fetch upstream

log_info "Merging upstream/main into your current branch..."
git merge upstream/main
if [ $? -ne 0 ]; then
    log_error "Issue merging upstream/main into your current branch."
fi

log_success "Successfully merged upstream/main into your current branch!"