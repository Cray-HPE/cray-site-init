# Copyright 2020 Cray Inc. All Rights Reserved.
Name: shasta-instance-control
License: Cray Proprietary
BuildArchitectures: noarch
Summary: Control shasta instances; bare-metal and deployed
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Requires: podman

%description
This tool enables control of a shasta instance by local or remote access. See usage for more info.

%prep
%setup -q

%build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on make build

%install
mkdir -pv ${RPM_BUILD_ROOT}/usr/bin/
cp -pv bin/sic ${RPM_BUILD_ROOT}/usr/bin/sic

%clean

%files
%license LICENSE
%defattr(755,root,root)
/usr/bin/sic

%changelog
