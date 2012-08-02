#!/bin/sh

set -e

distdir=$(cd "$1/" && pwd)
tmpdir=$(cd "$2/" && pwd)

venv=$tmpdir/venv
pkg_path=go1.0.2.linux-amd64

##
## Install go
##

mkdir -p $tmpdir/$pkg_path
cd $tmpdir/$pkg_path
curl -sO http://go.googlecode.com/files/$pkg_path.tar.gz
tar zxf $pkg_path.tar.gz
rm -f $pkg_path.tar.gz
cd -

##
## Install hg + bzr
##

python "$distdir/virtualenv-1.7/virtualenv.py" --python python2.7 --distribute --never-download $venv
. $venv/bin/activate
pip install --use-mirrors mercurial
pip install --use-mirrors bzr

