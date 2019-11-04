if [ -n "$1" ] 
then
	cd /home/ddouglas/workspace/monocle
	go mod tidy
	go mod vendor	
	docker build . -t devoverlord93/monocle:$1
	docker push devoverlord93/monocle:$1
	rm -rf vendor
fi
