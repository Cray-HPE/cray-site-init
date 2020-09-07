# Copyright 2020 Cray Inc. All Rights Reserved.
Name: cray-metal-sic
License: Cray Software License Agreement
Summary: Shasta Install Control
Version: %(cat .version)
Release: %(echo ${BUILD_METADATA})
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Requires: podman

%description
This RPM installs the Docker control

%prep
%setup -q

%build

%install
install -m 755 -d %{buildroot}/root/bin
cp -pv init/sic-init.sh %{buildroot}/root/bin/
install -m 755 -d %{buildroot}/usr/lib/systemd/system
cp -pv init/sic.service %{buildroot}/usr/lib/systemd/system/

%clean

%files
%license LICENSE
%defattr(-,root,root)
/root/bin/sic-init.sh
/usr/lib/systemd/system/sic.service
