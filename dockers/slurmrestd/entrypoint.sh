#!/bin/bash
set -euo pipefail

slurmrestd -v -f /etc/slurm/slurm.conf 0.0.0.0:6820