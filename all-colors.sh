#!/bin/bash
colors="abap algol algol_nu api arduino autumn borland bw colorful dracula emacs friendly fruity github igor lovelace manni monokai monokailight murphy native paraiso-dark paraiso-light pastie perldoc pygments rainbow_dash rrt solarized-dark solarized-dark256 solarized-light swapoff tango trac vim vs xcode"

if [ ! -e tmp ];then
  mkdir tmp
fi

code=testdata/test.go

go build -o test

for c in $colors;do
  ./test -l -t $c $code tmp/$c.png
done

rm ./test
