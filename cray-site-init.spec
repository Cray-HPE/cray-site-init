# Copyright 2021 Hewlett Packard Enterprise Development LP
Name: cray-site-init
License: MIT License
Summary: HPCaaS configuration and deployment tool for bare-metal or reinstallations
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Provides: csi

%ifarch %ix86
    %global GOARCH 386
%endif
%ifarch x86_64
    %global GOARCH amd64
%endif
%description
Installs the Cray Site Initiator GoLang binary onto a Linux system.

%prep
%setup -q

%build
CGO_ENABLED=0
GOOS=linux
GOARCH="%{GOARCH}"
GO111MODULE=on
export CGO_ENABLED GOOS GOARCH GO111MODULE

make build

%install
CGO_ENABLED=0
GOOS=linux
GOARCH="%{GOARCH}"
GO111MODULE=on
export CGO_ENABLED GOOS GOARCH GO111MODULE

mkdir -pv ${RPM_BUILD_ROOT}/usr/bin/
cp -pv bin/csi ${RPM_BUILD_ROOT}/usr/bin/csi

mkdir -pv ${RPM_BUILD_ROOT}/usr/local/bin/
cp -pv scripts/write-livecd.sh ${RPM_BUILD_ROOT}/usr/local/bin/write-livecd.sh

%clean

%files
%license LICENSE
%defattr(755,root,root)
/usr/bin/csi
/usr/local/bin/write-livecd.sh

%changelog