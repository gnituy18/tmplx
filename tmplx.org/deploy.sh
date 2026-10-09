#/bin/sh

env GOOS=linux GOARCH=arm64 go build .
ssh tmplx.org "rm tmplx.org"
scp -r -i ~/.ssh/tmplx.org tmplx.org assets ec2-user@ec2-3-92-67-184.compute-1.amazonaws.com:~
ssh tmplx.org "sudo systemctl restart tmplx.org"
rm tmplx.org
