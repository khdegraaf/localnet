# DEMO

## Prepare
Open:
- ./env.go
- ./tests/index.go
- ./tests/transfers/transfer.go
- terminal

```
rm -rf ~/.localnet
```

## Defaults
```
$ localnet start
## ctrl-b d
$ localnet start  <-- reattach to the same session
```

Show source code of dev set

### Environment console
```
$ localnet
```

### Shortcuts
```
start
```

### Logs
```
$ ls
$ logs sifchain
$ cat sifchain.logs | grep hash
```

### Client wrapper
```
$ whereis sifchain
$ nano /home/wojciech/.localnet/localnet/bin/sifchain
$ sifchain keys list
$ sifchain q bank balances <addr of master>
$ sifchain keys add test
$ sifchain tx bank send master <addr of test> 10rowan
$ sifchain q bank balances <addr of test>
$ exit
```

## Second instance of default set

Show how defaults are changed after entering environment

```
$ localnet start --help
don't run this, just show --> $ localnet start --env=localnet2 --network=127.2.0.0
$ localnet --env=localnet2 --network=127.2.0.0
$ start --help
$ start
$ exit
```

## Running another set
```
$ localnet --env=localnet3 --network=127.3.0.0 --set=full
```

Show source code of full set

### Spec before and after start
```
$ spec
$ start
$ spec
```

### Logs
```
$ ls
$ logs hermes
$ exit
```

## Integration tests

Run tests in tmux session

```
$ localnet --env=localnet4 --network=127.4.0.0
$ tests
$ spec
$ sifchain q tx <hash of transfer tx printed by test>
$ ls
$ logs sifchain
$ start <-- attach to the session
$ exit
```

### Filtering tests
```
$ localnet tests --env=localnet5 --network=127.5.0.0 --filter=Verify
```

Show how test names are related to test functions

Show how test set and test list are defined

Show implemented tests

### Start tests set separately
```
$ localnet start --env=localnet6 --network=127.6.0.0 --set=tests
$ localnet spec --env=localnet6
```

## Targets

### TMux
```
localnet start
```

### Direct
```
$ localnet --env=localnet7 --network=127.7.0.0 --target=direct
$ start
$ ps aux | grep sifnoded <-- last sifnoded instance with 127.7. IP
$ spec
$ logs sifchain
$ exit
```

### Docker

SWITCH TO ROOT!!!

```
$ su -
# /home/wojciech/sif/localnet/bin/localnet --env=localnet8 --set=full --target=docker --bin-dir=/home/wojciech/go/bin
# start
# spec
# podman ps
# ls
# logs sifchain-a
# exit
$ exit
```

## Environment variables
```
$ LOCALNET_ENV=localnet9 LOCALNET_NETWORK=127.9.0.0 localnet start
$ exit
```