# scm-go
implementation of scm in golang


api:

  * use('/fhgithub/trigger', this.trigger_handler) //clone and pull force clean done
  * use('/fhgithub/listfiles', this.listfiles_handler). //done
  * use('/fhgithub/create_tag', this.create_tag_handler). (dont think used)
  * use('/fhgithub/list_tags', this.list_tags_handler). //not used
  * use('/fhgithub/list_branches', this.list_branches_handler). //not used
  * use('/fhgithub/list_remote', this.list_remote_handler). done covers the above two
  * use('/fhgithub/zip', this.zip_handler). not used
  * use('/fhgithub/archive', this.archive_handler). done
  * use('/fhgithub/push', this.push_handler). not used
  * use('/fhgithub/mirror', this.mirror_handler). (used by openshift2. mirror does not seem to be an option in libgit2.)
  * use('/fhgithub/createfile', this.create_handler). done
  * use('/fhgithub/updatefile', this.update_handler). done
  * use('/fhgithub/deletefile', this.delete_handler). done
  * use('/fhgithub/getfile', this.get_handler).  done
  * use('/fhgithub/delete_app', this.delete_app_handler). done
  * use('/fhgithub/check_commit', this.check_commit_handler). done
  * use('/sys', this.sys_handler). ping/health done

##Local Development

### set up golang

Install golang [install](https://golang.org/doc/install)

as per install instructions set up the required env vars. The below env vars are what mine are set to

```bash

export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:/usr/local/go/bin
export GOPATH=/mnt/src/go
```

###setup scm-go
```bash
mkdir -p $GOPATH/src/github.com/maleck13
cd $GOPATH/src/github.com/maleck13
git clone git@github.com:maleck13/scm-go.git

#install godeps (manages dependency versions)
go get github.com/tools/godep
``` 

###libgit2

For local dev I recommend installing libgit2 for dynamic linking [https://libgit2.github.com/](github)
Mac osx I had to brew unlink libgit2 as it was an older version.

*OSX*
``` 
brew install pkg-config
brew install libssh2
brew install cmake

#install libgit2 on the mac. Useful for when running tests in intellij
cd $GOPATH/src/github.com/maleck13/scm-go
./scripts/build-libgit2-dynamic.sh
```

*Ubuntu*

```
sudo apt-get install cmake
sudo apt-get install libssh2-1-dev
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/mnt/src/go/src/github.com/maleck13/scm-go/vendor/libgit2/build/
cd $GOPATH/src/github.com/maleck13/scm-go
./scripts/build-libgit2-dynamic.sh

```

do a test build. This script builds libgit2.a so that we can static link it into scm-go
```
  ./scripts/build.sh clean 
```

run it 
```
  ./scm-go
```  

When testing with RHMAP I run it in the vm (note fhcap branch on the way)

```bash
#from inside vm
#change perms of existing repos as running as hadmin
sudo chmod -R 777 /opt/feedhenry/fh-scm/*

./scm-go /etc/feedhenry/fh-scm/conf.json

```

### IDE support
I use the golang plugin for intellij [golang plugin](https://plugins.jetbrains.com/plugin/5047)

### Test

you can run all tests from within intellij or from command line

Run all tests. From root dir of project:

```
 go test ./...
```


Run single test file

```
go test services/git_test.go 
```

Get coverage... Attempting to get to 70% + coverage as a baseline

```
go test -cover ./...
```

Get more verbose output

```
go test -v services/git_test.go 
```

#Adding new code

write your code. Write your tests. Then pre commit Run

```
go fmt ./...    formats the go code correctly
godep save -r ./...   ensures deps are from the godep saved dependencies

```


##Debugger

Intellij supports the dlv debugger. Just add break points and run with debugger. 

##Why golang

###Compiled
Compiled languages are good. They are generally faster, safer, they naturally catch a lot of what we unit test and lint for. The golang compiler
is very strict. It will not allow unused imports, unused vars type errors and much more. The go language compiles very quickly. Much faster than java, it reduces the overhead of a compiled language to almost 0.
you can simply run go run main.go to run it locally.

###Refactoring

//todo

###Asynchronous io
golang uses exactly the same underlying system calls as node for all network io (most of our io is network io). [Epoll](https://en.wikipedia.org/wiki/Epoll) and [kqueue](https://en.wikipedia.org/wiki/Kqueue). It is is just as asynchronous as node, except you avoid callbacks and can write your code in a sychronus  
form. Underneath golang uses very light weight go routines approx 2kB it then uses the go runtime scheduler to hide the intricacies of returning control to your function once the io is complete.


###Go Tools

Standard test tools ```go test```. Standard formatting ```go fmt``` . Race condition checking ``` go build -race ```. Built in documentation tool ```godoc```
 and much more.
 
###Always use the latest version.
 go builds an executable binary and packages the runtime in it. This means you can always use the latest version of go. Moving versions is up to us as engineers not whoever has to take on updating the runtime across all our servers.


### The standard lib

go compiles to machine code and is compiled by a compiler written in go. This means you can click on a function in go and be brought straight to the go regular go code and see how it was written.
This is extremely powerful. being able to easily see how the others of the language implemented the core of the language is really excellent.

### Simplicity
//todo javascript is not simple one of the main goals of golang is simplicity and a lack of magic
[good article for why golang will succeed](https://texlution.com/post/why-go-is-doomed-to-succeed/)


