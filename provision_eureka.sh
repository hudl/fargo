#!/usr/bin/env bash
sysctl -w net.ipv6.conf.all.disable_ipv6=1
systemctl stop firewalld.service
systemctl disable firewalld.service
#yum localinstall --nogpgcheck -y /vagrant/jdk-7u45-linux-x64.rpm
#yum install --nogpgcheck -y tomcat gradle git tomcat-admin-webapps tomcat-native jersey log4j htop vim
yum install --nogpgcheck -y tomcat gradle git tomcat-admin-webapps htop vim

echo "127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4" > /etc/hosts
echo "<?xml version='1.0' encoding='utf-8'?>
<tomcat-users>
  <user username=\"tomcatuser\" password=\"somep4ss\" roles=\"manager,admin,manager-gui,manager-status,manager-script,manager-jmx,admin-gui,admin-script\"/>
</tomcat-users>" > /etc/tomcat/tomcat-users.xml
chown tomcat:tomcat /etc/tomcat/tomcat-users.xml
chmod 644 /etc/tomcat/tomcat-users.xml

cd /vagrant/eureka
./gradlew clean build
cp ./eureka-server/build/libs/eureka-server-1.1.118.war /var/lib/tomcat/webapps/eureka.war
chmod a+x /var/lib/tomcat/webapps/eureka.war

#cp /vagrant/eureka-server-1.1.120-SNAPSHOT.war /usr/share/tomcat/webapps/eureka.war
#chown tomcat:tomcat /usr/share/tomcat/webapps/eureka.war
#chmod a+x /usr/share/tomcat/webapps/eureka.war

service tomcat restart

#echo "Done restarting tomcat. Go go go."
