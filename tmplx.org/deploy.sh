#/bin/sh

env GOOS=linux GOARCH=arm64 go build .
ssh tmplx.org "rm tmplx.org"
scp -r -i ~/.ssh/tmplx.org.pem tmplx.org assets ec2-user@ec2-3-92-67-184.compute-1.amazonaws.com:~
ssh tmplx.org "sudo pkill tmplx.org"
ssh tmplx.org "sudo CERT=/etc/letsencrypt/live/tmplx.org/fullchain.pem PK=/etc/letsencrypt/live/tmplx.org/privkey.pem ENV=prod nohup ./tmplx.org &>/dev/null &"
rm tmplx.org
