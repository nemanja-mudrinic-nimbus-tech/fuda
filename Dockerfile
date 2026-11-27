FROM --platform=$BUILDPLATFORM node:24-alpine AS client
RUN corepack enable
WORKDIR /src
COPY --chown=node:node client/package.json client/pnpm-lock.yaml client/
RUN chown node:node /src
USER node
RUN cd client && pnpm install --frozen-lockfile
COPY --chown=node:node client client
RUN mkdir -p api/internal/web/dist && cd client && pnpm build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS api
ARG TARGETOS TARGETARCH
ARG VERSION
WORKDIR /src/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./
COPY --from=client /src/api/internal/web/dist ./internal/web/dist
RUN test -n "$VERSION" && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X fuda/internal/version.Fuda=$VERSION" -o /out/fuda ./cmd/fuda && mkdir -p /out/data

FROM gcr.io/distroless/static:nonroot
ARG VERSION
LABEL org.opencontainers.image.version=$VERSION
COPY --from=api /out/fuda /fuda
COPY --from=api --chown=nonroot:nonroot /out/data /data
USER nonroot:nonroot
ENV FUDA_CACHE_DIR=/data
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/fuda"]
