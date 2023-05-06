FROM ubuntu:20.04 as build-base
RUN echo '\
deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse \n\
deb-src http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse \n\
deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse \n\
deb-src http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse \n\
deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse      \n\
deb-src http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse  \n\
deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse     \n\
deb-src http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse \n\
deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse    \n\
deb-src http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse\n\
' > /etc/apt/sources.list && apt update

FROM  build-base as tools-installer
COPY ./get-pip.py /
RUN export DEBIAN_FRONTEND=noninteractive && apt install -y iputils-ping build-essential \
    libncurses5-dev gnupg wget iproute2 inetutils-tools python telnet curl tzdata vim\
    && ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && apt install -y bind9-dnsutils \
    && python /get-pip.py && rm /get-pip.py && pip install pika==1.2.1 amqp enum

FROM tools-installer as deployed
COPY ./producer.py ./consumer.py /
CMD ["/bin/bash"]
