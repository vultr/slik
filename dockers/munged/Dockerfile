FROM ubuntu:22.04

RUN apt update && apt upgrade -y && apt install ca-certificates git -y
RUN apt install munge -y

COPY . .

CMD ["/entrypoint.sh"]
