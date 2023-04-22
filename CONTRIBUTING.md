# Contributing

use make static and make test for all the local checks. 
```bash
make static
make test
```

TO release a new verison run make release and specify the new version manually
```console
foo@bar:~$ make VERSION=0.0.1 release
foo@bar:~$ find release/
release/
release/0.0.1
release/0.0.1/ccloud-admin_0.0.1_linux_amd64.tar.gz
release/0.0.0
release/0.0.0/ccloud-admin_0.0.0_linux_amd64.tar.gz
release/ccloud-admin__linux_amd64.tar.gz

```
tag the commit with the manual version number, push the tag, upload the new release to github.