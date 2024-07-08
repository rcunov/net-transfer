# net-transfer
Simple client/server file transfer application

## Setup
### Server
Generate an SSL cert with `openssl req -new -nodes -x509 -out server.pem -keyout server.key -days 365` within the `server/` folder

## OpenSSL on Windows
Git for Windows comes with openssl installed - check `C:\Program Files\Git\usr\bin\openssl.exe`. Installing that the regular way looks tedious.