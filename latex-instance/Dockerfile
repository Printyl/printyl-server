FROM ubuntu:latest
LABEL authors="Gregor Gottschewski"

RUN apt-get update && apt-get install -y \
    texlive-latex-base \
    texlive-latex-recommended \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create a dedicated low-privilege user for running LaTeX
RUN useradd -m -s /bin/bash latexuser

# Only the /jobs directory is writable
RUN mkdir /jobs && chown latexuser:latexuser /jobs

USER latexuser
WORKDIR /jobs

CMD ["sleep", "infinity"]
