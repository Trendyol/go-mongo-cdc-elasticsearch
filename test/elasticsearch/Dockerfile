FROM docker.elastic.co/elasticsearch/elasticsearch:7.17.10

COPY config/elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml

EXPOSE 9200 9300

ENV discovery.type=single-node
ENV ES_JAVA_OPTS="-Xms512m -Xmx512m"
ENV xpack.security.enabled=false
ENV cluster.name=test-cluster
ENV node.name=test-node

USER elasticsearch

