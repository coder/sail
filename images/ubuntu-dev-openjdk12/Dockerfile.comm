# Taken From:
# https://github.com/docker-library/openjdk/blob/master/12/jdk/oracle/Dockerfile
#
# Modifications:
# Changed base image from oraclelinux:7-slim to be compatible with `buildfrom.sh`.
# This means things are going to have to be changed from a redhat linux base to a
# debian base, so some package installs may be changed.
# 
# Changed yum install to be apt based.
# Changed rm -rf yum cache to remove apt cache instead.
# Changed freetyp yum install to be libfreetype6.
# Removed oraclelinux comment about c.UTF8.
# Adjusted rhel java command `alternatives` to debian based `update-alternatives`.
# Removed CMD ["jshell"] and comments at the end.

FROM %BASE

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        gzip \
        tar \
    \
# java.lang.UnsatisfiedLinkError: /usr/java/openjdk-12/lib/libfontmanager.so: libfreetype.so.6: cannot open shared object file: No such file or directory
# https://github.com/docker-library/openjdk/pull/235#issuecomment-424466077
		libfreetype6 fontconfig \
	; \
	rm -rf /var/lib/apt/lists/*

# Default to UTF-8 file.encoding
#ENV LANG C.UTF-8

ENV JAVA_HOME /usr/java/openjdk-12
ENV PATH $JAVA_HOME/bin:$PATH

# https://jdk.java.net/
ENV JAVA_VERSION 12.0.1
ENV JAVA_URL https://download.java.net/java/GA/jdk12.0.1/69cfe15208a647278a19ef0990eea691/12/GPL/openjdk-12.0.1_linux-x64_bin.tar.gz
ENV JAVA_SHA256 151eb4ec00f82e5e951126f572dc9116104c884d97f91be14ec11e85fc2dd626

RUN set -eux; \
	\
	curl -fL -o /openjdk.tgz "$JAVA_URL"; \
	echo "$JAVA_SHA256 */openjdk.tgz" | sha256sum -c -; \
	mkdir -p "$JAVA_HOME"; \
	tar --extract --file /openjdk.tgz --directory "$JAVA_HOME" --strip-components 1; \
	rm /openjdk.tgz; \
	\
# https://github.com/oracle/docker-images/blob/a56e0d1ed968ff669d2e2ba8a1483d0f3acc80c0/OracleJava/java-8/Dockerfile#L17-L19
	ln -sfT "$JAVA_HOME" /usr/java/default; \
	ln -sfT "$JAVA_HOME" /usr/java/latest; \
	for bin in "$JAVA_HOME/bin/"*; do \
		base="$(basename "$bin")"; \
		[ ! -e "/usr/bin/$base" ]; \
		update-alternatives --install "/usr/bin/$base" "$base" "$bin" 20000; \
	done; \
	\
# https://github.com/docker-library/openjdk/issues/212#issuecomment-420979840
# https://openjdk.java.net/jeps/341
	java -Xshare:dump; \
	\
# basic smoke test
	java --version; \
	javac --version

