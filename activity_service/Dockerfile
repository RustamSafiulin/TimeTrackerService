
FROM centos:7

WORKDIR /app

ADD bin/activity_service /app/
ADD config.json /app/

EXPOSE 3000

RUN chmod +x ./activity_service

CMD ./activity_service
