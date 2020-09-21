Name: openlan-switch
Version: 5.4.9
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
cd %_source_dir && make linux-switch

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/packaging/script/curl %{buildroot}/usr/bin/openlan-curl
cp %_source_dir/packaging/script/check %{buildroot}/usr/bin/openlan-check
cp %_source_dir/build/openlan-switch %{buildroot}/usr/bin

mkdir -p %{buildroot}/etc/openlan/switch
cp %_source_dir/packaging/resource/ctrl.json.example %{buildroot}/etc/openlan/switch
cp %_source_dir/packaging/resource/switch.json*.example %{buildroot}/etc/openlan/switch
mkdir -p %{buildroot}/etc/openlan/switch/network
cp %_source_dir/packaging/resource/network.json.example %{buildroot}/etc/openlan/switch/network
mkdir -p %{buildroot}/etc/sysconfig/openlan
cp %_source_dir/packaging/resource/switch.cfg %{buildroot}/etc/sysconfig/openlan
mkdir -p %{buildroot}/etc/sysctl.d
cp %_source_dir/packaging/resource/90-openlan.conf %{buildroot}/etc/sysctl.d

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/openlan-switch.service %{buildroot}/usr/lib/systemd/system

mkdir -p %{buildroot}/var/openlan
cp -R %_source_dir/packaging/resource/ca %{buildroot}/var/openlan
cp -R %_source_dir/packaging/script %{buildroot}/var/openlan
cp -R %_source_dir/src/olsw/public %{buildroot}/var/openlan

%pre
/usr/bin/firewall-cmd --permanent --zone=public --permanent \
 --add-port=10002/udp --add-port=10002/tcp || {
  echo "YOU NEED ALLOW TCP/UDP PORT:10002."
}
/usr/bin/firewall-cmd --permanent --zone=public --permanent \
  --add-port=10000/tcp --add-port=11080-11084/tcp || {
  echo "YOU NEED ALLOW TCP PORT:10000 and 11080-11084"
}
firewall-cmd --reload || :

%post
if [ ! -e "/etc/openlan/switch/switch.json" ]; then
    cp -rvf /etc/openlan/switch/switch.json.example /etc/openlan/switch/switch.json
fi

%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/etc/sysctl.d/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan

%clean
rm -rf %_env
