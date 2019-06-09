# Based Upon:
# https://github.com/d11wtq/llvm-docker
#
# Modifications:
#
# - Use LLVM 8 instead of LLVM 3.9.
# - Change the signing key URL.
# - Merge `apt-get install` steps into the prior `apt-get update` step.
# - Check for file already existing when creating symlinks.

FROM %BASE

RUN apt-get update -qq -y && \
    apt-get install -qq -y wget

# Ubuntu Cosmic LLVM APT repository: http://apt.llvm.org
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | sudo apt-key add -
ADD llvm-8.list /etc/apt/sources.list.d/llvm-8.list

RUN apt-get update -qq -y && \
    apt-get install -qq -y \
    make \
    clang-8 \
    clang-8-doc \
    clang-format-8 \
    clang-tools-8 \
    libc++-8-dev \
    libc++abi-8-dev \
    libclang-8-dev \
    libclang-common-8-dev \
    libclang1-8 \
    libfuzzer-8-dev \
    libllvm-8-ocaml-dev \
    libllvm8 \
    libomp-8-dev \
    lld-8 \
    lldb-8 \
    llvm-8 \
    llvm-8-dev \
    llvm-8-doc \
    llvm-8-examples \
    llvm-8-runtime \
    llvm-8-tools \
    python-clang-8

RUN for f in $(find /usr/bin -name '*-8'); do \
      newname=`echo $f | sed s/-8//`; \
      [ ! -f $newname ] && ln -s $f $newname || true; \
    done
