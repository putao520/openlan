Name: openlan-point
Version: 5.4.2
Release: 1%{?dist}
Summary: OpenLan's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go

BuildRequires: go
Requires: iproute

%define _source_dir ${RPM_SOURCE_DIR}/openlan-go-%{version}

%description
OpenLan's Project Point Software

%build
cd %_source_dir && make linux-point

%install

mkdir -p %{buildroot}/usr/bin
cp %_source_dir/build/openlan-point %{buildroot}/usr/bin

mkdir -p %{buildroot}/etc/sysconfig/openlan
cp %_source_dir/packaging/resource/point.cfg %{buildroot}/etc/sysconfig/openlan
mkdir -p %{buildroot}/etc/openlan
cp %_source_dir/packaging/resource/point.json.example %{buildroot}/etc/openlan

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/openlan-point.service %{buildroot}/usr/lib/systemd/system

%pre


%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/usr/bin/*
/usr/lib/systemd/system/*

%clean
