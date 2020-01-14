FROM centos:7

RUN mkdir /var/openlan
WORKDIR /var/openlan
COPY resource/point.linux.x86_64 /var/openlan
COPY packaging/point/entrypoint.sh /var/openlan

RUN echo $'{\n\
  "vs.addr": "openlan.net",\n\
  "vs.auth": "hi@admin:hi123$",\n\
  "vs.tls": true,\n\
  "log.file": "/var/openlan/point.log"\n\
}\n' > /var/openlan/point.json

LABEL application="openlan-point"
LABEL maintainer="danieldin95@163.com"

ENTRYPOINT ["/var/openlan/entrypoint.sh"]