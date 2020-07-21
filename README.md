# edu
A command line interface for automating school interactions. This is a work in progress and it will be impossible to support the system that all schools use but this project will do its best to support the more popular systems (i.e. canvas).

I will not make any promises in terms of compatibility except that I will probably break things in the future.

Email me with any questions, harrybrown98@gmail.com.

## Installation
#### MacOS
```
brew install harrybrwn/tap/edu
```
#### Debian/Ubuntu
```
curl -LO https://github.com/harrybrwn/edu/releases/download/v0.0.3/edu_0.0.3_Linux_64-bit.deb
sudo dpkg -i edu_0.0.3_Linux_64-bit.deb
```
#### Rpm
```
curl -LO https://github.com/harrybrwn/edu/releases/download/v0.0.3/edu_0.0.3_Linux_64-bit.rpm
sudo rpm -i edu_0.0.3_Linux_64-bit.rpm
```
#### Windows
Download the zip file from the [releases page](https://github.com/harrybrwn/edu/releases) and good luck haha.
#### Compile from source
```sh
git clone https://github.com/harrybrwn/edu
cd edu
go install # or 'make install' if you want the correct version compiled into the binary
```

If your preferred method of installation is not supported you can always go to the releases page and download the zip or tar file.

## Canvas
To use any of the features that interact with canvas (the update and canvas commands), you need to [get an api token](https://community.canvaslms.com/docs/DOC-16005-42121018197) for your student account. For more info read the [configuration docs](/docs/config.md#token)

## Coniguration
See the [configuration docs](/docs/config.md). For config example, see [my example config](/docs/example_config.yml)