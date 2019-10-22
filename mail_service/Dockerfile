FROM centos:7

WORKDIR /app

ADD bin/mail_service /app

EXPOSE 3001

RUN chmod +x ./mail_service

CMD ./mail_service
