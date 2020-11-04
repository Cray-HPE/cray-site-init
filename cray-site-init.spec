# Copyright 2020 Cray Inc. All Rights Reserved.
Name: shasta-instance-control
License: Cray Proprietary
Summary: Control shasta instances; both bare-metal, and deployed
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Provides: sic
%ifarch %ix86
    %global GOARCH 386
%endif
%ifarch    x86_64
    %global GOARCH amd64
%endif
%description
This tool enables control of a shasta instance by local or remote access. See usage for more info.

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
cp -pv bin/sic ${RPM_BUILD_ROOT}/usr/bin/sic

%clean

%files
%license LICENSE
%defattr(755,root,root)
/usr/bin/sic

%changelog
