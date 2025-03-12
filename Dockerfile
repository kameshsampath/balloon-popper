FROM scratch
COPY balloon-popper /

ENTRYPOINT ["/balloon-popper"]
CMD ["--help"]
