#!/usr/bin/env bash
set -e

# Use direct GOPROXY if needed to avoid proxy timeouts
export GOPROXY="${GOPROXY:-direct}"

# Build and package the workflow
make pack

# Load into Alfred
open "AlfredMaestro.alfredworkflow"
