DIR=$(dirname "$0")
cd ${DIR}
echo "Build test binary"
GOOS=linux go build -o testzomb
echo "Run test binary"
docker run -v $PWD:/testzomb python:3.7 /testzomb/testzomb /testzomb/stage_1.py