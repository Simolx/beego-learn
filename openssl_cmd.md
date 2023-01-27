## generate cert
```bash
openssl req -new -x509 -keyout ca.key -out ca.crt -days 3650

# server private key PKCS#8
openssl genpkey -out server1.key -algorithm RSA -pkeyopt rsa_keygen_bits:4096
# openssl rsa -in server1.key -out server1.key
openssl req -new -key server1.key -out server1.csr
openssl x509 -req -in server1.csr -CA ca.crt -CAkey ca.key -out server1.crt -days 3650 -CAcreateserial
openssl rsa -aes256 -in server1.key -out server1Encrypted.key
```

## check matches
```bash
# crt
openssl x509 -noout -modulus -in ca.crt | openssl md5
# csr
openssl req -noout -modulus -in CSR.csr | openssl md5
# key
openssl rsa -noout -modulus -in ca.key | openssl md5
```

## PEM to JKS
```bash
openssl pkcs12 -export -out service1.pk12 -in service1.crt -inkey service1.key
# java -cp jetty-6.1.26.jar org.mortbay.jetty.security.PKCS12Import service1.p12 keystore.jks
keytool -importkeystore -srckeystore service1.p12 -srcstoretype pkcs12 -destkeystore keystore.jks -deststoretype pkcs12
keytool -import -file ca.crt -keystore truststore.jks
```

## CentOS 8 config
```bash
sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
```

## Kafka Config

https://support.huaweicloud.com/usermanual-kafka/kafka-ug-0010.html

## SSL Config use KeyStore
### server.properties
```conf
# listeners=PLAINTEXT://KafkaService1:9092,SSL://KafkaService1:9093
listeners=SSL://KafkaService1:9093
# advertised.listeners=SSL://KafkaService1:9093
inter.broker.listener.name=SSL
ssl.endpoint.identification.algorithm=
ssl.client.auth=required
ssl.keystore.location=server.keystore.jks
ssl.keystore.password=123456
ssl.key.password=123456
ssl.truststore.location=server.truststore.jks
ssl.struststore.password=123456
```

### producer.properties
```conf
security.protocol=SSL
ssl.truststore.location=client.struststore.jks
ssl.truststroe.password=123456
# ssl.endpoint.identification.algorithm=
ssl.keystore.location=client.keystore.jks
ssl.keystore.password=123456
ssl.key.password=123456
```

https://codingharbour.com/apache-kafka/using-pem-certificates-with-apache-kafka/

## SSL Config use PEM
### server.properties
```conf
listeners=SSL://KafkaService1:9093
inter.broker.listener.name=SSL
ssl.client.auth=required
# security.protocol=SSL
ssl.keystore.type=PEM
# ssl.key.password=123456
ssl.keystore.certificate.chain=
ssl.keystore.key=
ssl.truststore.type=PEM
ssl.truststore.certificates=
```

### producer.properties
```conf
security.protocol=SSL
ssl.keystore.type=PEM
ssl.keystore.certificate.chain=
ssl.keystore.key=
ssl.truststore.type=PEM
ssl.truststore.certificates=
```