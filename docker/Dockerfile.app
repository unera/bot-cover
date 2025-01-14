FROM docker.uvw.ru:5000/unera/cover-bot-base

RUN mkdir config bin go profiles
RUN git clone https://github.com/unera/bot-cover.git src
RUN make -C src update_version build
RUN mv src/bot-cover bin/bot
RUN ln -s src/fonts .
RUN cp src/config.example.yaml config.yaml
ENTRYPOINT [ "bash" ]
ENTRYPOINT [ "bot", "config.yaml" ]
