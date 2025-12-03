# ==========================================
# Estágio 1: O Construtor (Builder)
# ==========================================
FROM debian:bookworm AS builder

# 1. Instalar Toolchain Básica
RUN apt-get update && apt-get install -y \
    build-essential cmake git curl libssl-dev pkg-config clang \
    wget

# 2. Instalar Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# 3. Instalar Go (1.24)
RUN wget https://go.dev/dl/go1.24.11.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.11.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# 4. Compilar liboqs (A Base)
WORKDIR /tmp/liboqs
RUN git clone --branch main https://github.com/open-quantum-safe/liboqs.git . && \
    mkdir build && cd build && \
    cmake .. -DOQS_ALGS_ENABLED=STD -DCMAKE_INSTALL_PREFIX=/usr/local -DBUILD_SHARED_LIBS=OFF && \
    make -j$(nproc) && \
    make install

# 5. Copiar Código Fonte do Projeto
WORKDIR /app
COPY . .

# 6. Compilar Core Rust (Static Lib)
WORKDIR /app/rust-core
RUN cargo clean && cargo build --release

# 7. Compilar Binário Go
WORKDIR /app
# CGO_ENABLED=1 é obrigatório. As flags -ldflags="-s -w" removem símbolos de debug para diminuir o tamanho.
RUN go mod tidy
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o pqc-proxy cmd/proxy/main.go

# ==========================================
# Estágio 2: A Imagem Final (Runtime)
# ==========================================
FROM debian:bookworm-slim

# Instalar apenas dependências de runtime (OpenSSL é necessário para HTTPS do SaaS)
RUN apt-get update && apt-get install -y ca-certificates openssl && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copiar apenas o binário compilado do estágio anterior
COPY --from=builder /app/pqc-proxy .

# Expor a porta do túnel
EXPOSE 4433

# Comando de entrada padrão (Modo Server)
ENTRYPOINT ["./pqc-proxy"]
CMD ["-mode=server", "-cloud=https://api.pqc-shield.com"]