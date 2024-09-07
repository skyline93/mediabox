FROM mediabox-dev:latest AS build

USER root

WORKDIR "/mediabox/src"
COPY . .

RUN make

######################### PRODUCTION STAGE ###############################
FROM mediabox-dev:latest

WORKDIR "/mediabox"
USER mediabox

COPY --from=build /mediabox/src/mediabox /usr/local/bin/mediabox

EXPOSE 8000 8001

CMD ["/usr/local/bin/mediabox", "start"]
