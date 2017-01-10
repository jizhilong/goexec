#!/bin/sh

echo "# install goexec to this docker host:"
cat << EOF
docker run --rm -v /:/hostroot /goexec install
EOF

echo "# run goexec inside a docker container:"
cat << EOF
docker run -d -v /var/run/docker.sock:/var/run/docker.sock --net host jizhilong/goexec goexec -w -p 8080
EOF
