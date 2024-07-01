# MIT License
#
# (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
%global shortname csi

Name: %(echo $NAME)
License: MIT License
BuildArch: %(echo $ARCH)
Summary: HPCaaS configuration and deployment tool for bare-metal or reinstallations
Version: %(echo ${VERSION})
Release: 1
Source: %{name}-%{version}.tar.bz2
Vendor: Cray Inc.
Provides: %{shortname}

%ifarch %ix86
    %global GOARCH 386
%endif
%ifarch aarch64
    %global GOARCH arm64
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

go version

make %{shortname}

%install
CGO_ENABLED=0
GOOS=linux
GOARCH="%{GOARCH}"
GO111MODULE=on
export CGO_ENABLED GOOS GOARCH GO111MODULE

mkdir -pv ${RPM_BUILD_ROOT}/usr/bin/
cp -pv %{shortname} ${RPM_BUILD_ROOT}/usr/bin/%{shortname}

mkdir -pv ${RPM_BUILD_ROOT}/usr/local/bin/
cp -pv scripts/write-livecd.sh ${RPM_BUILD_ROOT}/usr/local/bin/write-livecd.sh

# Shell completion
install -d %{buildroot}%{_sysconfdir}/bash_completion.d
install -d %{buildroot}%{_sysconfdir}/zsh_completion.d
./%{shortname} completion bash >%{buildroot}%{_sysconfdir}/bash_completion.d/%{shortname}_complete
./%{shortname} completion zsh >%{buildroot}%{_sysconfdir}/zsh_completion.d/%{shortname}_complete

%clean

%files
%doc README.adoc
%license LICENSE
%defattr(755,root,root)
%{_sysconfdir}/bash_completion.d/%{name}_complete
%{_sysconfdir}/zsh_completion.d/%{name}_complete
/usr/bin/%{shortname}
/usr/local/bin/write-livecd.sh

%changelog
