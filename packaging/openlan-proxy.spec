Name: openlan-proxy
Version: 5.5.14
Release: 1%{?dist}
Summary: OpenLAN's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan
BuildRequires: go

%define _source_dir ${RPM_SOURCE_DIR}/openlan-%{version}

%description
OpenLAN's Project Software

%build
cd %_source_dir && make linux-proxy

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/build/openlan-proxy %{buildroot}/usr/bin

mkdir -p %{buildroot}/etc/openlan
cp %_source_dir/packaging/resource/proxy.json.example %{buildroot}/etc/openlan

mkdir -p %{buildroot}/etc/sysconfig/openlan
cp %_source_dir/packaging/resource/proxy.cfg %{buildroot}/etc/sysconfig/openlan

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/openlan-proxy.service %{buildroot}/usr/lib/systemd/system

%pre

%post


%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/usr/bin/*
/usr/lib/systemd/system/*

%clean
rm -rf %_env
