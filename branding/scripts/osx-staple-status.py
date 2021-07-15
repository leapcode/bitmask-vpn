#!/usr/bin/env python3
import sys
arg = sys.argv[1]
line = [x for x in arg.split('\n') if x.startswith('Status:')]
print(line[0].split('Status: ')[-1])
