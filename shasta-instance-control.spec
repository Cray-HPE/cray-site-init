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
mkdir -p ${RPM_BUILD_ROOT}/usr/bin
cp -pvrR bin/* ${RPM_BUILD_ROOT}/usr/bin | tee -a INSTALLED_FILES
cat INSTALLED_FILES | xargs -i sh -c 'test -L {} && exit || test -f $RPM_BUILD_ROOT/{} && echo {} || echo %dir {}' > INSTALLED_FILES_2

%clean

%files -f INSTALLED_FILES_2
%license LICENSE
%defattr(755,root,root)

%changelog
