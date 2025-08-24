# Use Debian 13 (trixie) as base image
FROM debian:trixie

# Make Go version configurable
ARG GO_VERSION=1.25.0


# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV GOPATH=/go
ENV GOROOT=/usr/local/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

RUN apt-get update && apt-get install wget -y \
	&& mkdir -p -m 755 /etc/apt/keyrings \
	&& out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
	&& cat $out | tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
	&& chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
	&& mkdir -p -m 755 /etc/apt/sources.list.d \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update

# Install basic packages and dependencies first
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    sudo \
    vim \
    nano \
    unzip \
    tar \
    tree \
    zsh \
    just \
    gh \
    && rm -rf /var/lib/apt/lists/*



# Create developer user with sudo privileges
RUN groupadd --gid 1000 developer \
    && useradd --uid 1000 --gid developer --shell /bin/bash --create-home developer \
    && mkdir -p /etc/sudoers.d \
    && echo developer ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/developer \
    && chmod 0440 /etc/sudoers.d/developer

# Install Go
RUN wget -O go.tar.gz "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz \
    && chmod +x /usr/local/go/bin/go

# Install Oracle JDK 21
RUN wget -O- https://apt.corretto.aws/corretto.key | gpg --dearmor -o /etc/apt/trusted.gpg.d/corretto.gpg && \
    echo "deb https://apt.corretto.aws stable main" | tee /etc/apt/sources.list.d/corretto.list && \
    apt-get update && \
    apt-get install -y java-21-amazon-corretto-jdk && \
    rm -rf /var/lib/apt/lists/*

# Set JAVA_HOME environment variable
ENV JAVA_HOME=/usr/lib/jvm/java-21-amazon-corretto
ENV PATH=$JAVA_HOME/bin:$PATH

# Create Go directories and set proper permissions
RUN mkdir -p /go/pkg/mod && \
    mkdir -p /go/bin && \
    chown -R developer:developer /go && \
    chmod -R 755 /go

# Change default shell for developer user
RUN chsh -s /bin/zsh developer


# Switch to developer user
USER developer

# Install Go tools as developer user
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# Install Rust for developer user
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
ENV PATH="/home/developer/.cargo/bin:$PATH"

# Install oh-my-zsh for developer user
RUN sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# Configure zsh with plugins and environment variables
RUN echo 'export GOROOT=/usr/local/go' >> ~/.zshrc \
    && echo 'export GOPATH=/go' >> ~/.zshrc \
    && echo 'export JAVA_HOME=/usr/lib/jvm/java-21-amazon-corretto' >> ~/.zshrc \
    && echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$JAVA_HOME/bin:$HOME/.cargo/bin:$PATH' >> ~/.zshrc \
    && sed -i 's/ZSH_THEME="robbyrussell"/ZSH_THEME="agnoster"/' ~/.zshrc \
    && sed -i 's/plugins=(git)/plugins=(git golang rust)/' ~/.zshrc

# Set working directory
WORKDIR /workspace

# Set default shell
SHELL ["/bin/zsh", "-c"]
