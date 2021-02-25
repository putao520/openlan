Name: openlan-switch
Version: 5.6.0
Release: 1%{?dist}
Summary: OpenLAN's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan
BuildRequires: go
Requires: net-tools, iptables, iputils, openvpn, openssl

%define _venv /opt/openlan-utils/env
%define _source_dir ${RPM_SOURCE_DIR}/openlan-%{version}

%description
OpenLAN's Project Software

%build
cd %_source_dir && make linux-switch

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/build/openlan %{buildroot}/usr/bin
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
cp -R %_source_dir/packaging/script %{buildroot}/var/openlan
cp -R %_source_dir/src/olsw/public %{buildroot}/var/openlan
mkdir -p %{buildroot}/var/openlan/cert
cp -R %_source_dir/build/cert/openlan/cert %{buildroot}/var/openlan
cp -R %_source_dir/build/cert/openlan/ca/ca.crt %{buildroot}/var/openlan/cert

mkdir -p %{buildroot}/var/openlan/openvpn
cp -R %_source_dir/packaging/resource/openvpn.md %{buildroot}/var/openlan/openvpn
cp -R %_source_dir/packaging/resource/example.ovpn %{buildroot}/var/openlan/openvpn

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
    /usr/bin/cp -rvf /etc/openlan/switch/switch.json.example /etc/openlan/switch/switch.json
fi
if [ ! -e "/var/openlan/openvpn/dh.pem" ]; then
    /usr/bin/openssl dhparam -out /var/openlan/openvpn/dh.pem 1024
fi
if [ ! -e "/var/openlan/openvpn/ta.key" ]; then
    /usr/sbin/openvpn --genkey --secret /var/openlan/openvpn/ta.key
fi


%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/etc/sysctl.d/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan/*

%clean
rm -rf %_env
