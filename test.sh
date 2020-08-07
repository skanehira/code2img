#!/bin/bash
if [ ! -e tmp ];then
  mkdir tmp
fi

./code2img -l $0 tmp/has_line.png
./code2img $0 tmp/no_line.png

files="has_line.png no_line.png"

for f in $files;do
  cmp tmp/$f testdata/$f
done

rm -rf tmp
