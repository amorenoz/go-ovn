#!/bin/sh -ex

WORKDIR=${WORKDIR:-`pwd`}
ovn_srcdir=${OVN_SRCDIR:-${WORKDIR}}
mkdir -p ${ovn_srcdir}
sandbox=${ovn_srcdir}/sandbox
mkdir -p ${sandbox}

