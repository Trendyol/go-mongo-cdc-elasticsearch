FROM mongo:7.0.15

# Add configuration script
ADD configure.sh /configure.sh
RUN chmod +x /configure.sh

# Expose MongoDB ports
# 27017: mongos (router)
# 27018: shard server
# 27019: config server
EXPOSE 27017 27018 27019

CMD ["/configure.sh"]

