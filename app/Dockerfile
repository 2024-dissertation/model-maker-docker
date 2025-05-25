# ------------------------------
# Stage 1: Build OpenMVG/OpenMVS
# ------------------------------
FROM --platform=linux/amd64 ubuntu:latest AS builder

ENV DEBIAN_FRONTEND=noninteractive TZ=Europe/Minsk

RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata cmake build-essential git wget curl \
    python3-dev libboost-all-dev libopencv-dev \
    libjpeg-dev libpng-dev libtiff-dev libglu1-mesa-dev \
    libglew-dev libglfw3-dev coinor-libclp-dev libceres-dev \
    libcgal-dev libcgal-qt5-dev graphviz liblemon-dev \
    ca-certificates pkg-config libxi6 \
    g++-x86-64-linux-gnu libc6-dev-amd64-cross \
    vim nano && \
    rm -rf /var/lib/apt/lists/* && update-ca-certificates

RUN git clone --depth=1 https://github.com/cdcseacave/VCG.git /vcglib

RUN git clone --recursive https://github.com/openMVG/openMVG.git /openMVG && \
    mkdir -p /openMVG_build && cd /openMVG_build && \
    cmake -DCMAKE_BUILD_TYPE=RELEASE /openMVG/src && \
    make -j$(nproc)

RUN git clone --branch develop https://github.com/cdcseacave/openMVS.git /openMVS && \
    sed -i 's|<CGAL/AABB_traits_3.h>|<CGAL/AABB_tree.h>|g' /openMVS/libs/MVS/SceneReconstruct.cpp && \
    sed -i 's|<CGAL/AABB_triangle_primitive_3.h>|<CGAL/AABB_triangle_primitive.h>|g' /openMVS/libs/MVS/SceneReconstruct.cpp && \
    mkdir -p /openMVS_build && cd /openMVS_build && \
    cmake -DCMAKE_BUILD_TYPE=RELEASE \
            -DOpenMVG_DIR=/openMVG_build/src/openMVG/cmake/ \
            -DVCG_ROOT=/vcglib \
            /openMVS && \
    make -j$(nproc)    

# -------------------------
# Stage 2: Build Go backend
# -------------------------
FROM --platform=linux/amd64 ubuntu:latest AS final

RUN apt-get update && apt-get install -y wget xz-utils libcgal-qt5-dev \
    libceres-dev libboost-all-dev libopencv-dev build-essential && \
    wget https://download.blender.org/release/Blender4.4/blender-4.4.0-linux-x64.tar.xz && \
    mkdir -p /opt/blender && \
    tar -xf blender-4.4.0-linux-x64.tar.xz -C /opt/blender --strip-components=1 && \
    ln -s /opt/blender/blender /usr/local/bin/blender && \
    rm blender-4.4.0-linux-x64.tar.xz && \
    wget https://go.dev/dl/go1.23.8.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.23.8.linux-amd64.tar.gz && \
    rm -rf go1.23.8.linux-amd64.tar.gz && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server
# RUN chmod +x ./server

COPY --from=builder /openMVG_build/Linux-x86_64-RELEASE /usr/local/bin
COPY --from=builder /openMVS_build/bin /usr/local/bin

ENV PATH="$PATH:/usr/local/go/bin:/root/go/bin"

RUN go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install github.com/joho/godotenv/cmd/godotenv@latest && \
    go install github.com/nikolaydubina/go-cover-treemap@latest

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /app

COPY . .

ARG GOOGLE_CREDENTIALS

RUN env

RUN echo "${GOOGLE_CREDENTIALS}" > service-account-key.json

EXPOSE 3333

CMD ["make"]
# CMD ["tail", "-f", "/dev/null"]