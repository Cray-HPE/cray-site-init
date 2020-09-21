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

%install

%clean

%files
%license LICENSE
%defattr(-,root,root)
