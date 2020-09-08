# Gorush container

# Stage1: download from github
FROM quay.io/cybozu/golang:1.13-bionic AS plugin-source

ARG CLASSIC_CJK_ANALYSIS_PLUGIN_VERSION="7.9.1"

WORKDIR /work
RUN git clone https://github.com/cybozu/classic-cjk-analysis-plugin.git -b "${CLASSIC_CJK_ANALYSIS_PLUGIN_VERSION}" . 

# Stage2: build plugin
FROM amazoncorretto:11.0.8-alpine AS plugin-build
COPY --from=plugin-source /work /work
RUN cd /work && ./gradlew clean zipPlugin

# Stage3: custom install 
FROM docker.elastic.co/elasticsearch/elasticsearch:7.9.1

COPY --from=plugin-build /work/build/dist/classic-cjk-analysis-plugin-0.0.1-es-7.9.1.zip /tmp/classic-cjk-analysis-plugin-0.0.1-es-7.9.1.zip 

RUN bin/elasticsearch-plugin install --batch analysis-icu && \
    bin/elasticsearch-plugin install --batch file:///tmp/classic-cjk-analysis-plugin-0.0.1-es-7.9.1.zip 
