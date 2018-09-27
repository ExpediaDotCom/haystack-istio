FROM scratch

ADD haystackadapter /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/haystackadapter"]
