%define app_home /opt/sysinner/indns
%define app_user root

Name: indns
Version: __version__
Release: __release__%{?dist}
Vendor:  sysinner.com
Summary: sysinner DNS server
License: Apache 2

Source0: %{name}-__version__.tar.gz
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}

Requires: redhat-lsb-core
Requires(pre): shadow-utils

%description

%prep
%setup -q -n %{name}-%{version}
%build

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{app_home}/bin
mkdir -p %{buildroot}%{app_home}/etc/conf.d
mkdir -p %{buildroot}%{app_home}/var/log
mkdir -p %{buildroot}%{app_home}/misc
mkdir -p %{buildroot}/lib/systemd/system

cp -rp misc/* %{buildroot}%{app_home}/misc/
install -m 755 bin/indnsd %{buildroot}%{app_home}/bin/indnsd
install -m 600 misc/systemd/systemd.service %{buildroot}/lib/systemd/system/indns.service

%clean
rm -rf %{buildroot}

%pre

%post
systemctl daemon-reload

%preun

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}
/lib/systemd/system/indns.service

%{app_home}/

