#!/bin/bash

result=0

for filename in `find -maxdepth 1 -name "*.go"` ; do
    newfile=`mktemp` || exit 1
    gofmt ${filename} > "${newfile}" 2>> /dev/null
    diff -u -p "${filename}" "${newfile}"
    r=$?
    rm "${newfile}"
    if [ $r != 0 ]; then
        echo "================================================================================================="
        echo "Code format error in $filename"
        echo "================================================================================================="
        result=$r
    fi
done

exit $result
