Name: openlan-vswitch
Version: 4.0
Release: 1%{?dist}
Summary: OpenLan's Project Software
Group: Applications/Communications
License: Apache 2.0
URL: https://github.com/danieldin95/openlan-go
BuildRequires:  go
Requires: net-tools, redis, openvswitch

%define _source_dir ${RPM_SOURCE_DIR}/openlan-%{version}

%description
OpenLan's Project Software

%build
cd %_source_dir
go build -o ./resource/vswitch.linux.x86_64 main/vswitch_linux.go

%install
mkdir -p %{buildroot}/usr/bin
cp %_source_dir/resource/vswitch.linux.x86_64 %{buildroot}/usr/bin/vswitch

mkdir -p %{buildroot}/etc
cp %_source_dir/resource/vswitch.json %{buildroot}/etc
cp %_source_dir/resource/vswitch.cfg %{buildroot}/etc

mkdir -p %{buildroot}/usr/lib/systemd/system
cp %_source_dir/resource/vswitch.service %{buildroot}/usr/lib/systemd/system

mkdir -p %{buildroot}/var/openlan
cp -R %_source_dir/resource/ca %{buildroot}/var/openlan
cp -R %_source_dir/public %{buildroot}/var/openlan

cat > %{buildroot}/etc/vswitch.password << EOF
hi:hi@123$
EOF

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

%clean
