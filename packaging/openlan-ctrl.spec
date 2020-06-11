Name: openlan-ctrl
Version: 5.2.7
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
cd %_source_dir && make linux/ctrl

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/build/openlan-ctrl %{buildroot}/usr/bin

mkdir -p %{buildroot}/etc/sysconfig/openlan
cp %_source_dir/packaging/resource/ctrl.cfg %{buildroot}/etc/sysconfig/openlan
mkdir -p %{buildroot}/var/openlan/ctrl
cp -R %_source_dir/packaging/resource/ca %{buildroot}/var/openlan/ctrl

mkdir -p %{buildroot}/etc/openlan/ctrl
cp %_source_dir/packaging/resource/auth.json.example %{buildroot}/etc/openlan/ctrl
mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/packaging/resource/openlan-ctrl.service %{buildroot}/usr/lib/systemd/system

%pre
firewall-cmd --permanent --zone=public --add-port=10088/tcp || {
  echo "You need allowed TCP port 10088 manually."
}
firewall-cmd --reload || :

%files
%defattr(-,root,root)
/etc/sysconfig/*
/etc/openlan/*
/usr/bin/*
/usr/lib/systemd/system/*
/var/openlan/*
