FROM golang:1.21

ARG SOURCE_MIRROR

RUN if [ -n "$SOURCE_MIRROR" ]; then \
        echo "Replacing source with $SOURCE_MIRROR"; \
        sed -i "s|http://deb.debian.org/debian|$SOURCE_MIRROR|g" /etc/apt/sources.list.d/debian.sources; \
    fi \
    && apt-get update -y \
    && apt-get install -y vim make wget curl lsof telnet git zip unzip sqlite3 rsync net-tools gcc g++ libvips libvips-dev tmux sudo zsh nodejs npm \
      darktable \
      dbus-x11 \
      libgl1-mesa-glx \
      libglib2.0-0 \
      libx11-6 \
      libxext6 \
      libxrender1 \
      libsm6 \
      libice6 \
      libxi6 \
      libxtst6 \
      libxt6 \
      libxmu6 \
      libxpm4 \
      libxft2 \
      libxinerama1 \
      libxxf86vm1 \
      libxrandr2 \
      --no-install-recommends \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g @vue/cli

ARG USER_ID
ARG GROUP_ID

RUN if ! getent group $GROUP_ID; then \
        groupadd -g $GROUP_ID dev; \
    fi && \
    useradd -m -u $USER_ID -g $GROUP_ID dev && \
    echo "dev ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER dev

ENV CGO_ENABLED='1'
ENV GO111MODULE=on
ENV GOPROXY='https://goproxy.io,direct'

WORKDIR /project

RUN sh -c "$(curl -fsSL https://gitee.com/mirrors/oh-my-zsh/raw/master/tools/install.sh)" "" --unattended \
    && go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /project/src

EXPOSE 8000 8001

CMD [ "zsh" ]
