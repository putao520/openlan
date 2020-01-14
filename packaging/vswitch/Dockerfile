FROM centos:7

RUN mkdir /var/openlan
WORKDIR /var/openlan

RUN mkdir /var/openlan/ca
RUN mkdir /var/openlan/public
COPY public/* /var/openlan/public/
COPY resource/ca/* /var/openlan/ca/

COPY resource/vswitch.linux.x86_64 /var/openlan
COPY packaging/vswitch/entrypoint.sh /var/openlan

RUN mkdir /var/openlan/vswitch
RUN echo $'{\n\
  "bridger": "virtual",\n\
  "http.addr": ":10000",\n\
  "vs.addr": "0.0.0.0:10002",\n\
  "http.dir": "/var/openlan/public",\n\
  "log.file": "/var/openlan/vswitch.log",\n\
  "crt.dir": "/var/openlan/ca"\n\
}\n' > /var/openlan/vswitch/vswitch.json
RUN echo $'[\n\
  {\n\
    "tenant": "admin",\n\
    "ifAddr": "192.168.100.1",\n\
    "ipRange": 32,\n\
    "ipAddr": "192.168.100.200",\n\
    "netmask": "255.255.255.0"\n\
  }\n\
]\n' > /var/openlan/vswitch/network.json
RUN echo "hi@admin:hi123$" >> /var/openlan/vswitch/password

LABEL application="openlan-vswitch"
LABEL maintainer="danieldin95@163.com"

EXPOSE 10000/tcp
EXPOSE 10002/tcp

ENTRYPOINT ["/var/openlan/entrypoint.sh"]