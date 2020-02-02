Name: openlan-vswitch
Version: 4.1
Release: 1%{?dist}
Summary: OpenLan's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go
BuildRequires: go
Requires: net-tools

%define _venv /opt/openlan-utils/env
%define _source_dir ${RPM_SOURCE_DIR}/openlan-%{version}

%description
OpenLan's Project Software

%build
cd %_source_dir
go build -mod=vendor -o ./resource/vswitch.linux.x86_64 main/vswitch.go

virtualenv %_venv
%_venv/bin/pip install --upgrade "%_source_dir/py"

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/resource/vswitch.linux.x86_64 %{buildroot}/usr/bin/vswitch

mkdir -p %{buildroot}/etc/vswitch
cp %_source_dir/resource/vswitch.json %{buildroot}/etc/vswitch/vswitch.json.example
cp %_source_dir/resource/network.json %{buildroot}/etc/vswitch/network.json.example
mkdir -p %{buildroot}/etc/sysconfig
cp %_source_dir/resource/vswitch.cfg %{buildroot}/etc/sysconfig

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/resource/vswitch.service %{buildroot}/usr/lib/systemd/system

mkdir -p %{buildroot}/var/openlan
cp -R %_source_dir/resource/ca %{buildroot}/var/openlan
cp -R %_source_dir/public %{buildroot}/var/openlan

cat > %{buildroot}/etc/vswitch/vswitch.password.example << EOF
hi:hi@123$
EOF

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
/etc/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan
/opt/openlan-utils/*

%clean
rm -rf %_env
