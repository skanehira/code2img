#!/bin/bash
if [ ! -e tmp ];then
  mkdir tmp
fi

code=testdata/test.go

./code2img -l $code tmp/has_line.png
./code2img $code tmp/no_line.png

files="has_line.png no_line.png"

for f in $files;do
  result=$(cmp tmp/$f testdata/$f)
  if [ -n "$result" ];then
    echo $result
    exit 1
  fi
done

rm -rf tmp
