from fedora:30

RUN dnf install -y @'Development Tools' rpm-build dnf-plugins-core go libtool autoconf automake

ADD .travis /src

WORKDIR /src/
RUN ls

RUN sh test_prepare.sh
