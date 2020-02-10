Name: openlan-point
Version: 4.2
Release: 1%{?dist}
Summary: OpenLan's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go

BuildRequires: go
Requires: iproute

%define _source_dir ${RPM_SOURCE_DIR}/openlan-%{version}

%description
OpenLan's Project Point Software

%build
cd %_source_dir
go build -mod=vendor -o ./resource/point.linux.x86_64 main/point_linux.go

%install

mkdir -p %{buildroot}/usr/bin
cp %_source_dir/resource/point.linux.x86_64 %{buildroot}/usr/bin/point

mkdir -p %{buildroot}/etc/sysconfig
cp %_source_dir/resource/point.cfg %{buildroot}/etc/sysconfig
mkdir -p %{buildroot}/etc/point
cp %_source_dir/resource/point.json %{buildroot}/etc/point/point.json.example

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/resource/point.service %{buildroot}/usr/lib/systemd/system

%pre


%files
%defattr(-,root,root)
/etc/*
/usr/bin/*
/usr/lib/systemd/system/*

%clean
