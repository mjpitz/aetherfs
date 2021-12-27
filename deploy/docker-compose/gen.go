package main

// Generate ca.key and ca.crt
//go:generate openssl genrsa -out ca.key 4096
//go:generate openssl req -new -x509 -key ca.key -days 1 -out ca.crt -config ca.conf

// Generate tls.key and tls.csr
//go:generate openssl genrsa -out tls.key 4096
//go:generate openssl req -new -key tls.key -out tls.csr -config tls.conf

// Sign tls.csr using the ca and output to tls.crt
//go:generate openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -days 1 -out tls.crt -extensions req_ext -extfile tls.conf

func main() {}
