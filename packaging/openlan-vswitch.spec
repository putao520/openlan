Name: openlan-vswitch
Version: 5.1.2
Release: 1%{?dist}
Summary: OpenLan's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go
BuildRequires: go
Requires: net-tools, iptables, iputils

%define _venv /opt/openlan-utils/env
%define _source_dir ${RPM_SOURCE_DIR}/openlan-go-%{version}

%description
OpenLan's Project Software

%build
cd %_source_dir && make linux/vswitch

virtualenv %_venv
%_venv/bin/pip install --upgrade "%_source_dir/py"

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/build/openlan-vswitch %{buildroot}/usr/bin

mkdir -p %{buildroot}/etc/openlan/vswitch
cp %_source_dir/packaging/resource/ctrl.json.example %{buildroot}/etc/openlan/vswitch
cp %_source_dir/packaging/resource/vswitch.json.example %{buildroot}/etc/openlan/vswitch
mkdir -p %{buildroot}/etc/sysconfig/openlan
cp %_source_dir/packaging/resource/vswitch.cfg %{buildroot}/etc/sysconfig/openlan

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/openlan-vswitch.service %{buildroot}/usr/lib/systemd/system

mkdir -p %{buildroot}/var/openlan
cp -R %_source_dir/packaging/resource/ca %{buildroot}/var/openlan
cp -R %_source_dir/vswitch/public %{buildroot}/var/openlan

mkdir -p %{buildroot}/opt/openlan-utils
cp -R /opt/openlan-utils/env %{buildroot}/opt/openlan-utils

%pre
firewall-cmd --permanent --zone=public --add-port=10000/tcp --permanent || {
  echo "You need allowed TCP port 10000 manually."
}
firewall-cmd --permanent --zone=public --add-port=10002/tcp --permanent || {
  echo "You need allowed TCP port 10000 manually."
}
firewall-cmd --reload || :

%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan
/opt/openlan-utils/*

%clean
rm -rf %_env
