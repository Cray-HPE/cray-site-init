# Copyright 2020 Cray Inc. All Rights Reserved.
Name: cray-site-init
License: Cray Proprietary
Summary: Initialize and Upgrade Cray HPCaaS both bare-metal or in the wild
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Provides: csi
Provides: sic
Provides: shasta-instance-control
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

%pre
# Replace the old application with a symlink to the new application.
if [ /usr/bin/sic ] ; then
    rm /usr/bin/sic
    (cd /usr/bin/ && rm -f sic && ln -snf ./csi sic)
fi

%clean

%files
%license LICENSE
%defattr(755,root,root)
/usr/bin/csi

%changelog
