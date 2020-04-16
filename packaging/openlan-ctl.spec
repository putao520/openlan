Name: openlan-ctl
Version: 4.3.16
Release: 1%{?dist}
Summary: OpenLan's Controller Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go

BuildRequires: go

%define _source_dir ${RPM_SOURCE_DIR}/openlan-go-%{version}

%description
OpenLan's Project Software

%build
cd %_source_dir && make linux/ctl

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/controller/controller.linux.x86_64 %{buildroot}/usr/bin/ol-ctl

mkdir -p %{buildroot}/etc/sysconfig
cp %_source_dir/packaging/resource/ctl.cfg %{buildroot}/etc/sysconfig/ol-ctl.cfg
mkdir -p %{buildroot}/var/openlan/ctl
cp -R %_source_dir/packaging/resource/ca %{buildroot}/var/openlan/ctl

mkdir -p %{buildroot}/etc/openlan/ctl
mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/ctl.service %{buildroot}/usr/lib/systemd/system/ol-ctl.service

%pre
firewall-cmd --permanent --zone=public --add-port=10088/tcp || {
  echo "You need allowed TCP port 10088 manually."
}
firewall-cmd --reload || :

%files
%defattr(-,root,root)
/etc/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan/*
