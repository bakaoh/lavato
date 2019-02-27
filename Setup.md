

## Install glibc 2.14

```
$ mkdir ~/glibc_install; cd ~/glibc_install 
$ wget http://ftp.gnu.org/gnu/glibc/glibc-2.14.tar.gz
$ tar zxvf glibc-2.14.tar.gz
$ cd glibc-2.14
$ mkdir build; cd build
$ ../configure --prefix=/opt/glibc-2.14
$ make -j4
$ sudo make install
$ export LD_LIBRARY_PATH=/opt/glibc-2.14/lib
```

## Install g++/gcc 4.8.2 in CentOS 6.6 

```
$ wget http://people.centos.org/tru/devtools-2/devtools-2.repo -O /etc/yum.repos.d/devtools-2.repo
$ yum install devtoolset-2-gcc devtoolset-2-binutils
$ yum install devtoolset-2-gcc-c++ devtoolset-2-gcc-gfortran
$ /opt/rh/devtoolset-2/root/usr/bin/gcc --version
$ scl enable devtoolset-2 bash
$ source /opt/rh/devtoolset-2/enable
```