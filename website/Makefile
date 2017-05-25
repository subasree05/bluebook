serve:
	docker run --rm -ti -v $(PWD):/site -w /site -p 9000:9000 bluebook_site harp server --ip=0.0.0.0

setup:
	docker pull snakeego/harpjs

docker-build:
	docker build -t bluebook_site .
