#!/bin/sh

check_and_copy() {
    if [ ! -e /sonic/$1 ]; then
        mkdir -p /sonic/$1
        cp -Rf /app/$1/* /sonic/$1/
    fi
}

make_and_copy() {
    mkdir -p /sonic/$1
    cp -Rf /app/$1/* /sonic/$1/
}

make_and_copy 'resources/admin'
make_and_copy 'resources/template/common'
check_and_copy 'conf'
check_and_copy 'resources/template/theme'

# Ensure default theme metadata exists even if volume already has partial files.
if [ ! -e /sonic/resources/template/theme/default-theme-anatole/theme.yml ]; then
    mkdir -p /sonic/resources/template/theme/default-theme-anatole
    cp -Rf /app/resources/template/theme/default-theme-anatole/* /sonic/resources/template/theme/default-theme-anatole/
fi

