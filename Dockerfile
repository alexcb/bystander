FROM golang:1.9.2
MAINTAINER Alex Couture-Beil <bystander@mofo.ca>

ADD https://github.com/alexcb/bystander/archive/master.zip /root
RUN apt-get update && apt-get install unzip
WORKDIR /root/
RUN unzip /root/master.zip
RUN pwd
WORKDIR /root/bystander-master
RUN ls -la
RUN ./build.sh

FROM golang:1.9.2
MAINTAINER Alex Couture-Beil <bystander@mofo.ca>

ADD https://download.docker.com/linux/static/stable/x86_64/docker-17.09.0-ce.tgz /root
RUN tar -xOf /root/docker-17.09.0-ce.tgz docker/docker > /usr/bin/docker
RUN chmod +x /usr/bin/docker
RUN rm /root/docker-17.09.0-ce.tgz

COPY --from=0 /root/bystander-master/bystander /app/bystander
COPY --from=0 /root/bystander-master/static /app/static

WORKDIR /app

CMD ["/app/bystander"]

