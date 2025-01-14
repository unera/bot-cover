all:

BASE_TAGS := -t docker.uvw.ru:5000/unera/cover-bot-base \
	     -t 217.16.23.115:5000/unera/cover-bot-base

TAGS := -t docker.uvw.ru:5000/unera/cover-bot \
        -t 217.16.23.115:5000/unera/cover-bot



base_docker:
	docker build $(CACHE) $(BASE_TAGS) -f docker/Dockerfile.base .

upload_base_docker: base_docker
	docker push docker.uvw.ru:5000/unera/cover-bot-base

docker:
	docker build --no-cache $(TAGS) -f docker/Dockerfile.app .

fast-docker:
	docker build --no-cache $(TAGS) -f docker/Dockerfile.fast-app .

upload_docker: docker 
	docker push docker.uvw.ru:5000/unera/cover-bot

.PHONY: \
	docker \
	upload_docker \
	base_docker \
	upload_base_docker \
	fast-docker \
	all
