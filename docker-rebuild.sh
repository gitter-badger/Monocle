mkdir -p /home/monocle/src/config
cp /home/monocle/config/env.json /home/monocle/src/config
docker build . -t monocle:latest
rm -rf /home/monocle/src/config
docker-compose up -d