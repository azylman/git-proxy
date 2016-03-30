# git-proxy

**NOTE**: This tool is very beta. See Caveats below for a list of things that don't work, that haven't been tested, or assumptions it makes about your file system layout.

`git-proxy` proxies git commands from one machine to another. There's two pieces:
* `git-daemon`: a server that opens up an SSH connection to a remote machine and runs any received git commands there, proxying the results back
* `git-remote`: a tool that sends all of its arguments to the `git-daemon` server and logs the results to stdout and stderr

## Installing

``` shell
$ go get github.com/azylman/cmd/git-daemon
$ go get github.com/azylman/cmd/git-remote
```

## Usage

First, start the daemon:

``` shell
$ git-daemon --localhome /Users/azylman --remotehome /home/vagrant vm
```

Then, run any git commands using `git-remote`

``` shell
$ git version
git version 2.6.4 (Apple Git-63)
$ git-remote version
git version 2.5.0
```

You can also symlink `git` on your machine to `git-remote` so that everything goes to the daemon by default.

## Caveats

`git-proxy` doesn't work for commands the require interactive input (e.g. `commit` without `-m`).

The daemon assumes that it can listen on port `12345`. If that's not the case, it will fail.

I'm not sure how the daemon reacts if the connection to the SSH server is interrupted.

The daemon assumes that your git repositories are structured the same on your local and remote machines, and that the only difference is the directory prefix.
For example, imagine all of your repositories are in `$GOPATH` on the local and remote machine.
In that case, `localhome` would be `$GOPATH` on your local machine and `remotehome` would be `$GOPATH` on the remote machine
