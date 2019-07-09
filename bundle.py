#!/usr/bin/env python3


from os import path
import sys
import re


def process(file_path, indent = '', const_list = ''):
    f = open(file_path)
    dirname = path.dirname(path.realpath(file_path))
    lines = f.readlines()
    for line in lines:
        match = re.match(r"^([ \t]*)'<include> ([^']+)';", line)
        match_const = re.match(r"^([ \t]*)'<constants>';", line)
        if match:
            module_indent = match.group(1)
            module_file = match.group(2)
            process(path.join(dirname, module_file), indent + module_indent)
        elif match_const:
            const_indent = match_const.group(1)
            for statement in const_list:
                print(const_indent + statement)
            print('')
        else:
            print(indent + line, end='')
    f.close()


def main():
    entry_file = sys.argv[1]
    const_file = sys.argv[2]
    const_table = {}
    const_values = {}
    with open(const_file) as f:
        for line in f.readlines():
            match = re.match (
                r'^[ \t]*(const)*[ \t]*([A-Z_]+) = "([A-Za-z0-9_]+)"',
                line
            )
            if match:
                name = match.group(2)
                value = match.group(3)
                if const_values.get(value, False):
                    print('Duplicate Const Value: %s' % value, file=sys.stderr)
                    sys.exit(1)
                const_table[name] = value
                const_values[value] = True
    const_list = []
    for name, value in const_table.items():
        const_list.append("const %s = '%s'" % (name, value))
    process(entry_file, const_list=const_list)


main()
