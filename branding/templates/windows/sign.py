#!/usr/bin/env python3
"""
This script is expected to be called from the main makefile, that should pass
the content of the WIN_CERT_PASS variable as the second argument.
"""
import subprocess
import os
import sys

WIN_CERT_PATH = sys.argv[1]
WIN_CERT_PASS = sys.argv[2]
SIGNTOOL = "signtool"

VERSION = subprocess.run(
    'git describe --tags',
    stdout=subprocess.PIPE).stdout.strip()

installer = "RiseupVPN-" + str(VERSION, 'utf-8') + '.exe'
target = str(os.path.join(os.path.abspath('.'), 'dist', installer))
cmd = [SIGNTOOL, "sign", "/f", WIN_CERT_PATH, "/p", WIN_CERT_PASS, target]
subprocess.run(cmd)
