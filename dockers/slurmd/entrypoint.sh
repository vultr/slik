#!/bin/bash
set -euo pipefail

# will not work if container is not running as privileged + root
#echo $HOSTNAME > /proc/sys/kernel/hostname

slurmd -D -v -f /etc/slurm/slurm.conf