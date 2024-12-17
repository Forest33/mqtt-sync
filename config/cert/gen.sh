#!/bin/bash

rm *.pem

# Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 3333 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=UZ/CN=*.boykevich.ru/emailAddress=anton@boykevich.ru"

# Generate server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=UZ/CN=*.boykevich.ru/emailAddress=anton@boykevich.ru"

# Use CA's private key to sign server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 3333 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

# Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=UZ/CN=*.boykevich.ru/emailAddress=anton@boykevich.ru"

# Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -req -in client-req.pem -days 3333 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

