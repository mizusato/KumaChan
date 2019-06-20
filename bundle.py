#!/usr/bin/env python3


from os import path
import sys
import re


def process(file_path, indent = ''):
    f = open(file_path)
    dirname = path.dirname(path.realpath(file_path))
    lines = f.readlines()
    for line in lines:
        match = re.match(r"^([ \t]*)'<include> ([^']+)';", line)
        if match:
            module_indent = match.group(1)
            module_file = match.group(2)
            process(path.join(dirname, module_file), indent + module_indent)
        else:
            print(indent + line, end='')
    f.close()


def main():
    entry_file = sys.argv[1]
    process(entry_file)


main()
