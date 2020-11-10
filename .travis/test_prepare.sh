#!/bin/bash -ex

SCRIPT=`realpath -s $0`
SCRIPTPATH=`dirname $SCRIPT`

source $SCRIPTPATH/test_common.sh

pushd $srcdir
rm -rf ovs
rm -rf ovn
rm -rf $sandbox
mkdir $sandbox

git clone --depth 1 -b master https://github.com/openvswitch/ovs.git
git clone --depth 1 -b master https://github.com/ovn-org/ovn.git
pushd ovs
./boot.sh && ./configure --enable-silent-rules
make -j4

popd
popd
