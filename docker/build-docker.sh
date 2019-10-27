if [ -n "$1" ] 
then
	cd /home/ddouglas/workspace/monocle/cmd/cli
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /home/ddouglas/workspace/monocle/docker/cli
	cd /home/ddouglas/workspace/monocle/docker
	docker build . -t devoverlord93/monocle:$1
	docker push devoverlord93/monocle:$1
fi
