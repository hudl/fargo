#!/usr/bin/env bash
sysctl -w net.ipv6.conf.all.disable_ipv6=1
# this is for vbox guest additions v 4.3.10, may not be necessary on your machine http://stackoverflow.com/a/22723807
ln -s /opt/VBoxGuestAdditions-4.3.10/lib/VBoxGuestAdditions /usr/lib/VBoxGuestAdditions || true
systemctl stop firewalld.service
systemctl disable firewalld.service
yum install --nogpgcheck -y unzip tomcat tomcat-admin-webapps htop vim

echo "127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4" > /etc/hosts
echo "<?xml version='1.0' encoding='utf-8'?>
<tomcat-users>
  <user username=\"tomcatuser\" password=\"somep4ss\" roles=\"manager,admin,manager-gui,manager-status,manager-script,manager-jmx,admin-gui,admin-script\"/>
</tomcat-users>" > /etc/tomcat/tomcat-users.xml
chown tomcat:tomcat /etc/tomcat/tomcat-users.xml
chmod 644 /etc/tomcat/tomcat-users.xml

echo "127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4
172.16.0.11 node1 node1.localdomain
172.16.0.22 node2 node2.localdomain
" > /etc/hosts

ARTIFACT="https://netflixoss.ci.cloudbees.com/job/eureka-master/lastSuccessfulBuild/artifact/eureka-server/build/libs/eureka-server-1.1.147-SNAPSHOT.war"

wget -O /var/lib/tomcat/webapps/eureka.zip ${ARTIFACT}
unzip -o -d /var/lib/tomcat/webapps/eureka /var/lib/tomcat/webapps/eureka.zip
cp /vagrant/tests/eureka_properties/*.properties /var/lib/tomcat/webapps/eureka/WEB-INF/classes/
chown -R tomcat:tomcat /var/lib/tomcat/webapps/eureka
rm -f /var/lib/tomcat/webapps/eureka.zip

service tomcat restart
chkconfig tomcat on
