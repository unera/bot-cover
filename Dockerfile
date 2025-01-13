FROM debian:trixie

RUN apt-get -yq update
RUN apt-get -yq install \
	golang-go \
	git \
	libmagickwand-dev \
	tree \
	sudo \
	tzdata
ENV TZ=Europe/Moscow

WORKDIR /cover-bot
ENV GOPATH=/cover-bot/go
ENV PATH="/cover-bot/bin:${GOPATH}/bin:${PATH}"
RUN mkdir -p /cover-bot
RUN mkdir config bin go profiles
RUN git clone https://github.com/unera/bot-cover.git src
RUN cd src && go build
RUN mv src/bot-cover bin/bot
RUN ln -s src/fonts .
RUN cp src/config.example.yaml config.yaml
ENTRYPOINT bash
ENTRYPOINT bot config.yaml
