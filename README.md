# Gozip
A simple and educational implementation of the zip format in go demonstrating the use of binary.Read.

```
$ go build
$ ./test/zip.sh
$ ./gozip test/test.zip
2021-11-27 12:39:52 +0100 CET: test/large.text
2021-11-27 12:39:52 +0100 CET: test/goodbye.text
2021-11-27 12:39:52 +0100 CET: test/hello.text
$ ./gozip -c test/test.zip
2021-11-27 12:39:52 +0100 CET: test/large.text
foo
foo
...
foo
foo
2021-11-27 12:39:52 +0100 CET: test/goodbye.text
Goodbye World!
2021-11-27 12:39:52 +0100 CET: test/hello.text
Hello World!
$
```

Bonus: You can also provide the zipfile on stdin instead of as an argument.

```
$ ./gozip < test/test.zip
2021-11-27 12:39:52 +0100 CET: test/large.text
2021-11-27 12:39:52 +0100 CET: test/goodbye.text
2021-11-27 12:39:52 +0100 CET: test/hello.text
$
```