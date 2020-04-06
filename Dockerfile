## docker run -it --rm --entrypoint=/bin/bash oracle/instantclient:12.2.0.1
#docker run -it --rm --entrypoint=/bin/bash fra.ocir.io/oraseemeatechse/eugenesimos/notifysql:0.0.25
#
# HOW TO BUILD THIS IMAGE
# -----------------------
# Put all downloaded files in the same directory as this Dockerfile
# Run: 
#      $ docker build --pull -t oracle/instantclient:12.2.0.1 .
#
#
FROM oraclelinux:7-slim  as  build-stage

WORKDIR  /tmp
 
ADD instantclient_12_2/* /tmp/lib/  
RUN yum -y install oracle-release-el7 && \
     yum -y install tar && \
	 yum -y install gzip && \
	 yum -y install git  && \ 
     yum -y install gcc  &&  \
	 rm -rf /var/cache/yum && \
     ls -la  /tmp/lib/* && \
	 mkdir -p /usr/lib/oracle/12.2/client64/lib && \
	 mv /tmp/lib/* /usr/lib/oracle/12.2/client64/lib && \
	 ls -la /usr/lib/oracle/12.2/client64/lib &&  \
     echo /usr/lib/oracle/12.2/client64/lib > /etc/ld.so.conf.d/oracle-instantclient12.2.conf && \
               ldconfig


ENV PATH $PATH:/usr/lib/oracle/12.2/client64/lib
ENV LD_LIBRARY_PATH /usr/lib/oracle/12.2/client64/lib

COPY go1.14.1.linux-amd64.tar.gz	 .

RUN  rm -rf  /tmp/go && \
     ls -la /tmp && \
	 tar -C /usr/local -xzf go1.14.1.linux-amd64.tar.gz && \
	 rm go1.14.1.linux-amd64.tar.gz &&  \
	 mkdir -p /function 
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH:/usr/lib/oracle/12.2/client64/lib
RUN mkdir -p $GOPATH/src  && \
    mkdir -p $GOPATH/bin && \
	chmod -R 777 $GOPATH && \
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh 
WORKDIR /function
ENV  LD_LIBRARY_PATH=/usr/lib/oracle/12.2/client64/lib
ENV PATH=/usr/sbin:/usr/bin:/usr/lib/oracle/12.2/client64/lib:/function:/go/bin:/usr/local/go/bin
ENV TNS_ADMIN=/function/wallet
RUN  groupadd --gid 1000 fn && \
     adduser --uid 1000 --gid fn fn 
COPY func.go func.yaml Gopkg.toml /go/src/func/
RUN cd /go/src/func/ && dep ensure &&  cd /go/src/func/ && go build -o func && mkdir -p /function 
FROM oraclelinux:7-slim
RUN  groupadd --gid 1000 fn && \
     adduser --uid 1000 --gid fn fn
WORKDIR /function
ENV  LD_LIBRARY_PATH=/usr/lib/oracle/12.2/client64/lib
ENV PATH=/usr/sbin:/usr/bin:/usr/lib/oracle/12.2/client64/lib:/function:/go/bin:/usr/local/go/bin
ENV TNS_ADMIN=/function/wallet
COPY --from=build-stage /usr/lib/oracle/12.2/client64/lib /usr/lib/oracle/12.2/client64/lib
COPY  --from=build-stage /go/src/func/func /function
COPY  test_sql.sh script.sh show_trace.sh /function/
COPY wallet_adw/*  /function/wallet/
## yum install -y strace && \
RUN  chmod -R ugo+rwx /function &&  \
	chown -R 1000:1000 /usr/lib/oracle/12.2/client64/lib  \
	&& ls -la /function && \
    sqlplus admin/WElcome1412#@adwfree_low && \
	  /function/test_sql.sh && \
	chmod ugo+rwx /usr/bin  && \
	cd /usr/bin && \
	rm -rf gpg2 dgawk gawk pgawk gpgv2 gpg-agent info diff find oldfind grep curl gpg-connect-agent gpgconf  gpgparsemail gpg-error mkdir
ENTRYPOINT ["./func"]


