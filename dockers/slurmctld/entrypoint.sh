#!/bin/bash
set -euo pipefail

slurmctld -D -v -f /etc/slurm/slurm.conf